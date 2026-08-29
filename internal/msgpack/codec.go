// Package msgpack implements the bounded MessagePack profile used by entity
// routes. The profile intentionally mirrors JSON values: nil, booleans, signed
// integers, finite float64 values, UTF-8 strings, arrays, and string-keyed maps.
package msgpack

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	// MediaType is the Content-Type used for MessagePack entity requests.
	MediaType = "application/vnd.msgpack"

	// The limits below bound allocations and recursive work for untrusted data.
	MaxDepth           = 128
	MaxContainerLength = 1_000_000
	MaxStringLength    = 16 * 1024 * 1024
	MaxMessageLength   = 64 * 1024 * 1024
	MaxNodeCount       = 2_000_000
)

var (
	jsonMarshalerType   = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	jsonUnmarshalerType = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
)

// Marshal encodes v using the Onyx JSON-compatible MessagePack profile.
func Marshal(v any) ([]byte, error) {
	e := encoder{buf: make([]byte, 0, 256)}
	if err := e.write(reflect.ValueOf(v), 0); err != nil {
		return nil, err
	}
	return e.buf, nil
}

// Unmarshal decodes exactly one MessagePack value into dst. Trailing values are
// rejected; use Decoder for a stream of concatenated values.
func Unmarshal(data []byte, dst any) error {
	if len(data) > MaxMessageLength {
		return fmt.Errorf("msgpack: message exceeds %d-byte limit", MaxMessageLength)
	}
	r := bytes.NewReader(data)
	d := NewDecoder(r)
	value, err := d.decodeValue()
	if err != nil {
		return err
	}
	if r.Len() != 0 {
		return errors.New("msgpack: trailing data after root value")
	}
	return assignDestination(dst, value)
}

type encoder struct {
	buf   []byte
	nodes int
}

func (e *encoder) appendByte(b byte) error {
	if len(e.buf) == MaxMessageLength {
		return fmt.Errorf("msgpack: message exceeds %d-byte limit", MaxMessageLength)
	}
	e.buf = append(e.buf, b)
	return nil
}

func (e *encoder) appendBytes(p []byte) error {
	if len(p) > MaxMessageLength-len(e.buf) {
		return fmt.Errorf("msgpack: message exceeds %d-byte limit", MaxMessageLength)
	}
	e.buf = append(e.buf, p...)
	return nil
}

func (e *encoder) write(v reflect.Value, depth int) error {
	if depth > MaxDepth {
		return fmt.Errorf("msgpack: nesting exceeds depth limit %d", MaxDepth)
	}
	e.nodes++
	if e.nodes > MaxNodeCount {
		return fmt.Errorf("msgpack: value exceeds %d-node limit", MaxNodeCount)
	}
	if !v.IsValid() {
		return e.appendByte(0xc0)
	}

	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return e.appendByte(0xc0)
		}
		return e.write(v.Elem(), depth)
	}

	if v.Kind() == reflect.Pointer && v.IsNil() {
		return e.appendByte(0xc0)
	}

	if v.CanInterface() {
		if n, ok := v.Interface().(json.Number); ok {
			return e.writeJSONNumber(n)
		}
		if v.Type().Implements(jsonMarshalerType) {
			return e.writeJSONMarshaler(v.Interface().(json.Marshaler), depth)
		}
	}
	if v.Kind() != reflect.Pointer && v.CanAddr() && v.Addr().CanInterface() && v.Addr().Type().Implements(jsonMarshalerType) {
		return e.writeJSONMarshaler(v.Addr().Interface().(json.Marshaler), depth)
	}
	if v.CanInterface() {
		if handled, err := e.writeCommon(v.Interface(), depth); handled {
			return err
		}
	}

	if v.Kind() == reflect.Pointer {
		return e.write(v.Elem(), depth)
	}

	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return e.appendByte(0xc3)
		}
		return e.appendByte(0xc2)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return e.writeInt(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n := v.Uint()
		if n > math.MaxInt64 {
			return fmt.Errorf("msgpack: unsigned integer %d exceeds signed profile", n)
		}
		return e.writeInt(int64(n))
	case reflect.Float32, reflect.Float64:
		return e.writeFloat(v.Float())
	case reflect.String:
		return e.writeString(v.String())
	case reflect.Slice:
		if v.IsNil() {
			return e.appendByte(0xc0)
		}
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// encoding/json represents byte slices as base64 strings. The binary
			// MessagePack family is deliberately outside this JSON-like profile.
			return e.writeString(base64.StdEncoding.EncodeToString(v.Bytes()))
		}
		return e.writeArray(v, depth)
	case reflect.Array:
		return e.writeArray(v, depth)
	case reflect.Map:
		if v.IsNil() {
			return e.appendByte(0xc0)
		}
		return e.writeMap(v, depth)
	case reflect.Struct:
		return e.writeStruct(v, depth)
	default:
		return fmt.Errorf("msgpack: unsupported Go type %s", v.Type())
	}
}

// writeCommon avoids reflection's map-key and element wrappers for the dynamic
// JSON-shaped values most frequently used by entity routes.
func (e *encoder) writeCommon(value any, depth int) (bool, error) {
	switch v := value.(type) {
	case bool:
		if v {
			return true, e.appendByte(0xc3)
		}
		return true, e.appendByte(0xc2)
	case string:
		return true, e.writeString(v)
	case int:
		return true, e.writeInt(int64(v))
	case int8:
		return true, e.writeInt(int64(v))
	case int16:
		return true, e.writeInt(int64(v))
	case int32:
		return true, e.writeInt(int64(v))
	case int64:
		return true, e.writeInt(v)
	case uint:
		return true, e.writeUint(uint64(v))
	case uint8:
		return true, e.writeUint(uint64(v))
	case uint16:
		return true, e.writeUint(uint64(v))
	case uint32:
		return true, e.writeUint(uint64(v))
	case uint64:
		return true, e.writeUint(v)
	case float32:
		return true, e.writeFloat(float64(v))
	case float64:
		return true, e.writeFloat(v)
	case []any:
		return true, e.writeAnyArray(v, depth)
	case []map[string]any:
		return true, e.writeMapArray(v, depth)
	case []string:
		return true, e.writeStringArray(v, depth)
	case map[string]any:
		return true, e.writeAnyMap(v, depth)
	case map[string]string:
		return true, e.writeStringMap(v, depth)
	default:
		return false, nil
	}
}

func (e *encoder) writeUint(n uint64) error {
	if n > math.MaxInt64 {
		return fmt.Errorf("msgpack: unsigned integer %d exceeds signed profile", n)
	}
	return e.writeInt(int64(n))
}

func (e *encoder) writeAnyArray(values []any, depth int) error {
	if len(values) > MaxContainerLength {
		return fmt.Errorf("msgpack: array exceeds %d-item limit", MaxContainerLength)
	}
	if err := e.writeArrayHeader(len(values)); err != nil {
		return err
	}
	for i, value := range values {
		if err := e.write(reflect.ValueOf(value), depth+1); err != nil {
			return fmt.Errorf("msgpack: array index %d: %w", i, err)
		}
	}
	return nil
}

func (e *encoder) writeMapArray(values []map[string]any, depth int) error {
	if len(values) > MaxContainerLength {
		return fmt.Errorf("msgpack: array exceeds %d-item limit", MaxContainerLength)
	}
	if err := e.writeArrayHeader(len(values)); err != nil {
		return err
	}
	for i, value := range values {
		if err := e.write(reflect.ValueOf(value), depth+1); err != nil {
			return fmt.Errorf("msgpack: array index %d: %w", i, err)
		}
	}
	return nil
}

func (e *encoder) writeStringArray(values []string, depth int) error {
	if len(values) > MaxContainerLength {
		return fmt.Errorf("msgpack: array exceeds %d-item limit", MaxContainerLength)
	}
	if err := e.writeArrayHeader(len(values)); err != nil {
		return err
	}
	for i, value := range values {
		if err := e.write(reflect.ValueOf(value), depth+1); err != nil {
			return fmt.Errorf("msgpack: array index %d: %w", i, err)
		}
	}
	return nil
}

func (e *encoder) writeAnyMap(values map[string]any, depth int) error {
	if len(values) > MaxContainerLength {
		return fmt.Errorf("msgpack: map exceeds %d-entry limit", MaxContainerLength)
	}
	if err := e.writeMapHeader(len(values)); err != nil {
		return err
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := e.enterMapKey(depth + 1); err != nil {
			return err
		}
		if err := e.writeString(key); err != nil {
			return err
		}
		if err := e.write(reflect.ValueOf(values[key]), depth+1); err != nil {
			return fmt.Errorf("msgpack: map value %q: %w", key, err)
		}
	}
	return nil
}

func (e *encoder) writeStringMap(values map[string]string, depth int) error {
	if len(values) > MaxContainerLength {
		return fmt.Errorf("msgpack: map exceeds %d-entry limit", MaxContainerLength)
	}
	if err := e.writeMapHeader(len(values)); err != nil {
		return err
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := e.enterMapKey(depth + 1); err != nil {
			return err
		}
		if err := e.writeString(key); err != nil {
			return err
		}
		if err := e.write(reflect.ValueOf(values[key]), depth+1); err != nil {
			return fmt.Errorf("msgpack: map value %q: %w", key, err)
		}
	}
	return nil
}

func (e *encoder) writeJSONMarshaler(m json.Marshaler, depth int) error {
	raw, err := m.MarshalJSON()
	if err != nil {
		return fmt.Errorf("msgpack: marshal JSON compatibility value: %w", err)
	}
	value, err := decodeJSONValue(raw)
	if err != nil {
		return fmt.Errorf("msgpack: invalid MarshalJSON result: %w", err)
	}
	// Replace the original Go value with its JSON representation rather than
	// counting both as separate wire nodes.
	e.nodes--
	return e.write(reflect.ValueOf(value), depth)
}

func (e *encoder) writeJSONNumber(n json.Number) error {
	s := n.String()
	if !strings.ContainsAny(s, ".eE") {
		if i, err := n.Int64(); err == nil {
			return e.writeInt(i)
		}
	}
	f, err := n.Float64()
	if err != nil {
		return fmt.Errorf("msgpack: invalid JSON number %q", s)
	}
	return e.writeFloat(f)
}

func (e *encoder) writeInt(n int64) error {
	switch {
	case n >= 0 && n <= 0x7f:
		return e.appendByte(byte(n))
	case n >= 0 && n <= math.MaxUint8:
		return e.appendFixed(0xcc, uint64(n), 1)
	case n >= 0 && n <= math.MaxUint16:
		return e.appendFixed(0xcd, uint64(n), 2)
	case n >= 0 && n <= math.MaxUint32:
		return e.appendFixed(0xce, uint64(n), 4)
	case n >= 0:
		return e.appendFixed(0xcf, uint64(n), 8)
	case n >= -32:
		return e.appendByte(byte(int8(n)))
	case n >= math.MinInt8:
		return e.appendFixed(0xd0, uint64(uint8(int8(n))), 1)
	case n >= math.MinInt16:
		return e.appendFixed(0xd1, uint64(uint16(int16(n))), 2)
	case n >= math.MinInt32:
		return e.appendFixed(0xd2, uint64(uint32(int32(n))), 4)
	default:
		return e.appendFixed(0xd3, uint64(n), 8)
	}
}

func (e *encoder) writeFloat(f float64) error {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return fmt.Errorf("msgpack: non-finite float %v is outside the JSON profile", f)
	}
	return e.appendFixed(0xcb, math.Float64bits(f), 8)
}

func (e *encoder) appendFixed(prefix byte, n uint64, width int) error {
	var buf [9]byte
	buf[0] = prefix
	switch width {
	case 1:
		buf[1] = byte(n)
	case 2:
		binary.BigEndian.PutUint16(buf[1:3], uint16(n))
	case 4:
		binary.BigEndian.PutUint32(buf[1:5], uint32(n))
	case 8:
		binary.BigEndian.PutUint64(buf[1:9], n)
	default:
		panic("unsupported fixed-width MessagePack value")
	}
	return e.appendBytes(buf[:width+1])
}

func (e *encoder) writeString(s string) error {
	if !utf8.ValidString(s) {
		return errors.New("msgpack: string is not valid UTF-8")
	}
	if len(s) > MaxStringLength {
		return fmt.Errorf("msgpack: string exceeds %d-byte limit", MaxStringLength)
	}
	if err := e.writeStringHeader(len(s)); err != nil {
		return err
	}
	return e.appendBytes([]byte(s))
}

func (e *encoder) writeStringHeader(n int) error {
	switch {
	case n <= 31:
		return e.appendByte(0xa0 | byte(n))
	case n <= math.MaxUint8:
		return e.appendFixed(0xd9, uint64(n), 1)
	case n <= math.MaxUint16:
		return e.appendFixed(0xda, uint64(n), 2)
	default:
		return e.appendFixed(0xdb, uint64(n), 4)
	}
}

func (e *encoder) writeArray(v reflect.Value, depth int) error {
	n := v.Len()
	if n > MaxContainerLength {
		return fmt.Errorf("msgpack: array exceeds %d-item limit", MaxContainerLength)
	}
	if err := e.writeArrayHeader(n); err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		if err := e.write(v.Index(i), depth+1); err != nil {
			return fmt.Errorf("msgpack: array index %d: %w", i, err)
		}
	}
	return nil
}

func (e *encoder) writeArrayHeader(n int) error {
	switch {
	case n <= 15:
		return e.appendByte(0x90 | byte(n))
	case n <= math.MaxUint16:
		return e.appendFixed(0xdc, uint64(n), 2)
	default:
		return e.appendFixed(0xdd, uint64(n), 4)
	}
}

func (e *encoder) writeMap(v reflect.Value, depth int) error {
	if v.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("msgpack: map key type %s is not a string", v.Type().Key())
	}
	n := v.Len()
	if n > MaxContainerLength {
		return fmt.Errorf("msgpack: map exceeds %d-entry limit", MaxContainerLength)
	}
	if err := e.writeMapHeader(n); err != nil {
		return err
	}

	keys := v.MapKeys()
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
	for _, key := range keys {
		if err := e.enterMapKey(depth + 1); err != nil {
			return err
		}
		if err := e.writeString(key.String()); err != nil {
			return err
		}
		if err := e.write(v.MapIndex(key), depth+1); err != nil {
			return fmt.Errorf("msgpack: map value %q: %w", key.String(), err)
		}
	}
	return nil
}

func (e *encoder) writeMapHeader(n int) error {
	switch {
	case n <= 15:
		return e.appendByte(0x80 | byte(n))
	case n <= math.MaxUint16:
		return e.appendFixed(0xde, uint64(n), 2)
	default:
		return e.appendFixed(0xdf, uint64(n), 4)
	}
}

func (e *encoder) writeStruct(v reflect.Value, depth int) error {
	t := v.Type()
	fields := make([]reflect.Value, 0, t.NumField())
	names := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		tag := field.Tag.Get("json")
		parts := strings.Split(tag, ",")
		if parts[0] == "-" {
			continue
		}
		// Embedded-field conflict resolution and the ,string option are best
		// delegated to encoding/json to preserve the SDK's existing behavior.
		if field.Anonymous || hasTagOption(parts[1:], "string") {
			raw, err := json.Marshal(v.Interface())
			if err != nil {
				return fmt.Errorf("msgpack: JSON-normalize struct %s: %w", t, err)
			}
			value, err := decodeJSONValue(raw)
			if err != nil {
				return err
			}
			e.nodes--
			return e.write(reflect.ValueOf(value), depth)
		}
		name := parts[0]
		if name == "" {
			name = field.Name
		}
		fv := v.Field(i)
		if hasTagOption(parts[1:], "omitempty") && isEmptyJSONValue(fv) {
			continue
		}
		names = append(names, name)
		fields = append(fields, fv)
	}
	if len(fields) > MaxContainerLength {
		return fmt.Errorf("msgpack: struct exceeds %d-field limit", MaxContainerLength)
	}
	if err := e.writeMapHeader(len(fields)); err != nil {
		return err
	}
	for i, name := range names {
		if err := e.enterMapKey(depth + 1); err != nil {
			return err
		}
		if err := e.writeString(name); err != nil {
			return err
		}
		if err := e.write(fields[i], depth+1); err != nil {
			return fmt.Errorf("msgpack: struct field %q: %w", name, err)
		}
	}
	return nil
}

func (e *encoder) enterMapKey(depth int) error {
	if depth > MaxDepth {
		return fmt.Errorf("msgpack: nesting exceeds depth limit %d", MaxDepth)
	}
	e.nodes++
	if e.nodes > MaxNodeCount {
		return fmt.Errorf("msgpack: value exceeds %d-node limit", MaxNodeCount)
	}
	return nil
}

func hasTagOption(options []string, target string) bool {
	for _, option := range options {
		if option == target {
			return true
		}
	}
	return false
}

func isEmptyJSONValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}
	return false
}

func decodeJSONValue(raw []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, errors.New("multiple JSON values")
		}
		return nil, err
	}
	return value, nil
}
