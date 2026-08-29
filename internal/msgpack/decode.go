package msgpack

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"unicode/utf8"
)

// valueUnmarshaler lets response wrapper types consume the already validated
// MessagePack value tree without a JSON round trip. That distinction matters
// for dynamic entity values because JSON decoding into any converts integers
// to float64 and cannot exactly represent every signed int64.
type valueUnmarshaler interface {
	UnmarshalMessagePackValue(any) error
}

// Decoder decodes a sequence of concatenated MessagePack values. Each value is
// independently subject to MaxMessageLength.
type Decoder struct {
	r        io.Reader
	consumed int
	nodes    int
}

// NewDecoder returns a bounded decoder reading from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the next complete MessagePack value into dst. It returns io.EOF
// when no more values are available.
func (d *Decoder) Decode(dst any) error {
	value, err := d.decodeValue()
	if err != nil {
		return err
	}
	return assignDestination(dst, value)
}

func (d *Decoder) decodeValue() (any, error) {
	d.consumed = 0
	d.nodes = 0
	var prefix [1]byte
	n, err := io.ReadFull(d.r, prefix[:])
	if err != nil {
		if n == 0 && errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
		return nil, io.ErrUnexpectedEOF
	}
	d.consumed = 1
	return d.decodePrefix(prefix[0], 0)
}

func (d *Decoder) decodePrefix(prefix byte, depth int) (any, error) {
	if depth > MaxDepth {
		return nil, fmt.Errorf("msgpack: nesting exceeds depth limit %d", MaxDepth)
	}
	d.nodes++
	if d.nodes > MaxNodeCount {
		return nil, fmt.Errorf("msgpack: value exceeds %d-node limit", MaxNodeCount)
	}
	switch {
	case prefix <= 0x7f:
		return int64(prefix), nil
	case prefix >= 0xe0:
		return int64(int8(prefix)), nil
	case prefix >= 0xa0 && prefix <= 0xbf:
		return d.readString(int(prefix & 0x1f))
	case prefix >= 0x90 && prefix <= 0x9f:
		return d.readArray(int(prefix&0x0f), depth)
	case prefix >= 0x80 && prefix <= 0x8f:
		return d.readMap(int(prefix&0x0f), depth)
	}

	switch prefix {
	case 0xc0:
		return nil, nil
	case 0xc2:
		return false, nil
	case 0xc3:
		return true, nil
	case 0xca:
		n, err := d.readUint(4)
		if err != nil {
			return nil, err
		}
		f := float64(math.Float32frombits(uint32(n)))
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, errors.New("msgpack: non-finite float is outside the JSON profile")
		}
		return f, nil
	case 0xcb:
		n, err := d.readUint(8)
		if err != nil {
			return nil, err
		}
		f := math.Float64frombits(n)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, errors.New("msgpack: non-finite float is outside the JSON profile")
		}
		return f, nil
	case 0xcc:
		return d.readPositiveInt(1)
	case 0xcd:
		return d.readPositiveInt(2)
	case 0xce:
		return d.readPositiveInt(4)
	case 0xcf:
		return d.readPositiveInt(8)
	case 0xd0:
		n, err := d.readUint(1)
		return int64(int8(n)), err
	case 0xd1:
		n, err := d.readUint(2)
		return int64(int16(n)), err
	case 0xd2:
		n, err := d.readUint(4)
		return int64(int32(n)), err
	case 0xd3:
		n, err := d.readUint(8)
		return int64(n), err
	case 0xd9:
		n, err := d.readLength(1)
		if err != nil {
			return nil, err
		}
		return d.readString(n)
	case 0xda:
		n, err := d.readLength(2)
		if err != nil {
			return nil, err
		}
		return d.readString(n)
	case 0xdb:
		n, err := d.readLength(4)
		if err != nil {
			return nil, err
		}
		return d.readString(n)
	case 0xdc:
		n, err := d.readLength(2)
		if err != nil {
			return nil, err
		}
		return d.readArray(n, depth)
	case 0xdd:
		n, err := d.readLength(4)
		if err != nil {
			return nil, err
		}
		return d.readArray(n, depth)
	case 0xde:
		n, err := d.readLength(2)
		if err != nil {
			return nil, err
		}
		return d.readMap(n, depth)
	case 0xdf:
		n, err := d.readLength(4)
		if err != nil {
			return nil, err
		}
		return d.readMap(n, depth)
	case 0xc1:
		return nil, errors.New("msgpack: reserved type marker 0xc1")
	case 0xc4, 0xc5, 0xc6:
		return nil, errors.New("msgpack: binary values are outside the JSON profile")
	case 0xc7, 0xc8, 0xc9, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8:
		return nil, errors.New("msgpack: extension values are outside the JSON profile")
	default:
		return nil, fmt.Errorf("msgpack: unsupported type marker 0x%02x", prefix)
	}
}

func (d *Decoder) readPositiveInt(width int) (any, error) {
	n, err := d.readUint(width)
	if err != nil {
		return nil, err
	}
	if n > math.MaxInt64 {
		return nil, fmt.Errorf("msgpack: unsigned integer %d exceeds signed profile", n)
	}
	return int64(n), nil
}

func (d *Decoder) readUint(width int) (uint64, error) {
	var buf [8]byte
	if err := d.readFull(buf[:width]); err != nil {
		return 0, err
	}
	switch width {
	case 1:
		return uint64(buf[0]), nil
	case 2:
		return uint64(binary.BigEndian.Uint16(buf[:2])), nil
	case 4:
		return uint64(binary.BigEndian.Uint32(buf[:4])), nil
	case 8:
		return binary.BigEndian.Uint64(buf[:8]), nil
	default:
		panic("unsupported fixed-width MessagePack value")
	}
}

func (d *Decoder) readLength(width int) (int, error) {
	n, err := d.readUint(width)
	if err != nil {
		return 0, err
	}
	if n > uint64(math.MaxInt) {
		return 0, errors.New("msgpack: length overflows int")
	}
	return int(n), nil
}

func (d *Decoder) readString(n int) (string, error) {
	if n > MaxStringLength {
		return "", fmt.Errorf("msgpack: string exceeds %d-byte limit", MaxStringLength)
	}
	buf := make([]byte, n)
	if err := d.readFull(buf); err != nil {
		return "", err
	}
	if !utf8.Valid(buf) {
		return "", errors.New("msgpack: string is not valid UTF-8")
	}
	return string(buf), nil
}

func (d *Decoder) readArray(n, depth int) ([]any, error) {
	if n > MaxContainerLength {
		return nil, fmt.Errorf("msgpack: array exceeds %d-item limit", MaxContainerLength)
	}
	items := make([]any, n)
	for i := range items {
		prefix, err := d.readPrefix()
		if err != nil {
			return nil, fmt.Errorf("msgpack: array index %d: %w", i, err)
		}
		items[i], err = d.decodePrefix(prefix, depth+1)
		if err != nil {
			return nil, fmt.Errorf("msgpack: array index %d: %w", i, err)
		}
	}
	return items, nil
}

func (d *Decoder) readMap(n, depth int) (map[string]any, error) {
	if n > MaxContainerLength {
		return nil, fmt.Errorf("msgpack: map exceeds %d-entry limit", MaxContainerLength)
	}
	items := make(map[string]any, n)
	for i := 0; i < n; i++ {
		prefix, err := d.readPrefix()
		if err != nil {
			return nil, fmt.Errorf("msgpack: map key %d: %w", i, err)
		}
		if depth+1 > MaxDepth {
			return nil, fmt.Errorf("msgpack: nesting exceeds depth limit %d", MaxDepth)
		}
		d.nodes++
		if d.nodes > MaxNodeCount {
			return nil, fmt.Errorf("msgpack: value exceeds %d-node limit", MaxNodeCount)
		}
		key, err := d.decodeStringPrefix(prefix)
		if err != nil {
			return nil, fmt.Errorf("msgpack: map key %d: %w", i, err)
		}
		prefix, err = d.readPrefix()
		if err != nil {
			return nil, fmt.Errorf("msgpack: map value %q: %w", key, err)
		}
		value, err := d.decodePrefix(prefix, depth+1)
		if err != nil {
			return nil, fmt.Errorf("msgpack: map value %q: %w", key, err)
		}
		items[key] = value
	}
	return items, nil
}

func (d *Decoder) decodeStringPrefix(prefix byte) (string, error) {
	switch {
	case prefix >= 0xa0 && prefix <= 0xbf:
		return d.readString(int(prefix & 0x1f))
	case prefix == 0xd9:
		n, err := d.readLength(1)
		if err != nil {
			return "", err
		}
		return d.readString(n)
	case prefix == 0xda:
		n, err := d.readLength(2)
		if err != nil {
			return "", err
		}
		return d.readString(n)
	case prefix == 0xdb:
		n, err := d.readLength(4)
		if err != nil {
			return "", err
		}
		return d.readString(n)
	default:
		return "", fmt.Errorf("non-string map key marker 0x%02x", prefix)
	}
}

func (d *Decoder) readPrefix() (byte, error) {
	var buf [1]byte
	if err := d.readFull(buf[:]); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (d *Decoder) readFull(buf []byte) error {
	if len(buf) > MaxMessageLength-d.consumed {
		return fmt.Errorf("msgpack: message exceeds %d-byte limit", MaxMessageLength)
	}
	n, err := io.ReadFull(d.r, buf)
	d.consumed += n
	if err != nil {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func assignDestination(dst any, src any) error {
	if dst == nil {
		return errors.New("msgpack: destination must be a non-nil pointer")
	}
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return errors.New("msgpack: destination must be a non-nil pointer")
	}
	if v.Type().Implements(jsonUnmarshalerType) {
		return assign(v, src)
	}
	return assign(v.Elem(), src)
}

func assign(dst reflect.Value, src any) error {
	if dst.CanInterface() {
		if unmarshaler, ok := dst.Interface().(valueUnmarshaler); ok {
			return unmarshaler.UnmarshalMessagePackValue(src)
		}
	}
	if dst.CanInterface() && dst.Type().Implements(jsonUnmarshalerType) {
		raw, err := json.Marshal(src)
		if err != nil {
			return err
		}
		return dst.Interface().(json.Unmarshaler).UnmarshalJSON(raw)
	}

	if dst.Kind() == reflect.Pointer {
		if src == nil {
			dst.SetZero()
			return nil
		}
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return assign(dst.Elem(), src)
	}

	if src == nil {
		dst.SetZero()
		return nil
	}

	switch dst.Kind() {
	case reflect.Interface:
		sv := reflect.ValueOf(src)
		if !sv.Type().AssignableTo(dst.Type()) && dst.Type().NumMethod() != 0 {
			return fmt.Errorf("msgpack: cannot assign %T to %s", src, dst.Type())
		}
		dst.Set(sv)
		return nil
	case reflect.Bool:
		v, ok := src.(bool)
		if !ok {
			return typeError(src, dst.Type())
		}
		dst.SetBool(v)
		return nil
	case reflect.String:
		v, ok := src.(string)
		if !ok {
			return typeError(src, dst.Type())
		}
		dst.SetString(v)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, ok := src.(int64)
		if !ok || dst.OverflowInt(v) {
			return typeError(src, dst.Type())
		}
		dst.SetInt(v)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v, ok := src.(int64)
		if !ok || v < 0 || dst.OverflowUint(uint64(v)) {
			return typeError(src, dst.Type())
		}
		dst.SetUint(uint64(v))
		return nil
	case reflect.Float32, reflect.Float64:
		var v float64
		switch value := src.(type) {
		case float64:
			v = value
		case int64:
			v = float64(value)
		default:
			return typeError(src, dst.Type())
		}
		if dst.OverflowFloat(v) {
			return typeError(src, dst.Type())
		}
		dst.SetFloat(v)
		return nil
	case reflect.Slice:
		if dst.Type().Elem().Kind() == reflect.Uint8 {
			value, ok := src.(string)
			if !ok {
				return typeError(src, dst.Type())
			}
			decoded, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				return fmt.Errorf("msgpack: decode base64 byte slice: %w", err)
			}
			dst.SetBytes(decoded)
			return nil
		}
		values, ok := src.([]any)
		if !ok {
			return typeError(src, dst.Type())
		}
		slice := reflect.MakeSlice(dst.Type(), len(values), len(values))
		for i, value := range values {
			if err := assign(slice.Index(i), value); err != nil {
				return fmt.Errorf("msgpack: slice index %d: %w", i, err)
			}
		}
		dst.Set(slice)
		return nil
	case reflect.Array:
		values, ok := src.([]any)
		if !ok {
			return typeError(src, dst.Type())
		}
		dst.SetZero()
		limit := len(values)
		if limit > dst.Len() {
			limit = dst.Len()
		}
		for i := 0; i < limit; i++ {
			if err := assign(dst.Index(i), values[i]); err != nil {
				return fmt.Errorf("msgpack: array index %d: %w", i, err)
			}
		}
		return nil
	case reflect.Map:
		values, ok := src.(map[string]any)
		if !ok || dst.Type().Key().Kind() != reflect.String {
			return typeError(src, dst.Type())
		}
		result := reflect.MakeMapWithSize(dst.Type(), len(values))
		for key, value := range values {
			mapValue := reflect.New(dst.Type().Elem()).Elem()
			if err := assign(mapValue, value); err != nil {
				return fmt.Errorf("msgpack: map value %q: %w", key, err)
			}
			mapKey := reflect.ValueOf(key).Convert(dst.Type().Key())
			result.SetMapIndex(mapKey, mapValue)
		}
		dst.Set(result)
		return nil
	case reflect.Struct:
		// Let encoding/json apply field selection, tags, embedded-field rules,
		// and nested UnmarshalJSON methods after the bounded wire decode.
		raw, err := json.Marshal(src)
		if err != nil {
			return err
		}
		return json.Unmarshal(raw, dst.Addr().Interface())
	default:
		return fmt.Errorf("msgpack: unsupported destination type %s", dst.Type())
	}
}

func typeError(src any, dst reflect.Type) error {
	return fmt.Errorf("msgpack: cannot assign %T to %s", src, dst)
}
