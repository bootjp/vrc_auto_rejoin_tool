package main

import (
	"testing"
)

func TestYmlLoad(t *testing.T) {
	cnf := loadSetting()
	if !cnf.EnableProcessCheck {
		t.Fatal()
	}

	if cnf.Debug {
		t.Fatal()
	}
	if !cnf.EnableRadioExercises {
		t.Fatal()
	}

}
