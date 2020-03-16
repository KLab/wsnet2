package pb

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
)

func (ts *Timestamp) Scan(val interface{}) error {
	t, ok := val.(time.Time)
	if !ok {
		return fmt.Errorf("type is not date.Time: %T, %v", val, val)
	}
	var err error
	ts.Timestamp, err = ptypes.TimestampProto(t)
	return err
}

func (ts Timestamp) Value() (driver.Value, error) {
	return ptypes.Timestamp(ts.Timestamp)
}

func (ts Timestamp) Time() time.Time {
	t, _ := ptypes.Timestamp(ts.Timestamp)
	return t
}

func (r *RoomInfo) SetCreated(t time.Time) error {
	var err error
	if r.Created == nil {
		r.Created = &Timestamp{}
	}
	r.Created.Timestamp, err = ptypes.TimestampProto(t)
	return err
}
