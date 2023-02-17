package main

import (
	"github.com/gopherjs/gopherjs/js"

	"wsnet2/binary"
)

func main() {
	ex := js.Module.Get("exports")
	ex.Set("UnmarshalRecursive", binary.UnmarshalRecursive)
}
