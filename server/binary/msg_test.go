package binary

import (
	"reflect"
	"testing"
)

func TestUnmarshalNullDict(t *testing.T) {
	tests := map[string]struct {
		data   []byte
		expect Dict
	}{
		"nil": {
			data:   MarshalDict(nil),
			expect: Dict{},
		},
		"dict": {
			MarshalDict(Dict{"a": MarshalBool(true), "b": MarshalStr8("bbb")}),
			Dict{"a": MarshalBool(true), "b": MarshalStr8("bbb")},
		},
	}
	for k, test := range tests {
		d, _, e := UnmarshalNullDict(test.data)
		if e != nil {
			t.Fatalf("%v: %v", k, e)
		}
		if !reflect.DeepEqual(d, test.expect) {
			t.Fatalf("unmarshaled(%v) = %#v, wants %#v", k, d, test.expect)
		}
	}
}

func TestUnmarshalRoomPropPayload(t *testing.T) {
	const flg = roomPropFlagsJoinable | roomPropFlagsWatchable
	const grp = 17
	const maxp = 13
	const cdl = 23
	pubp := Dict{"pub": MarshalBool(true)}
	prvp := Dict{"prv": MarshalStr8("ok")}

	payload := []byte{}
	payload = append(payload, MarshalByte(flg)...)
	payload = append(payload, MarshalUInt(grp)...)
	payload = append(payload, MarshalUShort(maxp)...)
	payload = append(payload, MarshalUShort(cdl)...)
	payload = append(payload, MarshalDict(pubp)...)
	payload = append(payload, MarshalDict(prvp)...)

	p, err := UnmarshalEvRoomPropPayload(payload)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if wants := flg&roomPropFlagsVisible != 0; p.Visible != wants {
		t.Fatalf("Visible = %v, wants %v", p.Visible, wants)
	}
	if wants := flg&roomPropFlagsJoinable != 0; p.Joinable != wants {
		t.Fatalf("Joinable = %v, wants %v", p.Joinable, wants)
	}
	if wants := flg&roomPropFlagsWatchable != 0; p.Watchable != wants {
		t.Fatalf("Watchable = %v, wants %v", p.Watchable, wants)
	}
	if p.SearchGroup != grp {
		t.Fatalf("SearchGroup = %v, wants %v", p.SearchGroup, grp)
	}
	if p.MaxPlayer != maxp {
		t.Fatalf("MaxPlayer = %v, wants %v", p.MaxPlayer, grp)
	}
	if p.ClientDeadline != cdl {
		t.Fatalf("ClientDeadline = %v, wants %v", p.ClientDeadline, cdl)
	}
	if !reflect.DeepEqual(p.PublicProps, pubp) {
		t.Fatalf("PublicProps = %#v, wants %#v", p.PublicProps, pubp)
	}
	if !reflect.DeepEqual(p.PrivateProps, prvp) {
		t.Fatalf("PrivateProps = %#v, wants %#v", p.PrivateProps, prvp)
	}

	v, err := GetRoomPropClientDeadline(payload)
	if err != nil {
		t.Fatalf("GetRoomPropClientDeadline: %v", err)
	}
	if v != cdl {
		t.Fatalf("RoomPropClientDeadline = %v, wants %v", v, cdl)
	}
}
