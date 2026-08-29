package msgpack

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

const canonicalGoldenHex = "82a6656e7469747986a26964f9a46e616d65ac4dc3b8c3b8736520f09f9a80a6616374697665c3a573636f7265cb4029000000000000a86e756c6c61626c65c0a47461677392a5616c706861a2ceb2a47061676502"

type goldenEntity struct {
	ID       int64    `json:"id"`
	Name     string   `json:"name"`
	Active   bool     `json:"active"`
	Score    float64  `json:"score"`
	Nullable any      `json:"nullable"`
	Tags     []string `json:"tags"`
}

type goldenEnvelope struct {
	Entity goldenEntity `json:"entity"`
	Page   int64        `json:"page"`
}

func canonicalFixture() goldenEnvelope {
	return goldenEnvelope{
		Entity: goldenEntity{
			ID:       -7,
			Name:     "Møøse 🚀",
			Active:   true,
			Score:    12.5,
			Nullable: nil,
			Tags:     []string{"alpha", "β"},
		},
		Page: 2,
	}
}

func TestCanonicalGoldenVector(t *testing.T) {
	want, err := hex.DecodeString(canonicalGoldenHex)
	if err != nil {
		t.Fatal(err)
	}

	got, err := Marshal(canonicalFixture())
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("golden mismatch\n got: %x\nwant: %x", got, want)
	}

	var decoded map[string]any
	if err := Unmarshal(want, &decoded); err != nil {
		t.Fatalf("Unmarshal golden: %v", err)
	}
	entity, ok := decoded["entity"].(map[string]any)
	if !ok {
		t.Fatalf("entity decoded as %T", decoded["entity"])
	}
	if entity["id"] != int64(-7) || entity["name"] != "Møøse 🚀" || entity["score"] != 12.5 {
		t.Fatalf("unexpected decoded golden: %#v", decoded)
	}
}

func TestReferenceImplementationVector(t *testing.T) {
	// {"a": 1, "b": [true, nil]} using only MessagePack's standard
	// fixed-map, fixed-string, fixed-array, boolean, nil, and integer forms.
	raw, err := hex.DecodeString("82a16101a16292c3c0")
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	want := map[string]any{"a": int64(1), "b": []any{true, nil}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestStructJSONCompatibility(t *testing.T) {
	type tagged struct {
		ID      int    `json:"id"`
		Ignored string `json:"-"`
		Empty   string `json:"empty,omitempty"`
		Bytes   []byte `json:"bytes"`
	}

	raw, err := Marshal(tagged{ID: 42, Ignored: "secret", Bytes: []byte{0, 1, 2}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got map[string]any
	if err := Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(got, map[string]any{"id": int64(42), "bytes": "AAEC"}) {
		t.Fatalf("unexpected tagged struct: %#v", got)
	}

	var decoded tagged
	if err := Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("typed Unmarshal: %v", err)
	}
	if decoded.ID != 42 || !bytes.Equal(decoded.Bytes, []byte{0, 1, 2}) {
		t.Fatalf("unexpected typed result: %#v", decoded)
	}
}

type customJSONValue struct{}

func (customJSONValue) MarshalJSON() ([]byte, error) {
	return []byte(`{"renamed":7,"raw":[null,true]}`), nil
}

func TestJSONMarshalerCompatibility(t *testing.T) {
	raw, err := Marshal(customJSONValue{})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got map[string]any
	if err := Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	want := map[string]any{"renamed": int64(7), "raw": []any{nil, true}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestContractCustomUnmarshalCompatibility(t *testing.T) {
	const largeID = int64(1<<53 + 1)
	value := map[string]any{
		"records": []any{map[string]any{
			"id": largeID,
			"nested": map[string]any{
				"revision": largeID + 2,
			},
		}},
		"nextPage": "cursor-2",
	}
	raw, err := Marshal(value)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var page contract.PageResult
	if err := Unmarshal(raw, &page); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if page.NextCursor != "cursor-2" || len(page.Items) != 1 || page.Items[0]["id"] != largeID {
		t.Fatalf("unexpected page result: %#v", page)
	}
	nested, ok := page.Items[0]["nested"].(map[string]any)
	if !ok || nested["revision"] != largeID+2 {
		t.Fatalf("nested signed integer was not preserved: %#v", page)
	}
}

func TestConcatenatedDecoder(t *testing.T) {
	first, err := Marshal(map[string]any{"id": int64(1)})
	if err != nil {
		t.Fatal(err)
	}
	flush, err := Marshal(nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Marshal(map[string]any{"id": int64(2)})
	if err != nil {
		t.Fatal(err)
	}

	d := NewDecoder(bytes.NewReader(bytes.Join([][]byte{first, flush, second}, nil)))
	var one map[string]any
	if err := d.Decode(&one); err != nil || one["id"] != int64(1) {
		t.Fatalf("first value %#v, err=%v", one, err)
	}
	var nilValue any = "not nil"
	if err := d.Decode(&nilValue); err != nil || nilValue != nil {
		t.Fatalf("nil value %#v, err=%v", nilValue, err)
	}
	var two map[string]any
	if err := d.Decode(&two); err != nil || two["id"] != int64(2) {
		t.Fatalf("second value %#v, err=%v", two, err)
	}
	if err := d.Decode(&two); !errors.Is(err, io.EOF) {
		t.Fatalf("final Decode error = %v, want io.EOF", err)
	}
}

func TestRejectsValuesOutsideProfile(t *testing.T) {
	tests := map[string][]byte{
		"binary":         {0xc4, 0x00},
		"extension":      {0xd4, 0x00, 0x00},
		"non-string key": {0x81, 0x01, 0xc0},
		"uint overflow":  {0xcf, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		"invalid utf8":   {0xd9, 0x01, 0xff},
		"nan":            {0xcb, 0x7f, 0xf8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			var value any
			if err := Unmarshal(raw, &value); err == nil {
				t.Fatal("expected rejection")
			}
		})
	}

	if _, err := Marshal(math.Inf(1)); err == nil {
		t.Fatal("expected non-finite encoder rejection")
	}
	if _, err := Marshal(map[int]string{1: "one"}); err == nil {
		t.Fatal("expected non-string map-key rejection")
	}
}

func TestDecodeLimitsAndFraming(t *testing.T) {
	tooLongString := []byte{0xdb, 0x01, 0x00, 0x00, 0x01}
	var value any
	if err := Unmarshal(tooLongString, &value); err == nil || !strings.Contains(err.Error(), "string exceeds") {
		t.Fatalf("string limit error = %v", err)
	}

	tooLargeArray := []byte{0xdd, 0x00, 0x0f, 0x42, 0x41}
	if err := Unmarshal(tooLargeArray, &value); err == nil || !strings.Contains(err.Error(), "array exceeds") {
		t.Fatalf("container limit error = %v", err)
	}

	tooDeep := append(bytes.Repeat([]byte{0x91}, MaxDepth+1), 0xc0)
	if err := Unmarshal(tooDeep, &value); err == nil || !strings.Contains(err.Error(), "depth limit") {
		t.Fatalf("depth error = %v", err)
	}

	if err := Unmarshal([]byte{0xc0, 0xc0}, &value); err == nil || !strings.Contains(err.Error(), "trailing") {
		t.Fatalf("trailing data error = %v", err)
	}
	if err := Unmarshal([]byte{0xd9, 0x02, 'x'}, &value); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("truncation error = %v, want io.ErrUnexpectedEOF", err)
	}
}

func representativeFixture() []map[string]any {
	rows := make([]map[string]any, 250)
	for i := range rows {
		rows[i] = map[string]any{
			"id":       int64(i + 1),
			"active":   i%3 != 0,
			"name":     "representative entity name",
			"score":    float64(i) * 1.25,
			"nullable": nil,
			"tags":     []any{"alpha", "β", "longer-repeated-value"},
			"nested": map[string]any{
				"region": "us-west",
				"rank":   int64(i % 10),
			},
		}
	}
	return rows
}

func TestRepresentativeEncodedSize(t *testing.T) {
	fixture := representativeFixture()
	jsonBody, err := json.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	msgpackBody, err := Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgpackBody) >= len(jsonBody) {
		t.Fatalf("MessagePack size %d is not smaller than JSON size %d", len(msgpackBody), len(jsonBody))
	}
	t.Logf("representative payload: JSON=%d bytes, MessagePack=%d bytes (%.1f%% smaller)", len(jsonBody), len(msgpackBody), 100*(1-float64(len(msgpackBody))/float64(len(jsonBody))))
}

var benchmarkSink any

func BenchmarkEncode(b *testing.B) {
	fixture := representativeFixture()
	jsonBody, _ := json.Marshal(fixture)
	msgpackBody, _ := Marshal(fixture)

	b.Run("JSON", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(jsonBody)))
		for i := 0; i < b.N; i++ {
			body, err := json.Marshal(fixture)
			if err != nil {
				b.Fatal(err)
			}
			benchmarkSink = body
		}
	})
	b.Run("MessagePack", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(msgpackBody)))
		for i := 0; i < b.N; i++ {
			body, err := Marshal(fixture)
			if err != nil {
				b.Fatal(err)
			}
			benchmarkSink = body
		}
	})
}

func BenchmarkDecode(b *testing.B) {
	fixture := representativeFixture()
	jsonBody, _ := json.Marshal(fixture)
	msgpackBody, _ := Marshal(fixture)

	b.Run("JSON", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(jsonBody)))
		for i := 0; i < b.N; i++ {
			var value any
			if err := json.Unmarshal(jsonBody, &value); err != nil {
				b.Fatal(err)
			}
			benchmarkSink = value
		}
	})
	b.Run("MessagePack", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(msgpackBody)))
		for i := 0; i < b.N; i++ {
			var value any
			if err := Unmarshal(msgpackBody, &value); err != nil {
				b.Fatal(err)
			}
			benchmarkSink = value
		}
	})
}
