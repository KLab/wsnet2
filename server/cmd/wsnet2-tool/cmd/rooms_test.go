package cmd

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"wsnet2/binary"
)

func TestParsePropsSimple(t *testing.T) {
	data := binary.MarshalDict(binary.Dict{
		"k1": binary.MarshalNull(),
		"k2": binary.MarshalBool(true),
		"k3": binary.MarshalULong(42),
		"k4": binary.MarshalStr8("hoge"),
		"k5": binary.MarshalBools([]bool{true, false, false, true, true}),
		"k6": binary.MarshalULongs([]uint64{1000, 2000}),
		"k7": binary.MarshalFloats([]float32{1, 1.41, 1.73}),
		"k8": binary.MarshalStrings([]string{"a", "b", "c"}),
	})
	exp := map[string]any{
		"k1": nil,
		"k2": true,
		"k3": float64(42),
		"k4": "hoge",
		"k5": "Bools[5]",
		"k6": []any{float64(1000), float64(2000)},
		"k7": []any{float64(1), float64(1.41), float64(1.73)},
		"k8": "List[3]",
	}

	str, err := parsePropsSimple(data)
	if err != nil {
		t.Fatalf("parsePropsSimple: %v - %v", err, str)
	}

	var dec map[string]any
	err = json.Unmarshal([]byte(str), &dec)
	if err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if diff := cmp.Diff(dec, exp); diff != "" {
		t.Fatalf("(-got +want)\n%s", diff)
	}
}

func TestAppendPrimitiveArraySimple(t *testing.T) {
	tests := map[string]struct {
		data []byte
		fnc  func(out, data []byte) ([]byte, error)
		exp  string
	}{
		"bools0": {
			data: binary.MarshalBools([]bool{}),
			fnc:  appendPrimitiveArraySimple[bool],
			exp:  "[],",
		},
		"bools4": {
			data: binary.MarshalBools([]bool{true, false, false, true}),
			fnc:  appendPrimitiveArraySimple[bool],
			exp:  "[true,false,false,true],",
		},
		"bools5": {
			data: binary.MarshalBools([]bool{true, false, false, true, false}),
			fnc:  appendPrimitiveArraySimple[bool],
			exp:  `"Bools[5]",`,
		},
		"ints4": {
			data: binary.MarshalInts([]int64{1, 2, 3, 4}),
			fnc:  appendPrimitiveArraySimple[int64],
			exp:  "[1,2,3,4],",
		},
		"bytes5": {
			data: binary.MarshalBytes([]int{1, 2, 3, 4, 5}),
			fnc:  appendPrimitiveArraySimple[int],
			exp:  `"Bytes[5]",`,
		},
		"floats4": {
			data: binary.MarshalDoubles([]float64{2.71, 3.14}),
			fnc:  appendPrimitiveArraySimple[float64],
			exp:  "[2.71,3.14],",
		},
	}

	for k, test := range tests {
		out, err := test.fnc(nil, test.data)
		if err != nil {
			t.Fatalf("%s: %v\n", k, err)
		}
		r := string(out)
		if diff := cmp.Diff(r, test.exp); diff != "" {
			t.Fatalf("%s: (-got +want)\n%s", k, diff)
		}
	}
}
