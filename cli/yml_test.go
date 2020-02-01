package main

import (
	"fmt"
	"testing"
)

func TestYmlLoad(t *testing.T) {
	cnf := loadSetting()
	fmt.Printf("%v", cnf)
	if cnf.Debug == false {
		t.Fatal()
	}
}
