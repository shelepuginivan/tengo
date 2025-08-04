package stdlib_test

import (
	"testing"
	"time"

	"github.com/shelepuginivan/tengo"
	"github.com/shelepuginivan/tengo/require"

	. "github.com/shelepuginivan/tengo/stdlib"
)

func TestNew(t *testing.T) {
	compiled, err := runWith(`
datetime := import("datetime")
r := datetime.new(2025, datetime.july, 26, 23, 44, 35, 0, "Europe/Berlin")
`, "datetime")

	require.NoError(t, err)

	location, _ := time.LoadLocation("Europe/Berlin")
	expected := time.Date(2025, time.July, 26, 23, 44, 35, 0, location)
	actual, ok := ToTime(compiled.Get("r").Object())

	require.True(t, ok)
	require.Equal(t, expected, actual.GoTime())

	_, err = runWith(`
datetime := import("datetime")
r := datetime.new(2025, datetime.july, 26, 23, 44, 35, 0, "Haro/Hawayu")
`, "datetime")

	require.Error(t, err)

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("datetime").new()`,
			`import("datetime").new(1, 1, 1, 1, 1, 1, 1, 1, 1)`,
			`import("datetime").new("this", 12, 26, 20, 04, 37)`,
			`import("datetime").new(2003, "was", 26, 20, 04, 37)`,
			`import("datetime").new(2003, 12, "originally", 20, 04, 37)`,
			`import("datetime").new(2003, 12, 26, "created", 04, 37)`,
			`import("datetime").new(2003, 12, 26, 20, "for", 37)`,
			`import("datetime").new(2003, 12, 26, 20, 04, "Saya")`,
			`import("datetime").new(2003, 12, 26, 20, 04, 37, "widget")`,
			`import("datetime").new(2003, 12, 26, 20, 04, 37, 0, "system")`,
			`import("datetime").new(2003, 12, 26, 20, 04, 37, 0, undefined)`,
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "datetime")
				require.Error(t, err)
			})
		}
	})
}

func TestNow(t *testing.T) {
	beforeRun := time.Now()

	compiled, err := runWith(`r := import("datetime").now()`, "datetime")
	require.NoError(t, err)

	afterRun := time.Now()

	r, ok := compiled.Get("r").Object().(*Time)
	require.True(t, ok)

	whenRun := r.GoTime()

	require.True(t, beforeRun.Before(whenRun))
	require.True(t, afterRun.After(whenRun))

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("datetime").now(undefined)`,
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "datetime")
				require.Error(t, err)
			})
		}
	})
}

func TestParse(t *testing.T) {
	compiled, err := runWith(`
datetime := import("datetime")
r := datetime.parse(datetime.rfc_email, "Sat, 26 Jul 2025 23:51:29 +0300")
`, "datetime")

	require.NoError(t, err)

	expected, _ := time.Parse(time.RFC1123Z, "Sat, 26 Jul 2025 23:51:29 +0300")
	actual, ok := ToTime(compiled.Get("r").Object())

	require.True(t, ok)
	require.Equal(t, expected, actual.GoTime())

	_, err = runWith(`
datetime := import("datetime")
r := datetime.parse(datetime.rfc_email, "Sat Jul 26 11:51:03 PM MSK 2025")
`, "datetime")

	require.Error(t, err)

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("datetime").parse("25-07-22")`,
			`import("datetime").parse(undefined, "23:21")`,
			`import("datetime").parse("", undefined)`,
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "datetime")
				require.Error(t, err)
			})
		}
	})
}

func TestParseDuration(t *testing.T) {
	compiled, err := runWith(`
datetime := import("datetime")
r := datetime.parse_duration("1h30m20s14ms2us234ns")
`, "datetime")

	require.NoError(t, err)

	expected := time.Hour +
		30*time.Minute +
		20*time.Second +
		14*time.Millisecond +
		2*time.Microsecond +
		234*time.Nanosecond

	actual, ok := tengo.ToInt(compiled.Get("r").Object())

	require.True(t, ok)
	require.Equal(t, expected, time.Duration(actual))

	_, err = runWith(`
datetime := import("datetime")
r := datetime.parse_duration("foo bar baz whatever")
`, "datetime")

	require.Error(t, err)

	t.Run("Signature", func(t *testing.T) {
		tests := []string{
			`import("datetime").parse_duration()`,
			`import("datetime").parse_duration(undefined)`,
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				_, err := runWith(tt, "datetime")
				require.Error(t, err)
			})
		}
	})
}
