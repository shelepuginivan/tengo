package stdlib

import (
	"time"

	"github.com/shelepuginivan/tengo"
)

var datetimeModule = map[string]tengo.Object{
	// Time layout constants.
	"ansic": &tengo.String{
		Value: time.ANSIC,
	},
	"unix_date": &tengo.String{
		Value: time.UnixDate,
	},
	"ruby_date": &tengo.String{
		Value: time.RubyDate,
	},
	"rfc_822": &tengo.String{
		Value: time.RFC822,
	},
	"rfc_822_z": &tengo.String{
		Value: time.RFC822Z,
	},
	"rfc_850": &tengo.String{
		Value: time.RFC850,
	},
	"rfc_1123": &tengo.String{
		Value: time.RFC1123,
	},
	"rfc_email": &tengo.String{
		Value: time.RFC1123Z,
	},
	"iso": &tengo.String{
		Value: time.RFC3339,
	},
	"iso_nano": &tengo.String{
		Value: time.RFC3339Nano,
	},
	"kitchen": &tengo.String{
		Value: time.Kitchen,
	},
	"stamp": &tengo.String{
		Value: time.Stamp,
	},
	"stamp_milli": &tengo.String{
		Value: time.StampMilli,
	},
	"stamp_micro": &tengo.String{
		Value: time.StampMicro,
	},
	"stamp_nano": &tengo.String{
		Value: time.StampNano,
	},
	"date_time": &tengo.String{
		Value: time.DateTime,
	},
	"date_only": &tengo.String{
		Value: time.DateOnly,
	},
	"time_only": &tengo.String{
		Value: time.TimeOnly,
	},

	// Time unit constants.
	"nanosecond": &tengo.Int{
		Value: int64(time.Nanosecond),
	},
	"microsecond": &tengo.Int{
		Value: int64(time.Microsecond),
	},
	"millisecond": &tengo.Int{
		Value: int64(time.Millisecond),
	},
	"second": &tengo.Int{
		Value: int64(time.Second),
	},
	"minute": &tengo.Int{
		Value: int64(time.Minute),
	},
	"hour": &tengo.Int{
		Value: int64(time.Hour),
	},

	// Weekday constants.
	"sunday": &tengo.Int{
		Value: int64(time.Sunday),
	},
	"monday": &tengo.Int{
		Value: int64(time.Monday),
	},
	"tuesday": &tengo.Int{
		Value: int64(time.Tuesday),
	},
	"wednesday": &tengo.Int{
		Value: int64(time.Wednesday),
	},
	"thursday": &tengo.Int{
		Value: int64(time.Thursday),
	},
	"friday": &tengo.Int{
		Value: int64(time.Friday),
	},
	"saturday": &tengo.Int{
		Value: int64(time.Saturday),
	},

	// Month constants.
	"january": &tengo.Int{
		Value: int64(time.January),
	},
	"february": &tengo.Int{
		Value: int64(time.February),
	},
	"march": &tengo.Int{
		Value: int64(time.March),
	},
	"april": &tengo.Int{
		Value: int64(time.April),
	},
	"may": &tengo.Int{
		Value: int64(time.May),
	},
	"june": &tengo.Int{
		Value: int64(time.June),
	},
	"july": &tengo.Int{
		Value: int64(time.July),
	},
	"august": &tengo.Int{
		Value: int64(time.August),
	},
	"september": &tengo.Int{
		Value: int64(time.September),
	},
	"october": &tengo.Int{
		Value: int64(time.October),
	},
	"november": &tengo.Int{
		Value: int64(time.November),
	},
	"december": &tengo.Int{
		Value: int64(time.December),
	},

	// Functions.
	"now": &tengo.UserFunction{
		Name:  "now",
		Value: datetimeNow,
	},
	"new": &tengo.UserFunction{
		Name:  "new",
		Value: datetimeNewTime,
	},
	"parse": &tengo.UserFunction{
		Name:  "parse",
		Value: datetimeParse,
	},
	"parse_duration": &tengo.UserFunction{
		Name:  "parse_duration",
		Value: datetimeParseDuration,
	},
}

func datetimeNow(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	return CreateTime(time.Now()), nil
}

func datetimeParse(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	layout, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "layout",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	value, ok := tengo.ToString(args[1])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "value",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	parsed, err := time.Parse(layout, value)
	if err != nil {
		return nil, err
	}

	return CreateTime(parsed), nil
}

func datetimeNewTime(args ...tengo.Object) (tengo.Object, error) {
	if len(args) < 6 || len(args) > 8 {
		return nil, tengo.ErrWrongNumArguments
	}

	var (
		location   = time.Local
		nanosecond = 0
	)

	year, ok := tengo.ToInt(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "year",
			Expected: "int",
			Found:    args[0].TypeName(),
		}
	}

	month, ok := tengo.ToInt(args[1])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "month",
			Expected: "int",
			Found:    args[1].TypeName(),
		}
	}

	day, ok := tengo.ToInt(args[2])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "days",
			Expected: "int",
			Found:    args[2].TypeName(),
		}
	}

	hour, ok := tengo.ToInt(args[3])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "hour",
			Expected: "int",
			Found:    args[3].TypeName(),
		}
	}

	minute, ok := tengo.ToInt(args[4])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "minute",
			Expected: "int",
			Found:    args[4].TypeName(),
		}
	}

	second, ok := tengo.ToInt(args[5])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "second",
			Expected: "int",
			Found:    args[5].TypeName(),
		}
	}

	if len(args) >= 7 {
		nsec, ok := tengo.ToInt(args[6])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "nanosecond",
				Expected: "int",
				Found:    args[6].TypeName(),
			}
		}

		nanosecond = nsec
	}

	if len(args) == 8 {
		loc, ok := tengo.ToString(args[7])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "location",
				Expected: "string",
				Found:    args[7].TypeName(),
			}
		}

		parsedLoc, err := time.LoadLocation(loc)
		if err != nil {
			return nil, err
		}

		location = parsedLoc
	}

	return CreateTime(time.Date(
		year,
		time.Month(month),
		day,
		hour,
		minute,
		second,
		nanosecond,
		location,
	)), nil
}

func datetimeParseDuration(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	durationString, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "duration",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	duration, err := time.ParseDuration(durationString)
	if err != nil {
		return nil, err
	}

	return &tengo.Int{Value: int64(duration)}, nil
}
