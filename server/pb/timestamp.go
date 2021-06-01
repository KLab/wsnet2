package pb

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/vmihailenco/msgpack/v4"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (ts *Timestamp) Scan(val interface{}) error {
	t, ok := val.(time.Time)
	if !ok {
		return fmt.Errorf("type is not date.Time: %T, %v", val, val)
	}
	ts.Timestamp = timestamppb.New(t)
	return nil
}

func (ts Timestamp) Value() (driver.Value, error) {
	return ts.Timestamp.AsTime(), nil
}

func (ts Timestamp) Time() time.Time {
	return ts.Timestamp.AsTime()
}

func (ts *Timestamp) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.Encode(ts.Timestamp.Seconds)
}

func (ts *Timestamp) DecodeMsgpack(dec *msgpack.Decoder) error {
	var sec int64
	err := dec.Decode(&sec)
	if err != nil {
		return err
	}

	ts.Timestamp = timestamppb.New(time.Unix(sec, 0))
	return nil
}

var _ msgpack.CustomEncoder = (*Timestamp)(nil)
var _ msgpack.CustomDecoder = (*Timestamp)(nil)
