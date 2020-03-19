package main

import (
	"testing"
)

func TestYmlLoad(t *testing.T) {
	cnf := loadSetting()
	if cnf.Debug {
		t.Fatal()
	}
}
