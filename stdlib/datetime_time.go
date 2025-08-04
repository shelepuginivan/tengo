package stdlib

import (
	"time"

	"github.com/shelepuginivan/tengo"
	"github.com/shelepuginivan/tengo/token"
)

type Time struct {
	*tengo.ImmutableMap

	time time.Time
}

func (t *Time) BinaryOp(op token.Token, rhs tengo.Object) (tengo.Object, error) {
	d, ok := tengo.ToInt64(rhs)
	if ok {
		return t.arithmeticOp(op, time.Duration(d))
	}

	u, ok := rhs.(*Time)
	if ok {
		return t.compareOp(op, u)
	}

	return nil, tengo.ErrInvalidOperator
}

func (t *Time) arithmeticOp(op token.Token, rhs time.Duration) (tengo.Object, error) {
	switch op {
	case token.Add:
		return CreateTime(t.time.Add(rhs)), nil
	case token.Sub:
		return CreateTime(t.time.Add(-rhs)), nil
	}

	return nil, tengo.ErrInvalidOperator
}

func (t *Time) compareOp(op token.Token, rhs *Time) (tengo.Object, error) {
	switch op {
	case token.Greater:
		return TengoBool(t.time.After(rhs.time)), nil
	case token.GreaterEq:
		return TengoBool(t.time.After(rhs.time) || t.time.Equal(rhs.time)), nil
	case token.Less:
		return TengoBool(t.time.Before(rhs.time)), nil
	case token.LessEq:
		return TengoBool(t.time.Before(rhs.time) || t.time.Equal(rhs.time)), nil
	default:
		return nil, tengo.ErrInvalidOperator
	}
}

func (t *Time) Equals(o tengo.Object) bool {
	u, ok := ToTime(o)
	if !ok {
		return false
	}

	return t.time.Equal(u.time)
}

func (t *Time) TypeName() string {
	return "datetime.Time"
}

func (t *Time) String() string {
	return t.time.Format("datetime.Time(Mon, 02 Jan 2006 15:04:05 -0700)")
}

func (t *Time) IsFalsy() bool {
	return t.time.IsZero()
}

func (t *Time) GoTime() time.Time {
	return t.time
}

func CreateTime(t time.Time) *Time {
	members := map[string]tengo.Object{
		// Time arithmetic.
		"add_date": &tengo.UserFunction{
			Name:  "add_date",
			Value: timeAddDate(t),
		},

		// Time unit functions.
		"clock": &tengo.UserFunction{
			Name:  "clock",
			Value: timeClock(t),
		},
		"nanosecond": &tengo.UserFunction{
			Name:  "nanosecond",
			Value: FuncARI(t.Nanosecond),
		},
		"second": &tengo.UserFunction{
			Name:  "second",
			Value: FuncARI(t.Second),
		},
		"minute": &tengo.UserFunction{
			Name:  "minute",
			Value: FuncARI(t.Minute),
		},
		"hour": &tengo.UserFunction{
			Name:  "hour",
			Value: FuncARI(t.Hour),
		},
		"date": &tengo.UserFunction{
			Name:  "date",
			Value: timeDate(t),
		},
		"day": &tengo.UserFunction{
			Name:  "day",
			Value: FuncARI(t.Day),
		},
		"month": &tengo.UserFunction{
			Name: "month",
			Value: FuncARI(func() int {
				return int(t.Month())
			}),
		},
		"year": &tengo.UserFunction{
			Name:  "year",
			Value: FuncARI(t.Year),
		},
		"week_day": &tengo.UserFunction{
			Name: "week_day",
			Value: FuncARI(func() int {
				return int(t.Weekday())
			}),
		},
		"year_day": &tengo.UserFunction{
			Name:  "year_day",
			Value: FuncARI(t.YearDay),
		},

		// Misc.
		"format": &tengo.UserFunction{
			Name:  "format",
			Value: FuncASRS(t.Format),
		},
		"unix": &tengo.UserFunction{
			Name:  "unix",
			Value: FuncARI64(t.Unix),
		},
		"unix_milli": &tengo.UserFunction{
			Name:  "unix_milli",
			Value: FuncARI64(t.UnixMilli),
		},
		"unix_micro": &tengo.UserFunction{
			Name:  "unix_micro",
			Value: FuncARI64(t.UnixMicro),
		},
		"unix_nano": &tengo.UserFunction{
			Name:  "unix_nano",
			Value: FuncARI64(t.UnixNano),
		},
		"utc": &tengo.UserFunction{
			Name:  "utc",
			Value: timeUTC(t),
		},
		"is_zero": &tengo.UserFunction{
			Name:  "is_zero",
			Value: FuncARB(t.IsZero),
		},
	}

	return &Time{
		ImmutableMap: &tengo.ImmutableMap{
			Value: members,
		},
		time: t,
	}
}

func ToTime(o tengo.Object) (*Time, bool) {
	time, ok := o.(*Time)
	if ok {
		return time, true
	}

	return nil, false
}

func timeDate(t time.Time) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 0 {
			return nil, tengo.ErrWrongNumArguments
		}

		year, month, day := t.Date()

		return &tengo.ImmutableArray{
			Value: []tengo.Object{
				&tengo.Int{Value: int64(year)},
				&tengo.Int{Value: int64(month)},
				&tengo.Int{Value: int64(day)},
			},
		}, nil
	}
}

func timeAddDate(t time.Time) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 3 {
			return nil, tengo.ErrWrongNumArguments
		}

		years, ok := tengo.ToInt(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "years",
				Expected: "int",
				Found:    args[0].TypeName(),
			}
		}

		month, ok := tengo.ToInt(args[1])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "months",
				Expected: "int",
				Found:    args[1].TypeName(),
			}
		}

		days, ok := tengo.ToInt(args[2])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "days",
				Expected: "int",
				Found:    args[2].TypeName(),
			}
		}

		return CreateTime(t.AddDate(years, month, days)), nil
	}
}

func timeClock(t time.Time) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 0 {
			return nil, tengo.ErrWrongNumArguments
		}

		hour, minute, second := t.Clock()

		return &tengo.ImmutableArray{
			Value: []tengo.Object{
				&tengo.Int{Value: int64(hour)},
				&tengo.Int{Value: int64(minute)},
				&tengo.Int{Value: int64(second)},
			},
		}, nil
	}
}

func timeUTC(t time.Time) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 0 {
			return nil, tengo.ErrWrongNumArguments
		}

		return CreateTime(t.UTC()), nil
	}
}
