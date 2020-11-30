package common

import (
	"wsnet2/binary"

	"golang.org/x/xerrors"
)

func InitProps(props []byte) (binary.Dict, []byte, error) {
	if len(props) == 0 || binary.Type(props[0]) == binary.TypeNull {
		dict := binary.Dict{}
		return dict, binary.MarshalDict(dict), nil
	}
	um, _, err := binary.Unmarshal(props)
	if err != nil {
		return nil, nil, err
	}
	dict, ok := um.(binary.Dict)
	if !ok {
		return nil, nil, xerrors.Errorf("type is not Dict: %v", binary.Type(props[0]))
	}
	return dict, props, nil
}
