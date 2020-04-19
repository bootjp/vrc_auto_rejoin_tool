package main

import (
	vrcajt "github.com/bootjp/vrc_auto_rejoin_tool"
	"log"
)

func main() {
	err := vrcajt.NewVRCAutoRejoinTool().Run()
	log.Fatal(err)
}
