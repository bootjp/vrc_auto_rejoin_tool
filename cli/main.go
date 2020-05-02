package main

import (
	vrcarjt "github.com/bootjp/vrc_auto_rejoin_tool"
	"log"
)

func main() {
	err := vrcarjt.NewVRCAutoRejoinTool().Run()
	log.Fatal(err)
}
