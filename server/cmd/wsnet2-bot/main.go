package main

import (
	"fmt"
	"log"
	"os"
)

var (
	appID  = "testapp"
	appKey = "testapppkey"
)

var (
	WSNet2Version string = "LOCAL"
	WSNet2Commit  string = "LOCAL"
)

type subcmd interface {
	Name() string
	Execute()
}

var cmds = []subcmd{
	NewNormalBot(),
	NewStressBot(),
}

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	fmt.Println("WSNet2-Bot")
	fmt.Println("WSNet2Version:", WSNet2Version)
	fmt.Println("WSNet2Commit:", WSNet2Commit)

	subcmd := "normal"
	if len(os.Args) > 1 {
		subcmd = os.Args[1]
	}
	for _, cmd := range cmds {
		if cmd.Name() == subcmd {
			cmd.Execute()
			return
		}
	}
	log.Printf("command not found: %v", subcmd)
}
