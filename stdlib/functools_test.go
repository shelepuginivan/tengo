package stdlib_test

import (
	"testing"

	"github.com/shelepuginivan/tengo"
	"github.com/shelepuginivan/tengo/require"
)

func TestFilter(t *testing.T) {
	compiled, err := runWith(`
r := import("functools").filter([1, 2, 3, 4, 5, 6], func(x) {
	return x % 2 == 0
})
`, "functools")

	require.NoError(t, err)
	require.Equal(t, []any{int64(2), int64(4), int64(6)}, compiled.Get("r").Value())

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("functools").filter()`,
			`import("functools").filter(0, func(x) { return x })`,
			`import("functools").filter("ok", func() { return 0 })`,
			`import("functools").filter([1, 2, 3], undefined)`,
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "functools")
				require.Error(t, err)
			})
		}
	})
}

func TestForeach(t *testing.T) {
	compiled, err := runWith(`
foreach := import("functools").foreach

foreach_1 := 0

foreach([1, 2, 3, 4, 5], func(x) {
	foreach_1 += x*x
})

foreach_2 := ""

foreach(["who", "is", "reading", "this", "?"], func(x) {
	foreach_2 += x
})
`, "functools")

	require.NoError(t, err)

	var (
		v  any
		ok bool
	)

	v, ok = tengo.ToInt(compiled.Get("foreach_1").Object())
	require.True(t, ok)
	require.Equal(t, 55, v)

	v, ok = tengo.ToString(compiled.Get("foreach_2").Object())
	require.True(t, ok)
	require.Equal(t, "whoisreadingthis?", v)

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("functools").foreach()`,
			`import("functools").foreach(nil, func(x) {})`,
			`import("functools").foreach([1, 2, 3], nil)`,
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "functools")
				require.Error(t, err)
			})
		}
	})
}

func TestMap(t *testing.T) {
	compiled, err := runWith(`
map := import("functools").map

map_1 := map([1, 2, 3, 4, 5], func(x) {
	return x * x * x
})

map_2 := map([1, 2, 3, 4, 5], string)

is_even := func(n) {
	return n % 2 == 0
}

map_3 := map([2, 4, 42, 69, 1337, 29374209475], is_even)
`, "functools")

	require.NoError(t, err)

	var (
		v   *tengo.Variable
		arr []any
	)

	v = compiled.Get("map_1")
	arr = v.Array()
	require.Equal(t, []any{int64(1), int64(8), int64(27), int64(64), int64(125)}, arr)

	v = compiled.Get("map_2")
	arr = v.Array()
	require.Equal(t, []any{"1", "2", "3", "4", "5"}, arr)

	v = compiled.Get("map_3")
	arr = v.Array()
	require.Equal(t, []any{true, true, true, false, false, false}, arr)

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("functools").map()`,
			`import("functools").map(nil, func(x) {})`,
			`import("functools").map([1, 2, 3], nil)`,
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "functools")
				require.Error(t, err)
			})
		}
	})
}

func TestPartial(t *testing.T) {
	compiled, err := runWith(`
functools := import("functools")

sum := func(a, b, c, d, e) {
	return a + b + c + d + e
}

p := functools.partial(sum, 1, 2)

v1 := p(3, 4, 5)
v2 := p(0, -1, -2)
v3 := p(1, 50, 9)
`, "functools")

	require.NoError(t, err)
	require.Equal(t, int64(15), compiled.Get("v1").Value())
	require.Equal(t, int64(0), compiled.Get("v2").Value())
	require.Equal(t, int64(63), compiled.Get("v3").Value())
}

func TestReduce(t *testing.T) {
	compiled, err := runWith(`
reduce := import("functools").reduce

reduce_1 := reduce([1, 2, 3, 4, 5], 0, func(acc, cur) {
	return acc + 2 * cur
})

reduce_2 := reduce(['s', 'a', 'y', 'a'], "", func(acc, cur) {
	return acc + cur
})
`, "functools")

	require.NoError(t, err)

	var (
		v  any
		ok bool
	)

	v, ok = tengo.ToInt(compiled.Get("reduce_1").Object())
	require.True(t, ok)
	require.Equal(t, 30, v)

	v, ok = tengo.ToString(compiled.Get("reduce_2").Object())
	require.True(t, ok)
	require.Equal(t, "saya", v)

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("functools").reduce()`,
			`import("functools").reduce(nil, 0, func(x) {})`,
			`import("functools").reduce([1, 2, 3], 0, nil)`,
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "functools")
				require.Error(t, err)
			})
		}
	})
}
