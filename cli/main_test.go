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
	fmt.Println(d)
	content, err := ioutil.ReadFile(d + "/test_log.txt")
	if err != nil {
		t.Error(err)
	}
	// time.Now().l
	res := parseLatestInstance(string(content), loc)
	if eq == res {
		t.Logf("%v", eq)
		t.Logf("%v", res)
		t.FailNow()
	}
}
