package tengo_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/shelepuginivan/tengo"
	"github.com/shelepuginivan/tengo/require"
	"github.com/shelepuginivan/tengo/stdlib"
	"github.com/shelepuginivan/tengo/token"
)

const module = `
add := func(a, b, ...c) {
	r := a + b
	for v in c {
		r += v
	}
	return r
}

mul := func(a, b, ...c) {
	r := a * b
	for v in c {
		r *= v
	}
	return r
}

square := func(a) {
	return mul(a, a)
}

fib := func(x) {
	if x == 0 {
		return 0
	} else if x == 1 {
		return 1
	}
	return fib(x-1) + fib(x-2)
}

stringer := func(s) {
	return string(s)
}
`

func TestScript_Add(t *testing.T) {
	s := tengo.NewScript([]byte(`a := b; c := test(b); d := test(5)`))
	require.NoError(t, s.Add("b", 5))     // b = 5
	require.NoError(t, s.Add("b", "foo")) // b = "foo"  (re-define before compilation)
	require.NoError(t, s.Add("test",
		func(args ...tengo.Object) (ret tengo.Object, err error) {
			if len(args) > 0 {
				switch arg := args[0].(type) {
				case *tengo.Int:
					return &tengo.Int{Value: arg.Value + 1}, nil
				}
			}

			return &tengo.Int{Value: 0}, nil
		}))
	c, err := s.CompileRun()
	require.NoError(t, err)
	require.Equal(t, "foo", c.Get("a").Value())
	require.Equal(t, "foo", c.Get("b").Value())
	require.Equal(t, int64(0), c.Get("c").Value())
	require.Equal(t, int64(6), c.Get("d").Value())
}

func TestScript_Remove(t *testing.T) {
	s := tengo.NewScript([]byte(`a := b`))
	err := s.Add("b", 5)
	require.NoError(t, err)
	require.True(t, s.Remove("b")) // b is removed
	_, err = s.CompileRun()        // should not compile because b is undefined
	require.Error(t, err)
}

func TestScript_CompileRun(t *testing.T) {
	s := tengo.NewScript([]byte(`a := b`))
	err := s.Add("b", 5)
	require.NoError(t, err)
	c, err := s.CompileRun()
	require.NoError(t, err)
	require.NotNil(t, c)
	compiledGet(t, c, "a", int64(5))
}

func TestScript_SourceModules(t *testing.T) {
	s := tengo.NewScript([]byte(`
enum := import("enum")
a := enum.all([1,2,3], func(_, v) { 
	return v > 0 
})
`))
	s.SetImports(stdlib.GetModuleMap("enum"))
	c, err := s.CompileRun()
	require.NoError(t, err)
	require.NotNil(t, c)
	compiledGet(t, c, "a", true)

	s.SetImports(nil)
	_, err = s.CompileRun()
	require.Error(t, err)
}

func TestScript_BuiltinModules(t *testing.T) {
	s := tengo.NewScript([]byte(`math := import("math"); a := math.abs(-19.84)`))
	s.SetImports(stdlib.GetModuleMap("math"))
	c, err := s.CompileRun()
	require.NoError(t, err)
	require.NotNil(t, c)
	compiledGet(t, c, "a", 19.84)

	c, err = s.CompileRun()
	require.NoError(t, err)
	require.NotNil(t, c)
	compiledGet(t, c, "a", 19.84)

	s.SetImports(stdlib.GetModuleMap("os"))
	_, err = s.CompileRun()
	require.Error(t, err)

	s.SetImports(nil)
	_, err = s.CompileRun()
	require.Error(t, err)
}

func TestCallByName(t *testing.T) {
	type testArgs struct {
		fn   string
		args []interface{}
		ret  interface{}
	}
	tests := []testArgs{
		{fn: "add", args: []interface{}{3, 4}, ret: int64(7)},
		{fn: "add", args: []interface{}{1, 2, 3, 4}, ret: int64(10)},
		{fn: "mul", args: []interface{}{3, 4}, ret: int64(12)},
		{fn: "mul", args: []interface{}{1, 2, 3, 4}, ret: int64(24)},
		{fn: "square", args: []interface{}{3}, ret: int64(9)},
		{fn: "fib", args: []interface{}{10}, ret: int64(55)},
		{fn: "stringer", args: []interface{}{12345}, ret: "12345"},
	}
	ctx := context.Background()
	script := tengo.NewScript([]byte(module))
	compl, err := script.CompileRun()
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		var comp *tengo.Compiled
		if i == 0 {
			// use same script for each test
			comp = compl
		} else if i == 1 {
			// create script for each test
			scr := tengo.NewScript([]byte(module))
			comp, err = scr.CompileRun()
			require.NoError(t, err)
		} else {
			// use clone
			comp = compl.Clone()
		}
		for _, test := range tests {
			result, err := comp.CallByName(test.fn, test.args...)
			require.NoError(t, err)
			require.Equal(t, test.ret, result)

			resultx, err := comp.CallByNameContext(ctx, test.fn, test.args...)
			require.NoError(t, err)
			require.Equal(t, test.ret, resultx)
		}
	}

}

func TestCallback(t *testing.T) {
	const callbackModule = `
b := 2

pass(func(a) {
	return a * b
})
`
	scr := tengo.NewScript([]byte(callbackModule))

	var callback *tengo.Callback
	scr.Add("pass", &tengo.UserFunction{
		Value: func(args ...tengo.Object) (tengo.Object, error) {
			callback = tengo.NewCallback(args[0])
			return tengo.UndefinedValue, nil
		},
	})

	compl, err := scr.CompileRun()
	require.NoError(t, err)
	require.NotNil(t, callback)
	// unset *Compiled throws error
	result, err := callback.Call(3)
	require.Error(t, err)

	// Set *Compiled before Call
	result, err = callback.Set(compl).Call(3)
	require.NoError(t, err)
	require.Equal(t, int64(6), result)

	result, err = callback.Set(compl).Call(5)
	require.NoError(t, err)
	require.Equal(t, int64(10), result)

	// Modify the global and check the new value is reflected in the function
	compl.Set("b", 3)
	result, err = callback.Set(compl).Call(5)
	require.NoError(t, err)
	require.Equal(t, int64(15), result)

	c := callback.Set(compl)
	resultx, err := c.CallContext(context.Background(), 5)
	require.Equal(t, result, resultx)
}

func TestClosure(t *testing.T) {
	const closureModule = `
mulClosure := func(a) {
	return func(b) {
		return a * b
	}
}

mul2 := mulClosure(2)
mul3 := mulClosure(3)
`

	scr := tengo.NewScript([]byte(closureModule))

	compl, err := scr.CompileRun()
	require.NoError(t, err)

	result, err := compl.CallByName("mul2", 3)
	require.NoError(t, err)
	require.Equal(t, int64(6), result)

	result, err = compl.CallByName("mul2", 5)
	require.NoError(t, err)
	require.Equal(t, int64(10), result)

	result, err = compl.CallByName("mul3", 5)
	require.NoError(t, err)
	require.Equal(t, int64(15), result)
}

func TestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	compl, err := tengo.NewScript([]byte("")).CompileRunContext(ctx)
	require.Error(t, err)
	require.Equal(t, context.Canceled.Error(), err.Error())

	ctx, cancel = context.WithTimeout(context.Background(), 0)
	defer cancel()
	scr := tengo.NewScript([]byte(module))
	compl, err = scr.CompileRunContext(context.Background())
	require.NoError(t, err)
	_, err = compl.CallByNameContext(ctx, "square", 2)
	require.Error(t, err)
	require.Equal(t, context.DeadlineExceeded.Error(), err.Error())
}

func TestImportCall(t *testing.T) {
	module := `contains := import("text").contains`
	scr := tengo.NewScript([]byte(module))
	mm := stdlib.GetModuleMap(stdlib.AllModuleNames()...)
	scr.SetImports(mm)
	compl, err := scr.CompileRun()
	require.NoError(t, err)
	v, err := compl.CallByName("contains", "foo bar", "bar")
	require.NoError(t, err)
	require.Equal(t, true, v)

	v, err = compl.CallByName("contains", "foo bar", "baz")
	require.NoError(t, err)
	require.Equal(t, false, v)

	v, err = compl.CallByName("containsX", "foo bar", "bar")
	require.True(t, strings.Contains(err.Error(), "not found"))
}

func TestCallable(t *testing.T) {
	// taken from tengo quick start example
	src := `
each := func(seq, fn) {
    for x in seq { fn(x) }
}

sum := 0
mul := 1

f := func(x) {
	sum += x
	mul *= x
}

each([a, b, c, d], f)`

	script := tengo.NewScript([]byte(src))

	// set values
	err := script.Add("a", 1)
	require.NoError(t, err)
	err = script.Add("b", 9)
	require.NoError(t, err)
	err = script.Add("c", 8)
	require.NoError(t, err)
	err = script.Add("d", 4)
	require.NoError(t, err)

	// compile and run the script
	compl, err := script.CompileRunContext(context.Background())
	require.NoError(t, err)

	// retrieve values
	sum := compl.Get("sum")
	mul := compl.Get("mul")
	require.Equal(t, 22, sum.Int())
	require.Equal(t, 288, mul.Int())

	_, err = compl.CallByName("f", 2)
	require.NoError(t, err)

	sum = compl.Get("sum")
	mul = compl.Get("mul")
	require.Equal(t, 22+2, sum.Int())
	require.Equal(t, 288*2, mul.Int())

	var args []tengo.Object
	eachf := &tengoCallable{
		callFunc: func(a ...tengo.Object) (tengo.Object, error) {
			if len(a) != 1 {
				panic(fmt.Errorf("1 argument is expected but got %d",
					len(a)))
			}
			args = append(args, a[0])
			return tengo.UndefinedValue, nil
		},
	}
	nums := []interface{}{1, 2, 3, 4}
	_, err = compl.CallByName("each", nums, eachf)
	require.NoError(t, err)
	require.Equal(t, 4, len(args))
	for i, v := range args {
		vv := tengo.ToInterface(v)
		require.Equal(t, int64(nums[i].(int)), vv.(int64))
	}
}

func TestGetAll(t *testing.T) {
	script := tengo.NewScript([]byte(module))
	compl, err := script.CompileRunContext(context.Background())
	require.NoError(t, err)
	vars := compl.GetAll()
	varsMap := make(map[string]bool)
	for _, v := range vars {
		varsMap[v.Name()] = true
	}
	names := []string{"add", "mul", "square", "fib", "stringer"}
	for _, v := range names {
		require.True(t, compl.IsDefined(v))
		require.True(t, varsMap[v])
	}
}

type tengoCallable struct {
	tengo.ObjectImpl
	callFunc tengo.CallableFunc
}

func (tc *tengoCallable) CanCall() bool {
	return true
}

func (tc *tengoCallable) Call(args ...tengo.Object) (tengo.Object, error) {
	return tc.callFunc(args...)
}

type Counter struct {
	tengo.ObjectImpl
	value int64
}

func (o *Counter) TypeName() string {
	return "counter"
}

func (o *Counter) String() string {
	return fmt.Sprintf("Counter(%d)", o.value)
}

func (o *Counter) BinaryOp(
	op token.Token,
	rhs tengo.Object,
) (tengo.Object, error) {
	switch rhs := rhs.(type) {
	case *Counter:
		switch op {
		case token.Add:
			return &Counter{value: o.value + rhs.value}, nil
		case token.Sub:
			return &Counter{value: o.value - rhs.value}, nil
		}
	case *tengo.Int:
		switch op {
		case token.Add:
			return &Counter{value: o.value + rhs.Value}, nil
		case token.Sub:
			return &Counter{value: o.value - rhs.Value}, nil
		}
	}

	return nil, errors.New("invalid operator")
}

func (o *Counter) IsFalsy() bool {
	return o.value == 0
}

func (o *Counter) Equals(t tengo.Object) bool {
	if tc, ok := t.(*Counter); ok {
		return o.value == tc.value
	}

	return false
}

func (o *Counter) Copy() tengo.Object {
	return &Counter{value: o.value}
}

func (o *Counter) Call(_ ...tengo.Object) (tengo.Object, error) {
	return &tengo.Int{Value: o.value}, nil
}

func (o *Counter) CanCall() bool {
	return true
}

func TestScript_CustomObjects(t *testing.T) {
	c := scriptCompileRun(t, `a := c1(); s := string(c1); c2 := c1; c2++`, M{
		"c1": &Counter{value: 5},
	})
	compiledGet(t, c, "a", int64(5))
	compiledGet(t, c, "s", "Counter(5)")
	compiledGetCounter(t, c, "c2", &Counter{value: 6})

	c = scriptCompileRun(t, `
arr := [1, 2, 3, 4]
for x in arr {
	c1 += x
}
out := c1()
`, M{
		"c1": &Counter{value: 5},
	})
	compiledGet(t, c, "out", int64(15))
}

func compiledGetCounter(
	t *testing.T,
	c *tengo.Compiled,
	name string,
	expected *Counter,
) {
	v := c.Get(name)
	require.NotNil(t, v)

	actual := v.Value().(*Counter)
	require.NotNil(t, actual)
	require.Equal(t, expected.value, actual.value)
}

func TestScriptSourceModule(t *testing.T) {
	// script1 imports "mod1"
	scr := tengo.NewScript([]byte(`out := import("mod")`))
	mods := tengo.NewModuleMap()
	mods.AddSourceModule("mod", []byte(`export 5`))
	scr.SetImports(mods)
	c, err := scr.CompileRun()
	require.NoError(t, err)
	require.Equal(t, int64(5), c.Get("out").Value())

	// executing module function
	scr = tengo.NewScript([]byte(`fn := import("mod"); out := fn()`))
	mods = tengo.NewModuleMap()
	mods.AddSourceModule("mod",
		[]byte(`a := 3; export func() { return a + 5 }`))
	scr.SetImports(mods)
	c, err = scr.CompileRun()
	require.NoError(t, err)
	require.Equal(t, int64(8), c.Get("out").Value())

	scr = tengo.NewScript([]byte(`out := import("mod")`))
	mods = tengo.NewModuleMap()
	mods.AddSourceModule("mod",
		[]byte(`text := import("text"); export text.title("foo")`))
	mods.AddBuiltinModule("text",
		map[string]tengo.Object{
			"title": &tengo.UserFunction{
				Name: "title",
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					s, _ := tengo.ToString(args[0])
					return &tengo.String{Value: strings.Title(s)}, nil
				}},
		})
	scr.SetImports(mods)
	c, err = scr.CompileRun()
	require.NoError(t, err)
	require.Equal(t, "Foo", c.Get("out").Value())
	scr.SetImports(nil)
	_, err = scr.CompileRun()
	require.Error(t, err)
}

type M map[string]interface{}

func TestCompiled_Get(t *testing.T) {
	c := scriptCompileRun(t, `a := 5`, nil)
	compiledGet(t, c, "a", int64(5))

	// user-defined variables
	c = scriptCompileRun(t, `a := b`, M{"b": "foo"})
	compiledGet(t, c, "a", "foo")
	compileError(t, `a := b`, nil)
}

func TestCompiled_GetAll(t *testing.T) {
	c := scriptCompileRun(t, `a := 5`, nil)
	compiledGetAll(t, c, M{"a": int64(5)})

	c = scriptCompileRun(t, `a := b`, M{"b": "foo"})
	compiledGetAll(t, c, M{"a": "foo", "b": "foo"})

	c = scriptCompileRun(t, `a := b; b = 5`, M{"b": "foo"})
	compiledGetAll(t, c, M{"a": "foo", "b": int64(5)})
}

func TestCompiled_Set(t *testing.T) {
	c := scriptCompileRun(t, `a := 5`, nil)
	compiledGet(t, c, "a", int64(5))
	c.Set("a", 6)
	compiledGet(t, c, "a", int64(6))
}

func TestCompiled_CustomObject(t *testing.T) {
	c := scriptCompileRun(t, `r := (t<130)`, M{"t": &customNumber{value: 123}})
	compiledGet(t, c, "r", true)

	c = scriptCompileRun(t, `r := (t>13)`, M{"t": &customNumber{value: 123}})
	compiledGet(t, c, "r", true)
}

// customNumber is a user defined object that can compare to tengo.Int
// very shitty implementation, just to test that token.Less and token.Greater in BinaryOp works
type customNumber struct {
	tengo.ObjectImpl
	value int64
}

func (n *customNumber) TypeName() string {
	return "Number"
}

func (n *customNumber) String() string {
	return strconv.FormatInt(n.value, 10)
}

func (n *customNumber) BinaryOp(op token.Token, rhs tengo.Object) (tengo.Object, error) {
	tengoInt, ok := rhs.(*tengo.Int)
	if !ok {
		return nil, tengo.ErrInvalidOperator
	}
	return n.binaryOpInt(op, tengoInt)
}

func (n *customNumber) binaryOpInt(op token.Token, rhs *tengo.Int) (tengo.Object, error) {
	i := n.value

	switch op {
	case token.Less:
		if i < rhs.Value {
			return tengo.TrueValue, nil
		}
		return tengo.FalseValue, nil
	case token.Greater:
		if i > rhs.Value {
			return tengo.TrueValue, nil
		}
		return tengo.FalseValue, nil
	case token.LessEq:
		if i <= rhs.Value {
			return tengo.TrueValue, nil
		}
		return tengo.FalseValue, nil
	case token.GreaterEq:
		if i >= rhs.Value {
			return tengo.TrueValue, nil
		}
		return tengo.FalseValue, nil
	}
	return nil, tengo.ErrInvalidOperator
}

func TestScript_ImportError(t *testing.T) {
	m := `
	exp := import("expression")
	r := exp(ctx)
`

	src := `
export func(ctx) {
	closure := func() {
		if ctx.actiontimes < 0 { // an error is thrown here because actiontimes is undefined
			return true
		}
		return false
	}

	return closure()
}`

	s := tengo.NewScript([]byte(m))
	mods := tengo.NewModuleMap()
	mods.AddSourceModule("expression", []byte(src))
	s.SetImports(mods)

	err := s.Add("ctx", map[string]interface{}{
		"ctx": 12,
	})
	require.NoError(t, err)

	_, err = s.CompileRun()
	require.True(t, strings.Contains(err.Error(), "expression:4:6"))
}

func TestCompiled_Clone(t *testing.T) {
	script := tengo.NewScript([]byte(`
count += 1
data["b"] = 2
`))

	err := script.Add("data", map[string]interface{}{"a": 1})
	require.NoError(t, err)

	err = script.Add("count", 1000)
	require.NoError(t, err)

	compiled, err := script.CompileRun()
	require.NoError(t, err)

	require.Equal(t, 1001, compiled.Get("count").Int())
	require.Equal(t, 2, len(compiled.Get("data").Map()))
}

func compileError(t *testing.T, input string, vars M) {
	s := tengo.NewScript([]byte(input))
	for vn, vv := range vars {
		err := s.Add(vn, vv)
		require.NoError(t, err)
	}
	_, err := s.CompileRun()
	require.Error(t, err)
}

func scriptCompileRun(t *testing.T, src string, vars M) *tengo.Compiled {
	s := tengo.NewScript([]byte(src))
	for vn, vv := range vars {
		err := s.Add(vn, vv)
		require.NoError(t, err)
	}
	c, err := s.CompileRun()
	require.NoError(t, err)

	return c
}

func compiledGet(
	t *testing.T,
	c *tengo.Compiled,
	name string,
	expected interface{},
) {
	v := c.Get(name)
	require.NotNil(t, v)
	require.Equal(t, expected, v.Value())
}

func compiledGetAll(
	t *testing.T,
	c *tengo.Compiled,
	expected M,
) {
	vars := c.GetAll()
	require.Equal(t, len(expected), len(vars)-1) // One variable is reserved.

	for k, v := range expected {
		var found bool
		for _, e := range vars {
			if e.Name() == k {
				require.Equal(t, v, e.Value())
				found = true
			}
		}
		require.True(t, found, "variable '%s' not found", k)
	}
}

func compiledIsDefined(
	t *testing.T,
	c *tengo.Compiled,
	name string,
	expected bool,
) {
	require.Equal(t, expected, c.IsDefined(name))
}
