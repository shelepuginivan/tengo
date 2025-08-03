package tengo

// A modified version of Tengo's script compilation and run implementation.
// Based on https://github.com/ozanh/tengox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/d5/tengo/v2/parser"
)

const reservedVar = "$out"

// Script can simplify compilation and execution of embedded scripts.
type Script struct {
	trace     io.Writer
	modules   ModuleGetter
	variables map[string]*Variable
	src       []byte
	maxAllocs int64
	importDir string
}

// NewScript creates a Script object.
func NewScript(src []byte) *Script {
	return &Script{
		src:       src,
		variables: make(map[string]*Variable),
		maxAllocs: -1,
	}
}

// CompileRun compiles the source script to bytecode and run it once to fill
// global objects.
func (s *Script) CompileRun() (*Compiled, error) {
	return s.compile(context.Background())
}

// CompileRunContext compiles the source script to bytecode and run it once to
// fill global objects.
func (s *Script) CompileRunContext(ctx context.Context) (*Compiled, error) {
	return s.compile(ctx)
}

// SetImportDir sets the initial import directory for script files.
func (s *Script) SetImportDir(dir string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	s.importDir = dir
	return nil
}

func (s *Script) compile(ctx context.Context) (*Compiled, error) {
	symbolTable, globals, err := s.prepCompile()
	if err != nil {
		return nil, err
	}

	fileSet := parser.NewFileSet()
	srcFile := fileSet.AddFile("(main)", -1, len(s.src))
	p := parser.NewParser(srcFile, s.src, nil)
	file, err := p.ParseFile()
	if err != nil {
		return nil, err
	}

	out := symbolTable.Define(reservedVar)
	globals[out.Index] = UndefinedValue

	cc := NewCompiler(srcFile, symbolTable, nil, s.modules, s.trace)
	cc.EnableFileImport(true)
	cc.SetImportDir(cc.importDir)
	if err := cc.Compile(file); err != nil {
		return nil, err
	}

	globals = globals[:symbolTable.MaxSymbols()+1]

	// global symbol names to indexes
	globalIndexes := make(map[string]int, len(globals))
	for _, name := range symbolTable.Names() {
		symbol, _, _ := symbolTable.Resolve(name, false)
		if symbol.Scope == ScopeGlobal {
			globalIndexes[name] = symbol.Index
		}
	}

	// run VM once and fill globals
	if ctx == nil {
		vm := NewVM(cc.Bytecode(), globals, s.maxAllocs)
		if err := vm.Run(); err != nil {
			return nil, err
		}
	} else {
		vm := NewVM(cc.Bytecode(), globals, s.maxAllocs)
		if err := runVMContext(ctx, vm); err != nil {
			return nil, err
		}
	}

	bc := cc.Bytecode()
	bc.RemoveDuplicates()
	return &Compiled{
		bytecode:  bc,
		globals:   globals,
		indexes:   globalIndexes,
		maxAllocs: s.maxAllocs,
		outIdx:    out.Index,
	}, nil
}

func (s *Script) prepCompile() (
	symbolTable *SymbolTable,
	globals []Object,
	err error,
) {
	var names []string
	for name := range s.variables {
		names = append(names, name)
	}

	symbolTable = NewSymbolTable()

	globals = make([]Object, GlobalsSize)

	for idx, name := range names {
		symbol := symbolTable.Define(name)
		if symbol.Index != idx {
			panic(fmt.Errorf("wrong symbol index: %d != %d",
				idx, symbol.Index))
		}
		globals[symbol.Index] = s.variables[name].Object()
	}
	return
}

// Add adds a new variable or updates an existing variable to the script.
func (s *Script) Add(name string, value interface{}) error {
	if name == reservedVar {
		return errors.New("variable name must be different")
	}

	v, err := NewVariable(name, value)
	if err != nil {
		return err
	}
	s.variables[name] = v
	return nil
}

// Remove removes (undefines) an existing variable for the script. It returns
// false if the variable name is not defined.
func (s *Script) Remove(name string) bool {
	if name == reservedVar {
		return false
	}

	if _, ok := s.variables[name]; !ok {
		return false
	}
	delete(s.variables, name)
	return true
}

// SetImports sets import modules.
func (s *Script) SetImports(modules ModuleGetter) {
	s.modules = modules
}

// SetMaxAllocs sets the maximum number of objects allocations during the run
// time. Compiled script will return ErrObjectAllocLimit error if it
// exceeds this limit.
func (s *Script) SetMaxAllocs(n int64) {
	s.maxAllocs = n
}

// Trace set a tracer for compiler and VM for debugging purposes.
func (s *Script) Trace(w io.Writer) {
	s.trace = w
}

// Compiled is a compiled instance of the user script. Use Script.CompileRun()
// to create Compiled object.
type Compiled struct {
	mu        sync.RWMutex
	bytecode  *Bytecode
	indexes   map[string]int
	globals   []Object
	maxAllocs int64
	outIdx    int
}

// Clone creates a new copy of Compiled. Cloned copies are safe for concurrent
// use. Clones occupy less memory..
func (c *Compiled) Clone() *Compiled {
	c.mu.RLock()
	defer c.mu.RUnlock()

	clone := &Compiled{
		indexes:   c.indexes,
		bytecode:  c.bytecode,
		globals:   make([]Object, len(c.globals)),
		maxAllocs: c.maxAllocs,
		outIdx:    c.outIdx,
	}
	// copy global objects
	for idx, g := range c.globals {
		if g != nil {
			clone.globals[idx] = g.Copy()
		}
	}
	return clone
}

// CallByName calls callable Object by its name and with given
// arguments, and returns result.
// args must be convertible to supported Tengo types.
func (c *Compiled) CallByName(fn string, args ...interface{}) (interface{}, error) {
	return c.CallByNameContext(context.Background(), fn, args...)
}

// CallByNameContext calls callable Object by its name and with given
// arguments, and returns result.
// args must be convertible to supported Tengo types.
func (c *Compiled) CallByNameContext(ctx context.Context,
	fn string, args ...interface{}) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	idx, ok := c.indexes[fn]
	if !ok {
		return nil, fmt.Errorf("'%s' not found", fn)
	}
	cfn := c.globals[idx]
	if cfn == nil {
		return nil, errors.New("callable expected, got nil")
	}
	if !cfn.CanCall() {
		return nil, errors.New("not a callable")
	}

	return c.call(ctx, cfn, args...)
}

// Call calls callable Object with given arguments, and returns result.
// args must be convertible to supported Tengo types.
func (c *Compiled) Call(fn Object, args ...interface{}) (interface{}, error) {
	return c.CallContext(context.Background(), fn, args...)
}

// CallContext calls callable Object with given arguments, and returns result.
// args must be convertible to supported Tengo types.
func (c *Compiled) CallContext(ctx context.Context, fn Object,
	args ...interface{}) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if fn == nil {
		return nil, errors.New("callable expected, got nil")
	}
	if !fn.CanCall() {
		return nil, errors.New("not a callable")
	}

	return c.call(ctx, fn, args...)
}

func (c *Compiled) call(ctx context.Context, cfn Object,
	args ...interface{}) (interface{}, error) {
	targs := make([]Object, 0, len(args))
	for i := range args {
		v, err := FromInterface(args[i])
		if err != nil {
			return nil, err
		}
		targs = append(targs, v)
	}

	v, err := c.callCompiled(ctx, cfn, targs...)
	if err != nil {
		return nil, err
	}
	return ToInterface(v), nil
}

func (c *Compiled) callCompiled(ctx context.Context, fn Object,
	args ...Object) (Object, error) {

	constsOffset := len(c.bytecode.Constants)

	// Load fn
	inst := MakeInstruction(parser.OpConstant, constsOffset)

	// Load args
	for i := range args {
		inst = append(inst,
			MakeInstruction(parser.OpConstant, constsOffset+i+1)...)
	}

	// Call, set value to a global, stop
	inst = append(inst, MakeInstruction(parser.OpCall, len(args))...)
	inst = append(inst, MakeInstruction(parser.OpSetGlobal, c.outIdx)...)
	inst = append(inst, MakeInstruction(parser.OpSuspend)...)

	c.bytecode.Constants = append(c.bytecode.Constants, fn)
	c.bytecode.Constants = append(c.bytecode.Constants, args...)

	// orig := s.bytecode.MainFunction
	c.bytecode.MainFunction = &CompiledFunction{
		Instructions: inst,
	}

	var err error
	if ctx == nil {
		vm := NewVM(c.bytecode, c.globals, c.maxAllocs)
		err = vm.Run()
	} else {
		vm := NewVM(c.bytecode, c.globals, c.maxAllocs)
		err = runVMContext(ctx, vm)
	}

	// TODO: go back to normal if required
	// s.bytecode.MainFunction = orig
	// avoid memory leak.
	for i := constsOffset; i < len(c.bytecode.Constants); i++ {
		c.bytecode.Constants[i] = nil
	}
	c.bytecode.Constants = c.bytecode.Constants[:constsOffset]

	// get symbol using index and return it
	return c.globals[c.outIdx], err
}

// Get returns a variable identified by the name.
func (c *Compiled) Get(name string) *Variable {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value := UndefinedValue
	if idx, ok := c.indexes[name]; ok {
		value = c.globals[idx]
		if value == nil {
			value = UndefinedValue
		}
	}

	v, err := NewVariable(name, value)
	if err != nil {
		// This should never happen, since the value will already be a tengo object
		// tengo's (*Script).Get avoids this because it constructs the object directly.
		panic(fmt.Errorf("unable to create new variable from global value, type was %T", value))
	}

	return v
}

// GetAll returns all the variables that are defined by the compiled script.
func (c *Compiled) GetAll() []*Variable {
	vars := make([]*Variable, 0, len(c.indexes))
	for name, idx := range c.indexes {
		value := c.globals[idx]
		if value == nil {
			value = UndefinedValue
		}
		tv, err := NewVariable(name, value)
		if err != nil {
			// This should never happen, since the value will already be a tengo object
			panic(fmt.Errorf("unable to create new variable from global value, type was %T", value))
		}
		vars = append(vars, tv)
	}
	return vars
}

// Set replaces the value of a global variable identified by the name. An error
// will be returned if the name was not defined during compilation.
func (c *Compiled) Set(name string, value interface{}) error {
	obj, err := FromInterface(value)
	if err != nil {
		return err
	}

	idx, ok := c.indexes[name]
	if !ok {
		return fmt.Errorf("'%s' is not defined", name)
	}

	c.globals[idx] = obj
	return nil
}

// IsDefined returns true if the variable name is defined (has value)
// after the execution.
func (c *Compiled) IsDefined(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	idx, ok := c.indexes[name]
	if !ok {
		return false
	}
	v := c.globals[idx]
	if v == nil {
		return false
	}
	return v != UndefinedValue
}

func runVMContext(ctx context.Context, vm *VM) (err error) {
	errch := make(chan error)
	go func() {
		errch <- vm.Run()
	}()
	select {
	case err = <-errch:
	case <-ctx.Done():
		vm.Abort()
		<-errch
		err = ctx.Err()
	}
	if err != nil {
		return
	}
	return
}

// Callback is a wrapper to call a callable Object from Go with Call() and
// CallContext() methods. *Compiled object must be set before Call() and
// CallContext() calls, otherwise an error is thrown.
// Callback is intended to be created in a Go function that is invoked by tengo
// script to capture callable Object and additional arguments if required
// later. Args is deliberately exposed to use it as arguments to CallXXX methods
// but it is optional.
// Note: Do not call CallXXX methods while script is running, it locks the VM.
type Callback struct {
	Args     []interface{}
	compiled *Compiled
	fn       Object
}

// NewCallback creates Callback object. See Callback type.
func NewCallback(fn Object, args ...interface{}) *Callback {
	return &Callback{
		fn:   fn,
		Args: args,
	}
}

// Set sets compiled object, which is required to call callables.
func (cb *Callback) Set(c *Compiled) *Callback {
	cb.compiled = c
	return cb
}

// Call calls Object and returns result. Set *Compiled object before
// Call().
func (cb *Callback) Call(args ...interface{}) (interface{}, error) {
	if cb.compiled == nil {
		return nil, errors.New("compiled code not set")
	}
	return cb.compiled.Call(cb.fn, args...)
}

// CallContext calls Object and returns result. Set *Compiled object
// before CallContext().
func (cb *Callback) CallContext(ctx context.Context,
	args ...interface{}) (interface{}, error) {
	if cb.compiled == nil {
		return nil, errors.New("compiled code not set")
	}
	return cb.compiled.CallContext(ctx, cb.fn, args...)
}
