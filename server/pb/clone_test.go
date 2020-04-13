package pb

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

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
		if v1.IsNil() || v2.IsNil() {
			return nil
		}
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

func testCloned(s, d interface{}) error {
	return testRef(reflect.ValueOf(s), reflect.ValueOf(d))
}

func TestCloneRoomInfo(t *testing.T) {
	s := &RoomInfo{}
	if err := testCloned(s, s.Clone()); err != nil {
		t.Fatalf("%T: %v", s, err)
	}
}

func TestCloneClientInfo(t *testing.T) {
	s := &ClientInfo{}
	if err := testCloned(s, s.Clone()); err != nil {
		t.Fatalf("%T: %v", s, err)
	}
}
