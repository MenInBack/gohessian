// Decode hessian data
package gohessian

import (
	"bufio"
	"fmt"
	"io"
	"time"

	log "github.com/cihub/seelog"
)

const (
	PARSE_DEBUG = true
)

func NewHessian(r io.Reader) (h *Hessian) {
	return &Hessian{reader: bufio.NewReader(r)}
}

// peekByte read the byte and do not move the point
func (h *Hessian) peekByte() (b byte) {
	b = h.peek(1)[0]
	return
}

// appendRefs append reference
func (h *Hessian) appendRefs(v interface{}) {
	h.refs = append(h.refs, v)
}

// len return length of buffer
func (h *Hessian) len() (l int) {
	h.peek(1) // read the resources in order to get the length of buffer
	l = h.reader.Buffered()
	return
}

// readByte read a byte in hessian struct and move back a byte
func (h *Hessian) readByte() (c byte, err error) {
	c, err = h.reader.ReadByte()
	return
}

// next read the bytes of the specified length and move back N bytes
func (h *Hessian) next(n int) (b []byte) {
	if n <= 0 {
		return
	}
	if n >= h.reader.Buffered() {
		n = h.reader.Buffered()
	}
	b = make([]byte, n)
	h.reader.Read(b)
	return
}

// peek read the bytes of the specified length and do not move the point
func (h *Hessian) peek(n int) (b []byte) {
	b, _ = h.reader.Peek(n)
	return
}

// nextRune reads the utf8 character of the specified length
func (h *Hessian) nextRune(n int) (s []rune) {
	for i := 0; i < n; i++ {
		if r, ri, e := h.reader.ReadRune(); e == nil && ri > 0 {
			s = append(s, r)
		}
	}
	return
}

// readType read the type of data for list and map
func (h *Hessian) readType() string {
	if h.peekByte() != 't' {
		return ""
	}
	tLen, _ := UnpackInt16(h.peek(3)[1:3]) // take the length of type name
	tName := h.nextRune(int(3 + tLen))[3:] // take the type name
	return string(tName)
}

// Parse hessian data
func (h *Hessian) Parse() (v interface{}, err error) {
	t, err := h.readByte()
	if err == io.EOF {
		return
	}
	switch t {
	case 'r': // reply
		h.next(2)
		return h.Parse()

	case 'f': // fault
		h.Parse() // drop "code"
		code, _ := h.Parse()
		h.Parse() // drop "message"
		message, _ := h.Parse()
		v = nil
		err = log.Errorf("%s : %s", code, message)
	case 'N': // null
		v = nil

	case 'T': // true
		v = true

	case 'F': // false
		v = false

	case 'I': // int
		if v, err = UnpackInt32(h.next(4)); err != nil {
			return nil, err
		}

	case 'L': // long
		if v, err = UnpackInt64(h.next(8)); err != nil {
			v = nil
			return
		}

	case 'D': // double
		if v, err = UnpackFloat64(h.next(8)); err != nil {
			v = nil
			return
		}

	case 'd': // date
		var ms int64
		if ms, err = UnpackInt64(h.next(8)); err != nil {
			v = nil
			return
		}
		v = time.Unix(ms/1000, ms%1000*10E5)

	case 'S', 's', 'X', 'x': // string, xml
		var strChunks []rune
		var l int16
		for { // avoid recursive readings Chunks
			if l, err = UnpackInt16(h.next(2)); err != nil {
				strChunks = nil
				return
			}
			strChunks = append(strChunks, h.nextRune(int(l))...)
			if t == 'S' || t == 'X' {
				break
			}
			if t, err = h.readByte(); err != nil {
				strChunks = nil
				return
			}
		}
		v = string(strChunks)

	case 'B', 'b': // binary
		var bChunks []byte // Equivalent to []uint8
		var l int16
		for { // avoid recursive readings Chunks
			if l, err = UnpackInt16(h.next(2)); err != nil {
				bChunks = nil
				return
			}
			bChunks = append(bChunks, h.next(int(l))...)
			if t == 'B' {
				break
			}
			if t, err = h.readByte(); err != nil {
				bChunks = nil
				return
			}
		}
		v = bChunks

	case 'V': // list
		h.readType()
		var listChunks []interface{}
		if h.peekByte() == 'l' {
			h.next(5)
		}
		for h.peekByte() != 'z' {
			if _v, _e := h.Parse(); _e != nil {
				listChunks = nil
				return
			} else {
				listChunks = append(listChunks, _v)
			}
		}
		h.readByte()
		v = listChunks
		h.appendRefs(&listChunks)

	case 'M': // map
		h.readType()
		var mapChunks = make(map[interface{}]interface{})
		for h.peekByte() != 'z' {
			_kv, _ke := h.Parse()
			if _ke != nil {
				mapChunks = nil
				err = _ke
				return
			}
			_vv, _ve := h.Parse()
			if _ve != nil {
				mapChunks = nil
				err = _ve
				return
			}
			mapChunks[_kv] = _vv
		}
		h.readByte()
		v = mapChunks
		h.appendRefs(&mapChunks)

	case 'R': //ref
		var refIdx int32
		if refIdx, err = UnpackInt32(h.next(4)); err != nil {
			return
		}
		if len(h.refs) > int(refIdx) {
			v = &h.refs[refIdx]
		}

	default:
		err = fmt.Errorf("Invalid type: %v,>>%v<<<", string(t), h.peek(h.len()))
	} // switch
	return
} // Parse end
