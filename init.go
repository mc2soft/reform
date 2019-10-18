package reform

import (
	"reflect"
	"strings"
)

func InitView(v interface{}) (name string, lastPkIdx int, cols []string, oe []bool, zv []interface{}) {
	t := reflect.ValueOf(v).Elem().Type()
	f := t.Field(0)
	if !f.Anonymous {
		panic(v)
	}

	sql := strings.Split(f.Tag.Get("sql"), ".")
	name = sql[len(sql)-1]

	for i := 1; i < t.NumField(); i++ {
		f := t.Field(i)
		sql = strings.Split(f.Tag.Get("sql"), ",")
		if sql[0] == "" {
			// skip fields without sql tag
			continue
		}
		cols = append(cols, strings.TrimSpace(sql[0]))
		if len(sql) > 1 {
			if sql[1] == "pk" {
				if lastPkIdx != i-1 {
					panic("primary keys should come at the top")
				}
				lastPkIdx = i
			}
			oe = append(oe, (sql[1] == "omitempty") || (sql[1] == "pk"))
		} else {
			oe = append(oe, false)
		}
		zv = append(zv, reflect.Zero(f.Type).Interface())
	}

	return
}
