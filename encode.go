// Encode values of types to hessian protocol 2.0:
//	- int, int32, int64
//	- float64
//	- bool
//	- time.Time
//	- []byte
//	- slice
//	- array
//	- map
//	- struct
//	- nil
package gohessian

import (
	"bytes"
	"errors"
	"reflect"
	"time"
	"unicode/utf8"

	log "github.com/cihub/seelog"
)

type Encoder struct {
}

type HessianName struct{}

const (
	CHUNK_SIZE    = 0x8000
	ENCODER_DEBUG = false
	hessianTag    = "hs"
	nameTypeName  = "HessianName"
	fieldName     = "Name"
)

//func init() {
//	_, filename, _, _ := runtime.Caller(1)
//	if ENCODER_DEBUG {
//		log.SetPrefix(filename + "\n")
//	}
//}

// Encode do encode var to binary under hessian protocol
func Encode(v interface{}) (b []byte, err error) {
	t := reflect.TypeOf(v)

	// dereference any pointer
	for t.Kind() == reflect.Ptr {
		if reflect.ValueOf(v).IsNil() {
			return encodeNull(v)
		}
		v = reflect.ValueOf(v).Elem().Interface()
	}

	// basic types
	switch v.(type) {
	case []byte:
		b, err = encodeBinary(v.([]byte))

	case bool:
		b, err = encodeBool(v.(bool))

	case time.Time:
		b, err = encodeTime(v.(time.Time))

	case float64:
		b, err = encodeFloat64(v.(float64))

	case int:
		if v.(int) >= -2147483648 && v.(int) <= 2147483647 {
			b, err = encodeInt32(int32(v.(int)))
		} else {
			b, err = encodeInt64(int64(v.(int)))
		}

	case int32:
		b, err = encodeInt32(v.(int32))

	case int64:
		b, err = encodeInt64(v.(int64))

	case string:
		b, err = encodeString(v.(string))
	}

	// reference types
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		b, err = encodeList(v)

	case reflect.Struct:
		b, err = encodeStruct(v)

	case reflect.Map:
		b, err = encodeMap(v)

	default:
		return nil, errors.New("unkown kind")
	}

	if ENCODER_DEBUG {
		log.Debug(SprintHex(b))
	}
	return
}

// encodeBinary binary
func encodeBinary(v []byte) (b []byte, err error) {
	var (
		tag  byte
		lenB []byte
		lenN int
	)

	if len(v) == 0 {
		if lenB, err = PackUint16(0); err != nil {
			b = nil
			return
		}
		b = append(b, 'B')
		b = append(b, lenB...)
		return
	}

	rBuf := *bytes.NewBuffer(v)

	for rBuf.Len() > 0 {
		if rBuf.Len() > CHUNK_SIZE {
			tag = 'b'
			if lenB, err = PackUint16(uint16(CHUNK_SIZE)); err != nil {
				b = nil
				return
			}
			lenN = CHUNK_SIZE
		} else {
			tag = 'B'
			if lenB, err = PackUint16(uint16(rBuf.Len())); err != nil {
				b = nil
				return
			}
			lenN = rBuf.Len()
		}
		b = append(b, tag)
		b = append(b, lenB...)
		b = append(b, rBuf.Next(lenN)...)
	}
	return
}

// encodeBool encode boolean
func encodeBool(v bool) (b []byte, err error) {
	if v == true {
		b = append(b, 'T')
		return
	}
	b = append(b, 'F')
	return
}

// encodeTime encode date
func encodeTime(v time.Time) (b []byte, err error) {
	var tmpV []byte
	b = append(b, 'd')
	if tmpV, err = PackInt64(v.UnixNano() / 1000000); err != nil {
		b = nil
		return
	}
	b = append(b, tmpV...)
	return
}

// encodeFloat64 encode double
func encodeFloat64(v float64) (b []byte, err error) {
	var tmpV []byte
	if tmpV, err = PackFloat64(v); err != nil {
		b = nil
		return
	}
	b = append(b, 'D')
	b = append(b, tmpV...)
	return
}

// encodeInt32 encode int
func encodeInt32(v int32) (b []byte, err error) {
	var tmpV []byte
	if tmpV, err = PackInt32(v); err != nil {
		b = nil
		return
	}
	b = append(b, 'I')
	b = append(b, tmpV...)
	return
}

// encodeInt64 encode long
func encodeInt64(v int64) (b []byte, err error) {
	var tmpV []byte
	if tmpV, err = PackInt64(v); err != nil {
		b = nil
		return
	}
	b = append(b, 'L')
	b = append(b, tmpV...)
	return

}

// encodeNull encode null
func encodeNull(v interface{}) (b []byte, err error) {
	b = append(b, 'N')
	return
}

// encodeString encode string
func encodeString(v string) (b []byte, err error) {
	var (
		lenB []byte
		sBuf = *bytes.NewBufferString(v)
		rLen = utf8.RuneCountInString(v)

		sChunk = func(_len int) {
			for i := 0; i < _len; i++ {
				if r, s, err := sBuf.ReadRune(); s > 0 && err == nil {
					b = append(b, []byte(string(r))...)
				}
			}
		}
	)

	if "" == v {
		if lenB, err = PackUint16(uint16(rLen)); err != nil {
			b = nil
			return
		}
		b = append(b, 'S')
		b = append(b, lenB...)
		b = append(b, []byte{}...)
		return
	}

	for {
		rLen = utf8.RuneCount(sBuf.Bytes())
		if rLen == 0 {
			break
		}
		if rLen > CHUNK_SIZE {
			if lenB, err = PackUint16(uint16(CHUNK_SIZE)); err != nil {
				b = nil
				return
			}
			b = append(b, 's')
			b = append(b, lenB...)
			sChunk(CHUNK_SIZE)
		} else {
			if lenB, err = PackUint16(uint16(rLen)); err != nil {
				b = nil
				return
			}
			b = append(b, 'S')
			b = append(b, lenB...)
			sChunk(rLen)
		}
	}
	return
}

// encodeList encode list for slice and array
func encodeList(in interface{}) (b []byte, err error) {
	if reflect.TypeOf(in).Kind() != reflect.Slice && reflect.TypeOf(in).Kind() != reflect.Array {
		return nil, errors.New("invalid slice")
	}
	b = append(b, 'V')

	v := reflect.ValueOf(in)
	b_len, err := PackInt32(int32(v.Len()))
	if err != nil {
		return nil, err
	}
	b = append(b, 'l')
	b = append(b, b_len...)

	for i := 0; i < v.Len(); i++ {
		tmp, err := Encode(v.Index(i).Interface())
		if nil != err {
			log.Error(err)
			return nil, err
		}
		b = append(b, tmp...)
	}
	b = append(b, 'z')
	return b, nil
}

// encodeStruct encode struct as map
func encodeStruct(in interface{}) (b []byte, err error) {
	if reflect.TypeOf(in).Kind() != reflect.Struct {
		return nil, errors.New("invalid struct")
	}

	v := reflect.ValueOf(in)
	t := reflect.TypeOf(in)
	name := getStructName(t)
	b = append(b, 'M', 't')
	l_name, err := PackInt16(int16(len(name)))
	if err != nil {
		return nil, err
	}
	b = append(b, l_name...)
	b = append(b, name...)
	for i := 0; i < t.NumField(); i++ {
		tag := getFieldTag(t.Field(i))
		if t.Field(i).Type.Name() == nameTypeName {
			continue // pass hessian name field
		}
		tmp_k, err := encodeString(tag)
		if err != nil {
			return nil, err
		}
		tmp_v, err := Encode(v.Field(i).Interface())
		if err != nil {
			return nil, err
		}
		b = append(b, tmp_k...)
		b = append(b, tmp_v...)
	}
	b = append(b, 'z')
	return b, nil
}

// encodeMap encode map
func encodeMap(in interface{}) (b []byte, err error) {
	if reflect.TypeOf(in).Kind() != reflect.Map {
		return nil, errors.New("invalid map")
	}
	b = append(b, 'M')
	v := reflect.ValueOf(in)

	for _, key := range v.MapKeys() {
		tmp_k, err := Encode(key.Interface())
		if err != nil {
			return nil, err
		}
		tmp_v, err := Encode(v.MapIndex(key).Interface())
		if err != nil {
			return nil, err
		}
		b = append(b, tmp_k...)
		b = append(b, tmp_v...)
	}
	b = append(b, 'z')
	return b, nil
}

// getStructName return struct name
func getStructName(t reflect.Type) (name string) {
	nameField, ok := t.FieldByName(fieldName)
	if ok {
		name = nameField.Tag.Get(hessianTag)
	} else {
		name = t.Name()
	}
	return
}

// getFieldTag return tag of field
func getFieldTag(f reflect.StructField) string {
	tag := f.Tag.Get(hessianTag)
	if tag == "" {
		tag = f.Name
	}
	return tag
}

// encodeObject encode object
func encodeObject(v Any) (_ []byte, err error) {
	valueV := reflect.ValueOf(v)
	typeV := reflect.TypeOf(v)
	log.Debug("v => ", v)

	var b bytes.Buffer

	_, exist := typeV.FieldByName(ObjectType)
	if !exist {
		b.Reset()
		err = errors.New("Object Type not Set")
		return b.Bytes(), err
	}
	objectTypeField := valueV.FieldByName(ObjectType)
	if objectTypeField.Type().String() != "string" {
		b.Reset()
		err = errors.New("type of Type Field is not String")
		return b.Bytes(), err
	}
	// Object Type
	b.WriteByte('C')
	b.WriteByte(byte(len(objectTypeField.String())))
	b.WriteString(objectTypeField.String())

	// Object Field Length
	if lenField, err := PackInt16(0x90 + int16(typeV.NumField()) - 1); err != nil { // -1 是为了排除 Type Field
		b.Reset()
		err = errors.New("can not count field length, error: " + err.Error())
		return b.Bytes(), err
	} else {
		log.Debug("lenField =>", lenField[1:])
		b.Write(lenField[1:])
	}

	// Every Field Name
	for i := 0; i < typeV.NumField(); i++ {
		if typeV.Field(i).Name == ObjectType {
			continue
		}
		b.WriteByte(byte(len(typeV.Field(i).Name)))
		b.WriteString(typeV.Field(i).Name)
	}

	b.WriteByte('`')
	// Object Value
	for i := 0; i < typeV.NumField(); i++ {
		if typeV.Field(i).Name == ObjectType {
			continue
		}
		b.WriteByte(byte(len(valueV.Field(i).String())))
		if value, err := encodeByType(valueV.Field(i), typeV.Field(i).Type.Name()); err != nil { // TODO Encode 无法识别复杂类型
			b.Reset()
			err = errors.New("encode field value failed, error: " + err.Error())
			return b.Bytes(), err
		} else {
			b.Write(value)
		}
	}

	return b.Bytes(), nil
}

func encodeByType(v reflect.Value, t string) ([]byte, error) {
	switch t {
	case "string":
		// return []byte(v.String()), nil
		return Encode(v.String())
	case "int":
		return Encode(v.Int())
	case "int32":
		return Encode(int32(v.Int()))
	case "int64":
		return Encode(v.Int())
	case "bool":
		return Encode(v.Bool())
	case "[]byte":
		return Encode(v.Bytes())
	case "float64":
		return Encode(v.Float())
	case "nil":
		return Encode(nil)
	default:
		return nil, errors.New("unsupport type in object...")
	}
}
