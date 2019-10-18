package reform

import (
	"github.com/mc2soft/go-sqltypes"
	"reflect"
	"time"
)

func internalCopy(ptr reflect.Value, val reflect.Value) {
	switch val.Type().Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return
		}
		newval := reflect.New(val.Type().Elem())
		internalCopy(newval, val.Elem())
		ptr.Elem().Set(newval)
	case reflect.Slice:
		if val.IsNil() {
			return
		}
		newval := reflect.MakeSlice(val.Type(), val.Len(), val.Cap())
		for i := 0; i < val.Len(); i++ {
			newval.Index(i).Set(val.Index(i))
		}
		ptr.Elem().Set(newval)
	default:
		ptr.Elem().Set(val)
	}
}

// Copy does semi-deep copying of Struct, slices and pointers
func Copy(s Struct) Struct {
	typ := reflect.ValueOf(s).Type().Elem()

	result := reflect.New(typ).Interface().(Struct)

	for i := range s.Values() {
		ptr := result.Pointers()[i]
		val := s.Values()[i]

		internalCopy(reflect.ValueOf(ptr), reflect.ValueOf(val))
	}

	return result
}

// ChangedFields calculates differences in fields between o and n
func ChangedFields(o Struct, n Struct) []string {
	result := []string{}

	for i := range o.Values() {
		oval, nval := o.Values()[i], n.Values()[i]

		// dereference pointer if this is non-nil pointer
		typ := reflect.ValueOf(nval).Type()
		if typ.Kind() == reflect.Ptr {
			oref := reflect.ValueOf(oval)
			nref := reflect.ValueOf(nval)
			if !oref.IsNil() && !nref.IsNil() {
				oval = oref.Elem().Interface()
				nval = nref.Elem().Interface()
			}
		}

		var equal bool

		switch reflect.ValueOf(oval).Type() {
		case timeType:
			// compare times without timezone, as we have everything in UTC
			equal = oval.(time.Time).Equal(nval.(time.Time))
		case jsonType:
			var o, n interface{}
			oval.(sqltypes.JsonText).Unmarshal(&o)
			nval.(sqltypes.JsonText).Unmarshal(&n)
			equal = reflect.DeepEqual(o, n)
		default:
			equal = reflect.DeepEqual(oval, nval)
		}

		if !equal {
			result = append(result, o.View().Columns()[i])
		}
	}

	return result
}
