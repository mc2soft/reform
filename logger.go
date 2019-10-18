package reform

import (
	"log"
	"reflect"
	"regexp"
	"time"

	"github.com/mc2soft/go-sqltypes"
)

type logLogger struct {
	*log.Logger
}

var (
	timeType, jsonType reflect.Type
	sqlPlaceholderRe   *regexp.Regexp
)

func (l *logLogger) Log(query string, args []interface{}) {
	values := make([]interface{}, len(args))

	// trying to make log look like SQL query, this is not supposed
	// to be correct translation to SQL
	for i, p := range args {
		r := reflect.ValueOf(p)
		if r.IsValid() {
			kind := r.Kind()

			// if value is pointer, replace it with NULL if
			// pointer is nil, otherwise use pointer value
			if kind == reflect.Ptr {
				if !r.IsNil() {
					r = r.Elem()
					p = r.Interface()
				} else {
					p = "NULL"
				}
			}

			// print time value as in SQL
			if kind == reflect.Struct && r.Type() == timeType {
				p = r.Interface().(time.Time).Format("2006-01-02 15:04:05")
				kind = reflect.String
			}

			// print json as string
			if r.Type() == jsonType {
				p = string(r.Interface().(sqltypes.JsonText))
				kind = reflect.String
			}

			// boolean as in PostgreSQL 't'/'f'
			if kind == reflect.Bool {
				if r.Bool() {
					p = "t"
				} else {
					p = "f"
				}
				kind = reflect.String
			}

			// string quoting, simplified with just single quotes
			if kind == reflect.String {
				p = "'" + p.(string) + "'"
			}
		}

		values[i] = p
	}

	query = sqlPlaceholderRe.ReplaceAllString(query, "%[$1]v")

	l.Printf(query, values...)
}

func NewLogLogger(l *log.Logger) Logger {
	return &logLogger{l}
}

func init() {
	sqlPlaceholderRe = regexp.MustCompile("\\$(\\d+)")
	timeType = reflect.TypeOf(time.Time{})
	jsonType = reflect.TypeOf(sqltypes.JsonText(nil))
}
