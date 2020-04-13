package pb

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestRoomInfo_Clone(t *testing.T) {
	s := &RoomInfo{}
	fill(s)
	if err := testCloned(s, s.Clone()); err != nil {
		t.Fatalf("%T: %v", s, err)
	}
}

func TestClientInfo_Clone(t *testing.T) {
	s := &ClientInfo{}
	fill(s)
	if err := testCloned(s, s.Clone()); err != nil {
		t.Fatalf("%T: %v", s, err)
	}
}

func TestTimestamp_Clone(t *testing.T) {
	s := &Timestamp{}
	fill(s)
	if err := testCloned(s, s.Clone()); err != nil {
		t.Fatalf("%T: %v", s, err)
	}
}

// testCloned : Cloneできているか判定
func testCloned(s, d interface{}) error {
	return testRef(reflect.ValueOf(s), reflect.ValueOf(d))
}

// testRef : structに含まれる slice/map/pointer が同一でないことをテスト
func testRef(v1, v2 reflect.Value) error {
	if v1.Type() != v2.Type() {
		return fmt.Errorf("type mismatch")
	}

	switch v1.Kind() {
	case reflect.Slice:
		if v1.Pointer() == v2.Pointer() {
			return fmt.Errorf("same slice")
		}
	case reflect.Map:
		if v1.Pointer() == v2.Pointer() {
			return fmt.Errorf("same map")
		}
	case reflect.Ptr:
		if v1.Pointer() == v2.Pointer() {
			return fmt.Errorf("same pointer")
		}
		return testRef(v1.Elem(), v2.Elem())
	case reflect.Struct:
		for i := 0; i < v1.NumField(); i++ {
			name := v1.Type().Field(i).Name
			if strings.HasPrefix(name, "XXX_") {
				// skip field defineded by protobuf
				continue
			}
			if err := testRef(v1.Field(i), v2.Field(i)); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
	}

	return nil
}

// fill : structを適当な値で埋める
// ptrの中身がcloneされるかを検出するためにはなにか値が入っている必要がある
func fill(p interface{}) {
	fillv(reflect.ValueOf(p))
}

func fillv(p reflect.Value) {
	v := p.Elem()
	switch v.Kind() {
	case reflect.Ptr:
		v.Set(reflect.New(v.Type().Elem()))
		fillv(v)
	case reflect.Map:
		v.Set(reflect.MakeMap(v.Type()))
	case reflect.Slice:
		v.Set(reflect.MakeSlice(v.Type(), 0, 1))
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			fillv(f.Addr())
		}
	}
}
