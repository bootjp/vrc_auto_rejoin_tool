package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestParseLatestInstance(t *testing.T) {
	loc, err := time.LoadLocation(location)
	if err != nil {
		loc = time.FixedZone(location, 9*60*60)
	}

	lt, err := time.ParseInLocation("2006.01.02 15:04:05", "2019.08.18 21:02:38", loc)
	if err != nil {
		t.Error(err)
	}

	eq := Instance{Time: lt, ID: "2019.08.18 21:02:38 Log        -  [VRCFlowManagerVRC] Destination set: wrld_cc124ed6-acec-4d55-9866-54ab66af172d"}
	d, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	content, err := ioutil.ReadFile(d + "/test_log.txt")
	if err != nil {
		t.Error(err)
	}

	res, err := parseLatestInstance(string(content), loc)
	if err != nil {
		t.Log(err)
	}
	if eq == res {
		t.Logf("%v", eq)
		t.Logf("%v", res)
		t.FailNow()
	}
}

func TestNewInstanceByLog(t *testing.T) {
	loc, err := time.LoadLocation(location)
	if err != nil {
		loc = time.FixedZone(location, 9*60*60)
	}

	log := `2019.08.18 21:02:38 Log        -  [VRCFlowManagerVRC] Destination set: wrld_cc124ed6-acec-4d55-9866-54ab66af172d`
	ti, err := time.ParseInLocation(timeFormat, "2019.08.18 21:02:38", loc)

	if err != nil {
		t.Fatal(err)
	}
	expect := Instance{ID: "wrld_cc124ed6-acec-4d55-9866-54ab66af172d", Time: ti}
	got := NewInstanceByLog(log, loc)
	if expect != got {
		fmt.Printf("%v\n", expect)
		fmt.Printf("%v\n", got)
		t.Fatal()
	}
}
