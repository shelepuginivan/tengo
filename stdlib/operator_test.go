package stdlib_test

import (
	"testing"

	"github.com/shelepuginivan/tengo/require"
)

func TestAdd(t *testing.T) {
	compiled, err := runWith(`
add := import("operator").add

v1 := add(1, 3)
v2 := add(2.5, 2.5)
v3 := add(2, 3.5)
v4 := add("a", 'b')
v5 := add("c", "d")
`, "operator")

	require.NoError(t, err)
	require.Equal(t, int64(4), compiled.Get("v1").Value())
	require.InDelta(t, 5.0, compiled.Get("v2").Value(), 0.001)
	require.InDelta(t, 5.5, compiled.Get("v3").Value(), 0.001)
	require.Equal(t, "ab", compiled.Get("v4").Value())
	require.Equal(t, "cd", compiled.Get("v5").Value())

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("operator").add()`,
			`import("operator").add(1, "s")`,
			`import("operator").add([], {})`,
			`import("operator").add(undefined, 3)`,
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "operator")
				require.Error(t, err)
			})
		}
	})
}

func TestSub(t *testing.T) {
	compiled, err := runWith(`
sub := import("operator").sub

v1 := sub(69, 42)
v2 := sub(10.5, 5.5)
v3 := sub(1, 1.5)
`, "operator")

	require.NoError(t, err)
	require.Equal(t, int64(27), compiled.Get("v1").Value())
	require.InDelta(t, 5.0, compiled.Get("v2").Value(), 0.001)
	require.InDelta(t, -0.5, compiled.Get("v3").Value(), 0.001)

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("operator").sub()`,
			`import("operator").sub("a", 'b')`,
			`import("operator").sub(2.3, "10q")`,
			`import("operator").sub([], {})`,
			`import("operator").sub(undefined, 3)`,
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "operator")
				require.Error(t, err)
			})
		}
	})
}

func TestMul(t *testing.T) {
	compiled, err := runWith(`
mul := import("operator").mul

v1 := mul(2, 6)
v2 := mul(1.5, 1.5)
v3 := mul(4, 2.5)
`, "operator")

	require.NoError(t, err)
	require.Equal(t, int64(12), compiled.Get("v1").Value())
	require.InDelta(t, 2.25, compiled.Get("v2").Value(), 0.001)
	require.InDelta(t, 10.0, compiled.Get("v3").Value(), 0.001)

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("operator").mul()`,
			`import("operator").mul("aaa", 'b')`,
			`import("operator").mul(10.3, "some")`,
			`import("operator").mul({}, [1, 2])`,
			`import("operator").mul(8, undefined)`,
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "operator")
				require.Error(t, err)
			})
		}
	})
}

func TestDiv(t *testing.T) {
	compiled, err := runWith(`
div := import("operator").div

v1 := div(10, 2)
v2 := div(7, 2)
v3 := div(5.5, 2)
v4 := div(9, 3.0)
`, "operator")

	require.NoError(t, err)
	require.InDelta(t, 5.0, compiled.Get("v1").Value(), 0.001)
	require.InDelta(t, 3.5, compiled.Get("v2").Value(), 0.001)
	require.InDelta(t, 2.75, compiled.Get("v3").Value(), 0.001)
	require.InDelta(t, 3.0, compiled.Get("v4").Value(), 0.001)

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("operator").div()`,
			`import("operator").div("a", 2)`,
			`import("operator").div([], {})`,
			`import("operator").div(undefined, 3)`,
		}
		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "operator")
				require.Error(t, err)
			})
		}
	})
}

func TestFloordiv(t *testing.T) {
	compiled, err := runWith(`
floordiv := import("operator").floordiv

v1 := floordiv(10, 3)
v2 := floordiv(7, 2)
v3 := floordiv(9, 3)
`, "operator")

	require.NoError(t, err)
	require.Equal(t, int64(3), compiled.Get("v1").Value())
	require.Equal(t, int64(3), compiled.Get("v2").Value())
	require.Equal(t, int64(3), compiled.Get("v3").Value())

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("operator").floordiv()`,
			`import("operator").floordiv("a", 2)`,
			`import("operator").floordiv([], {})`,
			`import("operator").floordiv(undefined, 3)`,
		}
		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "operator")
				require.Error(t, err)
			})
		}
	})
}

func TestNeg(t *testing.T) {
	compiled, err := runWith(`
neg := import("operator").neg

v1 := neg(5)
v2 := neg(-3)
v3 := neg(2.5)
`, "operator")

	require.NoError(t, err)
	require.Equal(t, int64(-5), compiled.Get("v1").Value())
	require.Equal(t, int64(3), compiled.Get("v2").Value())
	require.InDelta(t, -2.5, compiled.Get("v3").Value(), 0.001)

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("operator").neg()`,
			`import("operator").neg("a")`,
			`import("operator").neg([])`,
			`import("operator").neg(undefined)`,
		}
		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "operator")
				require.Error(t, err)
			})
		}
	})
}

func TestEq(t *testing.T) {
	compiled, err := runWith(`
eq := import("operator").eq

v1 := eq(1, 1)
v2 := eq(1, 2)
v3 := eq("a", "a")
v4 := eq("a", "b")
`, "operator")

	require.NoError(t, err)
	require.Equal(t, true, compiled.Get("v1").Value())
	require.Equal(t, false, compiled.Get("v2").Value())
	require.Equal(t, true, compiled.Get("v3").Value())
	require.Equal(t, false, compiled.Get("v4").Value())

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("operator").eq()`,
			`import("operator").eq(undefined)`,
		}
		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "operator")
				require.Error(t, err)
			})
		}
	})
}

func TestNe(t *testing.T) {
	compiled, err := runWith(`
ne := import("operator").ne

v1 := ne(1, 1)
v2 := ne(1, 2)
v3 := ne("a", "a")
v4 := ne("a", "b")
`, "operator")

	require.NoError(t, err)
	require.Equal(t, false, compiled.Get("v1").Value())
	require.Equal(t, true, compiled.Get("v2").Value())
	require.Equal(t, false, compiled.Get("v3").Value())
	require.Equal(t, true, compiled.Get("v4").Value())

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("operator").ne()`,
			`import("operator").ne(1)`,
		}
		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "operator")
				require.Error(t, err)
			})
		}
	})
}

func TestGtGeLtLe(t *testing.T) {
	compiled, err := runWith(`
gt := import("operator").gt
ge := import("operator").ge
lt := import("operator").lt
le := import("operator").le

v1 := gt(3, 2)
v2 := ge(3, 3)
v3 := lt(2, 3)
v4 := le(3, 3)
v5 := gt("b", "a")
v6 := lt("a", "b")
`, "operator")

	require.NoError(t, err)
	require.Equal(t, true, compiled.Get("v1").Value())
	require.Equal(t, true, compiled.Get("v2").Value())
	require.Equal(t, true, compiled.Get("v3").Value())
	require.Equal(t, true, compiled.Get("v4").Value())
	require.Equal(t, true, compiled.Get("v5").Value())
	require.Equal(t, true, compiled.Get("v6").Value())

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("operator").gt()`,
			`import("operator").ge("a")`,
			`import("operator").lt([], {})`,
			`import("operator").le(undefined, 3)`,
		}
		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "operator")
				require.Error(t, err)
			})
		}
	})
}

func TestNot(t *testing.T) {
	compiled, err := runWith(`
not := import("operator").not

v1 := not(true)
v2 := not(false)
v3 := not(0)
v4 := not(1)
`, "operator")

	require.NoError(t, err)
	require.Equal(t, false, compiled.Get("v1").Value())
	require.Equal(t, true, compiled.Get("v2").Value())
	require.Equal(t, true, compiled.Get("v3").Value())
	require.Equal(t, false, compiled.Get("v4").Value())

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("operator").not()`,
		}
		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "operator")
				require.Error(t, err)
			})
		}
	})
}

func TestAndOrXor(t *testing.T) {
	compiled, err := runWith(`
and := import("operator").and
or := import("operator").or
xor := import("operator").xor

v1 := and(6, 3)
v2 := or(6, 3)
v3 := xor(6, 3)
`, "operator")

	require.NoError(t, err)
	require.Equal(t, int64(2), compiled.Get("v1").Value())
	require.Equal(t, int64(7), compiled.Get("v2").Value())
	require.Equal(t, int64(5), compiled.Get("v3").Value())

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("operator").and()`,
			`import("operator").or("a", 2)`,
			`import("operator").xor([], {})`,
			`import("operator").and(undefined, 3)`,
		}
		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "operator")
				require.Error(t, err)
			})
		}
	})
}

func TestInv(t *testing.T) {
	compiled, err := runWith(`
inv := import("operator").inv

v1 := inv(2)
v2 := inv(0)
v3 := inv(-1)
`, "operator")

	require.NoError(t, err)
	require.Equal(t, int64(-3), compiled.Get("v1").Value())
	require.Equal(t, int64(-1), compiled.Get("v2").Value())
	require.Equal(t, int64(0), compiled.Get("v3").Value())

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("operator").inv()`,
			`import("operator").inv("a")`,
			`import("operator").inv([])`,
			`import("operator").inv(undefined)`,
		}
		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "operator")
				require.Error(t, err)
			})
		}
	})
}

func TestShiftLeftRight(t *testing.T) {
	compiled, err := runWith(`
shift_left := import("operator").shift_left
shift_right := import("operator").shift_right

v1 := shift_left(2, 3)
v2 := shift_right(16, 2)
`, "operator")

	require.NoError(t, err)
	require.Equal(t, int64(16), compiled.Get("v1").Value())
	require.Equal(t, int64(4), compiled.Get("v2").Value())

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("operator").shift_left()`,
			`import("operator").shift_left("a", 2)`,
			`import("operator").shift_right([], {})`,
			`import("operator").shift_right(undefined, 3)`,
		}
		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "operator")
				require.Error(t, err)
			})
		}
	})
}

func TestItemgetter(t *testing.T) {
	compiled, err := runWith(`
itemgetter := import("operator").itemgetter

arr := [1, 2, 3, 4, 5]

v1 := itemgetter(0)(arr)
v2 := itemgetter(1)(arr)
v3 := itemgetter(2)(arr)
v4 := itemgetter(3)(arr)
v5 := itemgetter(4)(arr)

dict := {
	a: 1,
	b: 2,
	c: 3
}

v6 := itemgetter("a")(dict)
v7 := itemgetter("b")(dict)
v8 := itemgetter("c")(dict)
`, "operator")

	require.NoError(t, err)
	require.Equal(t, int64(1), compiled.Get("v1").Value())
	require.Equal(t, int64(2), compiled.Get("v2").Value())
	require.Equal(t, int64(3), compiled.Get("v3").Value())
	require.Equal(t, int64(4), compiled.Get("v4").Value())
	require.Equal(t, int64(5), compiled.Get("v5").Value())
	require.Equal(t, int64(1), compiled.Get("v6").Value())
	require.Equal(t, int64(2), compiled.Get("v7").Value())
	require.Equal(t, int64(3), compiled.Get("v8").Value())

	_, err = runWith(`
itemgetter := import("operator").itemgetter

itemgetter("ok")([1, 2, 3])
`, "operator")
	require.Error(t, err)

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("operator").itemgetter()`,
		}
		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "operator")
				require.Error(t, err)
			})
		}
	})
}
