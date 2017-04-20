package gohessian

import (
	"reflect"

	"froad.com/server/common/log"
)

// help extracting map data into struct
func extractData(data reflect.Value, typ reflect.Type) (rslt reflect.Value) {
	rslt = reflect.New(typ)
	value := rslt.Elem()
	defer func() {
		rslt = rslt.Elem()
		if r := recover(); r != nil {
			log.Debug("Recovered in f", r)
		}
		log.Trace("rslt: ", rslt)
	}()

	for value.Kind() == reflect.Ptr {
		typ = typ.Elem()
		value.Set(reflect.New(typ))
		value = value.Elem()
	}

	for data.Type().Kind() == reflect.Interface ||
		(data.Type().Kind() == reflect.Ptr && !reflect.ValueOf(data).IsNil()) {
		data = data.Elem()
	}
	log.Trace("data: ", data.Interface(), ", kind: ", data.Kind())
	log.Trace("type: ", typ, ", kind: ", typ.Kind())

	switch typ.Kind() {
	case reflect.Struct:
		if data.Kind() != reflect.Map {
			return
		}
		value.Set(extractStruct(data.Interface(), typ))
	case reflect.Slice:
		if data.Kind() != reflect.Slice {
			return
		}
		value.Set(extractSlice(data.Interface(), typ))
	case reflect.Map:
		if data.Kind() != reflect.Map {
			return
		}
		value.Set(extractMap(data.Interface(), typ))
	default:
		if data.Kind() == reflect.Map {
			k := data.MapKeys()[0]
			data = data.MapIndex(k)
		}
		value.Set(data)
	}
	return
}

func extractStruct(data interface{}, typ reflect.Type) (value reflect.Value) {
	if typ.Kind() != reflect.Struct {
		return value
	}
	dataMap := data.(map[interface{}]interface{})
	value = reflect.New(typ).Elem()
	for i := 0; i < typ.NumField(); i++ {
		tf := typ.Field(i)
		vf := value.Field(i)
		if !vf.CanSet() {
			return value
		}

		if tf.Type.Name() == nameTypeName {
			continue // pass hessian name field
		}

		var name string
		var vd interface{}
		var ok bool
		tag := tf.Tag.Get(hessianTag)
		if len(tag) > 0 {
			d, ok := dataMap[tag]
			if ok {
				name = tag
				vd = d
			}
		}
		d, ok := dataMap[tf.Name]
		if ok {
			name = tf.Name
			vd = d
		}
		if len(name) <= 0 || vd == nil {
			continue
		}
		log.Trace("parsing: ", name)
		vf.Set(extractData(reflect.ValueOf(vd), tf.Type))
	}
	return value
}

func extractSlice(data interface{}, typ reflect.Type) (value reflect.Value) {
	dataSlice := reflect.ValueOf(data)
	value = reflect.MakeSlice(typ, 0, dataSlice.Len())
	for i := 0; i < dataSlice.Len(); i++ {
		v := extractData(dataSlice.Index(i), typ.Elem())
		value = reflect.Append(value, v)
	}
	return value
}

func extractMap(data interface{}, typ reflect.Type) (value reflect.Value) {
	value = reflect.MakeMap(typ)
	keys := reflect.ValueOf(data).MapKeys()
	for _, kd := range keys {
		kv := extractData(kd, typ.Key())
		vv := extractData(reflect.ValueOf(data).MapIndex(kd), typ.Elem())
		value.SetMapIndex(kv, vv)
	}
	return value
}
