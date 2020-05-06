package main

import (
	"log"

	vrcarjt "github.com/bootjp/vrc_auto_rejoin_tool"
)

func main() {
	err := vrcarjt.NewVRCAutoRejoinTool().Run()
	if err != nil {
		log.Fatal(err)
	}
}
