package stdlib

import (
	"github.com/shelepuginivan/tengo"
)

// BuiltinModules are builtin type standard library modules.
var BuiltinModules = map[string]map[string]tengo.Object{
	"datetime": datetimeModule,
	"math":     mathModule,
	"os":       osModule,
	"text":     textModule,
	"rand":     randModule,
	"fmt":      fmtModule,
	"json":     jsonModule,
	"base64":   base64Module,
	"hex":      hexModule,
}
