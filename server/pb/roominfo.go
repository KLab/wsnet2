package pb

import (
	"database/sql/driver"
	"time"

	"golang.org/x/xerrors"

	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (r *RoomInfo) SetCreated(t time.Time) {
	if r.Created == nil {
		r.Created = &Timestamp{}
	}
	r.Created.Timestamp = timestamppb.New(t)
}

func (n *RoomNumber) Scan(val interface{}) error {
	switch v := val.(type) {
	case nil:
		n.Number = 0
		return nil
	case int64:
		n.Number = int32(v)
		return nil
	}
	return xerrors.Errorf("invalid value type: %T %v", val, val)
}

func (n *RoomNumber) Value() (driver.Value, error) {
	if n.Number == 0 {
		return nil, nil
	}
	return int64(n.Number), nil
}

func (n *RoomNumber) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.Encode(n.Number)
}

func (n *RoomNumber) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.Decode(&n.Number)
}
