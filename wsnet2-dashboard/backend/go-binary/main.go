package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"wsnet2/binary"
)

func main() {
	js.Global().Set("binary", map[string]any{
		"UnmarshalRecursive": js.FuncOf(unmarshalRecursive),
	})
	<-(chan struct{})(nil)
}

// unmarshalRecursive unmarshals binary formatted custom props.
// binary.UnmarshalRecursive(arg number[]): { val: string, err: string }
func unmarshalRecursive(this js.Value, args []js.Value) (ret any) {
	defer func() {
		if err := recover(); err != nil {
			ret = map[string]any{
				"val": "",
				"err": "UnmarshalRecursive: " + err.(error).Error(),
			}
		}
	}()

	arg := args[0]
	len := arg.Length()
	b := make([]byte, len)
	for i := range len {
		v := arg.Index(i).Int() // can be panic
		if v > 255 {
			panic(fmt.Errorf("arg[%v]=%v > 255", i, v))
		}
		b[i] = byte(v)
	}

	v, err := binary.UnmarshalRecursive(b)
	if err != nil {
		panic(err)
	}

	u, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return map[string]any{
		"val": string(u),
		"err": "",
	}
}
