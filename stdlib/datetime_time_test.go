package stdlib_test

import (
	"testing"
	"time"

	"github.com/shelepuginivan/tengo/require"

	. "github.com/shelepuginivan/tengo/stdlib"
)

func TestToTime(t *testing.T) {
	compiled, err := runWith(`
datetime := import("datetime")

valid := datetime.now()
invalid := "????"
`, "datetime")

	require.NoError(t, err)

	_, ok := ToTime(compiled.Get("valid").Object())
	require.True(t, ok)

	_, ok = ToTime(compiled.Get("invalid").Object())
	require.False(t, ok)
}

func TestTime(t *testing.T) {
	t.Run("Type name", func(t *testing.T) {
		compiled, err := runWith(`r := type_name(import("datetime").now())`, "datetime")

		require.NoError(t, err)
		require.Equal(t, "datetime.Time", compiled.Get("r").Value())
	})

	t.Run("String", func(t *testing.T) {
		compiled, err := runWith(`
datetime := import("datetime")
v1 := string(datetime.new(2003, datetime.december, 26, 20, 04, 37, 0, "UTC"))
v2 := string(datetime.new(2025, datetime.july, 29, 23, 27, 34, 0, "UTC"))
`, "datetime")

		require.NoError(t, err)
		require.Equal(t,
			"datetime.Time(Fri, 26 Dec 2003 20:04:37 +0000)",
			compiled.Get("v1").Value(),
		)
		require.Equal(t,
			"datetime.Time(Tue, 29 Jul 2025 23:27:34 +0000)",
			compiled.Get("v2").Value(),
		)
	})

	t.Run("Methods", func(t *testing.T) {
		compiled, err := runWith(`
datetime := import("datetime")

base := datetime.new(2003, datetime.december, 26, 20, 04, 37)

t1 := base.add_date(1, 6, -20)

v1 := base.format(datetime.date_time)
v2 := base.nanosecond()
v3 := base.second()
v4 := base.minute()
v5 := base.hour()

__clock := base.clock()
__date := base.date()

v6  := __clock[0]
v7  := __clock[1]
v8  := __clock[2]
v9  := __date[0]
v10  := __date[1]
v11 := __date[2]
v12 := base.week_day()
v13 := base.year_day()
v14 := base.month()
v15 := base.unix()

utc := base.utc()
`, "datetime")

		require.NoError(t, err)

		baseExpected := time.Date(2003, time.December, 26, 20, 04, 37, 0, time.Local)
		baseActual, ok := ToTime(compiled.Get("base").Object())
		require.True(t, ok)
		require.Equal(t, baseExpected, baseActual.GoTime())

		t1, ok := ToTime(compiled.Get("t1").Object())
		require.True(t, ok)
		require.Equal(t, time.Date(2005, time.June, 6, 20, 04, 37, 0, time.Local), t1.GoTime())

		require.Equal(t, "2003-12-26 20:04:37", compiled.Get("v1").Value())
		require.Equal(t, int64(0), compiled.Get("v2").Value())
		require.Equal(t, int64(37), compiled.Get("v3").Value())
		require.Equal(t, int64(4), compiled.Get("v4").Value())
		require.Equal(t, int64(20), compiled.Get("v5").Value())
		require.Equal(t, int64(20), compiled.Get("v6").Value())
		require.Equal(t, int64(4), compiled.Get("v7").Value())
		require.Equal(t, int64(37), compiled.Get("v8").Value())
		require.Equal(t, int64(2003), compiled.Get("v9").Value())
		require.Equal(t, int64(12), compiled.Get("v10").Value())
		require.Equal(t, int64(26), compiled.Get("v11").Value())
		require.Equal(t, int64(5), compiled.Get("v12").Value())
		require.Equal(t, int64(360), compiled.Get("v13").Value())
		require.Equal(t, int64(12), compiled.Get("v14").Value())
		require.Equal(t, baseExpected.Unix(), compiled.Get("v15").Value())

		utc, ok := ToTime(compiled.Get("utc").Object())
		require.True(t, ok)
		require.Equal(t, baseExpected.UTC(), utc.GoTime())
	})

	t.Run("Operators", func(t *testing.T) {
		compiled, err := runWith(`
datetime := import("datetime")

base := datetime.new(2003, datetime.december, 26, 20, 04, 37)
now := datetime.now()

v1 := base > now
v2 := base < now
v3 := base >= now
v4 := base <= now
v5 := base == now
v6 := base != now
v7 := base == base
v8 := base == "not a date"
v9 := bool(base)
v10 := bool(datetime.new(1, 1, 1, 0, 0, 0, 0, "UTC"))

t1 := base + datetime.hour
t2 := base - datetime.minute
`, "datetime")

		require.NoError(t, err)

		require.Equal(t, false, compiled.Get("v1").Value())
		require.Equal(t, true, compiled.Get("v2").Value())
		require.Equal(t, false, compiled.Get("v3").Value())
		require.Equal(t, true, compiled.Get("v4").Value())
		require.Equal(t, false, compiled.Get("v5").Value())
		require.Equal(t, true, compiled.Get("v6").Value())
		require.Equal(t, true, compiled.Get("v7").Value())
		require.Equal(t, false, compiled.Get("v8").Value())
		require.Equal(t, true, compiled.Get("v9").Value())
		require.Equal(t, false, compiled.Get("v10").Value())

		t1, ok := ToTime(compiled.Get("t1").Object())
		require.True(t, ok)
		require.Equal(t, time.Date(2003, time.December, 26, 21, 04, 37, 0, time.Local), t1.GoTime())

		t2, ok := ToTime(compiled.Get("t2").Object())
		require.True(t, ok)
		require.Equal(t, time.Date(2003, time.December, 26, 20, 03, 37, 0, time.Local), t2.GoTime())

		scripts := []string{
			`
datetime := import("datetime")
datetime.now() > "something"
`,
			`
datetime := import("datetime")
datetime.now() / datetime.now()
`,
			`
datetime := import("datetime")
datetime.now() / datetime.second
`,
		}

		for _, script := range scripts {
			_, err := runWith(script, "datetime")
			require.Error(t, err)
		}
	})

	t.Run("Signatures", func(t *testing.T) {
		tests := []struct {
			name   string
			script string
		}{
			{
				name: "date",
				script: `
datetime := import("datetime")
datetime.now().date(false)
`,
			},
			{
				name: "add_date",
				script: `
datetime := import("datetime")
datetime.now().add_date()
`,
			},
			{
				name: "add_date-arg1",
				script: `
datetime := import("datetime")
datetime.now().add_date("", -1, 23)
`,
			},
			{
				name: "add_date-arg2",
				script: `
datetime := import("datetime")
datetime.now().add_date(29, "foo", 0)
`,
			},
			{
				name: "add_date-arg3",
				script: `
datetime := import("datetime")
datetime.now().add_date(7, -15, [1, 2])
`,
			},
			{
				name: "clock",
				script: `
datetime := import("datetime")
datetime.now().clock([])
`,
			},
			{
				name: "utc",
				script: `
datetime := import("datetime")
datetime.now().utc("ong fr")
`,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := runWith(tt.script, "datetime")
				require.Error(t, err)
			})
		}
	})
}
