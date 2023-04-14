package binary

import (
	"reflect"
	"testing"
)

func TestUnmarshalNullDict(t *testing.T) {
	tests := map[string]struct {
		data []byte
		exp  Dict
	}{
		"nil": {
			data: MarshalDict(nil),
			exp:  Dict{},
		},
		"dict": {
			MarshalDict(Dict{"a": MarshalBool(true), "b": MarshalStr8("bbb")}),
			Dict{"a": MarshalBool(true), "b": MarshalStr8("bbb")},
		},
	}
	for k, tc := range tests {
		d, _, e := UnmarshalNullDict(tc.data)
		if e != nil {
			t.Fatalf("%v: %v", k, e)
		}
		if !reflect.DeepEqual(d, tc.exp) {
			t.Fatalf("unmarshaled(%v) = %#v, wants %#v", k, d, tc.exp)
		}
	}
}

func TestLeavePayload(t *testing.T) {
	tests := map[string]struct {
		msg string
		exp string
	}{
		"short": {
			"abcdあいうえ",
			"abcdあいうえ",
		},
		"long": {
			"あいうえおあいうえおあいうえおあいうえおあいうえおあいうえおあいうえおあいうえおあいうえお",
			"あいうえおあいうえおあいうえおあいうえおあいうえおあいうえおあいうえおあいうえおあ",
		},
		"truncate": {
			"aあいうえおあいうえおあいうえおあいうえおあいうえおあいうえおあいうえおあいうえおあいうえお",
			"aあいうえおあいうえおあいうえおあいうえおあいうえおあいうえおあいうえおあいうえお",
		},
	}
	for k, tc := range tests {
		p := MarshalLeavePayload(tc.msg)

		u, err := UnmarshalLeavePayload(p)
		if err != nil {
			t.Fatalf("%v: %v", k, err)
		}
		if u != tc.exp {
			t.Fatalf("%v: %v, wants %v", k, u, tc.exp)
		}
	}
}

func TestRoomPropPayload(t *testing.T) {
	const v = true
	const j = false
	const w = true
	const grp = 17
	const maxp = 13
	const cdl = 23
	pubp := Dict{"pub": MarshalBool(true)}
	prvp := Dict{"prv": MarshalStr8("ok")}

	p := MarshalRoomPropPayload(v, j, w, grp, maxp, cdl, pubp, prvp)
	u, err := UnmarshalEvRoomPropPayload(p)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if u.Visible != v {
		t.Fatalf("Visible = %v, wants %v", u.Visible, v)
	}
	if u.Joinable != j {
		t.Fatalf("Joinable = %v, wants %v", u.Joinable, j)
	}
	if u.Watchable != w {
		t.Fatalf("Watchable = %v, wants %v", u.Watchable, w)
	}
	if u.SearchGroup != grp {
		t.Fatalf("SearchGroup = %v, wants %v", u.SearchGroup, grp)
	}
	if u.MaxPlayer != maxp {
		t.Fatalf("MaxPlayer = %v, wants %v", u.MaxPlayer, maxp)
	}
	if u.ClientDeadline != cdl {
		t.Fatalf("ClientDeadline = %v, wants %v", u.ClientDeadline, cdl)
	}
	if !reflect.DeepEqual(u.PublicProps, pubp) {
		t.Fatalf("PublicProps = %#v, wants %#v", u.PublicProps, pubp)
	}
	if !reflect.DeepEqual(u.PrivateProps, prvp) {
		t.Fatalf("PrivateProps = %#v, wants %#v", u.PrivateProps, prvp)
	}

	gcdl, err := GetRoomPropClientDeadline(p)
	if err != nil {
		t.Fatalf("GetRoomPropClientDeadline: %v", err)
	}
	if gcdl != cdl {
		t.Fatalf("RoomPropClientDeadline = %v, wants %v", gcdl, cdl)
	}
}

func TestClientPropPayload(t *testing.T) {
	tests := map[string]struct {
		prop Dict
		exp  Dict
	}{
		"null": {
			prop: nil,
			exp:  Dict{},
		},
		"dic": {
			prop: Dict{"a": MarshalBool(true), "b": MarshalNull()},
			exp:  Dict{"a": MarshalBool(true), "b": MarshalNull()},
		},
	}
	for k, tc := range tests {
		p := MarshalClientPropPayload(tc.prop)
		u, err := UnmarshalClientPropPayload(p)
		if err != nil {
			t.Fatalf("%v: %v", k, err)
		}
		if !reflect.DeepEqual(u, tc.exp) {
			t.Fatalf("%v: %#v, watns %#v", k, u, tc.exp)
		}
	}
}

func TestSwitchMasterPayload(t *testing.T) {
	const newmaster = "NewMasterId"

	p := MarshalSwitchMasterPayload(newmaster)
	u, err := UnmarshalSwitchMasterPayload(p)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if u != newmaster {
		t.Fatalf("new master: %v, wants %v", u, newmaster)
	}
}
