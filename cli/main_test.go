package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestParseLatestInstance(t *testing.T) {
	loc, err := time.LoadLocation(Location)
	if err != nil {
		loc = time.FixedZone(Location, 9*60*60)
	}
	t.Run("parse log latest instance", func(t *testing.T) {
		lt, err := time.ParseInLocation("2006.01.02 15:04:05", "2019.08.18 21:02:38", loc)
		if err != nil {
			t.Error(err)
		}

		eq := Instance{Time: lt, ID: "wrld_cc124ed6-acec-4d55-9866-54ab66af172d"}
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
		if eq != res {
			t.Logf("%v", eq)
			t.Logf("%v", res)
			t.FailNow()
		}
	})

	t.Run("parse log latest instance in private", func(t *testing.T) {
		lt, err := time.ParseInLocation("2006.01.02 15:04:05", "2019.08.24 13:30:14", loc)
		if err != nil {
			t.Error(err)
		}
		//
		eq := Instance{Time: lt, ID: "wrld_7344b9f5-06e1-4e30-bede-fde72d2e5455:37969~private(usr_d97adcdc-718b-4361-9b75-2c97c0a4993d)~canRequestInvite~nonce(3A7A1F9FFE3F87C45D978535DADD3CEFB007D9249366A1BCED70A96FD4740D3C)"}
		d, err := os.Getwd()
		if err != nil {
			t.Error(err)
		}

		content, err := ioutil.ReadFile(d + "/test_log2.txt")
		if err != nil {
			t.Error(err)
		}

		res, err := parseLatestInstance(string(content), loc)
		if err != nil {
			t.Log(err)
		}

		if err != nil {
			t.Log(err)
		}
		t.Logf("%v", eq)
		t.Logf("%v", res)

		if eq != res {
			t.Logf("%v", eq)
			t.Logf("%v", res)
			t.FailNow()
		}
	})

}

func TestNewInstanceByLog(t *testing.T) {
	loc, err := time.LoadLocation(Location)
	if err != nil {
		loc = time.FixedZone(Location, 9*60*60)
	}

	log := `2019.08.18 21:02:38 Log        -  [VRCFlowManagerVRC] Destination set: wrld_cc124ed6-acec-4d55-9866-54ab66af172d`
	ti, err := time.ParseInLocation(TimeFormat, "2019.08.18 21:02:38", loc)

	if err != nil {
		t.Fatal(err)
	}
	expect := Instance{ID: "wrld_cc124ed6-acec-4d55-9866-54ab66af172d", Time: ti}
	got, err := NewInstanceByLog(log, loc)
	if err != nil {
		t.Error(err)
	}
	if expect != got {
		fmt.Printf("%v\n", expect)
		fmt.Printf("%v\n", got)
		t.Fatal()
	}
}

func TestMove(t *testing.T) {
	loc, err := time.LoadLocation(Location)
	if err != nil {
		loc = time.FixedZone(Location, 9*60*60)
	}

	freeze := time.Date(2018, 1, 1, 0, 0, 0, 0, loc)

	ti, err := time.ParseInLocation(TimeFormat, "2019.08.18 21:02:38", loc)
	expect := Instance{ID: "wrld_cc124ed6-acec-4d55-9866-54ab66af172d", Time: ti}

	t.Run("success case", func(t *testing.T) {
		log := `2019.08.18 21:02:38 Log        -  [VRCFlowManagerVRC] Destination set: wrld_cc124ed6-acec-4d55-9866-54ab66af172d`
		if err != nil {
			t.Error(err)
		}
		got, err := moved(freeze, log, loc)
		if err != nil {
			t.Error(err)
		}

		if expect != got {
			fmt.Printf("%v\n", expect)
			fmt.Printf("%v\n", got)
			t.FailNow()
		}
	})

	t.Run("log not found case", func(t *testing.T) {
		log := `2019.08.18 21:02:38 Log `
		if err != nil {
			t.Error(err)
		}
		_, err := moved(freeze, log, loc)
		if err != NotMoved {
			t.FailNow()
		}
	})

	t.Run("log has nonce", func(t *testing.T) {

		ti, err := time.ParseInLocation(TimeFormat, "2019.08.18 21:48:39", loc)
		expect := Instance{ID: "wrld_58260f57-0076-41d3-a617-c0d0bc8f3d6f:43710~private(usr_d97adcdc-718b-4361-9b75-2c97c0a4993d)~nonce(86CB2A7F4E4AC916CD5A1313F656863C1E80BD2ED63738EA789E2B4C25B48F39)", Time: ti}

		log := `2019.08.18 21:48:39 Log        -  [VRCFlowManagerVRC] Destination set: wrld_58260f57-0076-41d3-a617-c0d0bc8f3d6f:43710~private(usr_d97adcdc-718b-4361-9b75-2c97c0a4993d)~nonce(86CB2A7F4E4AC916CD5A1313F656863C1E80BD2ED63738EA789E2B4C25B48F39)`

		if err != nil {
			t.Error(err)
		}
		got, err := moved(freeze, log, loc)
		if err != nil {
			t.Error(err)
		}

		if expect != got {
			fmt.Printf("%v\n", expect)
			fmt.Printf("%v\n", got)
			t.FailNow()
		}
	})

	t.Run("success case", func(t *testing.T) {
		log := `2019.08.18 21:02:38 Log        -  [VRCFlowManagerVRC] Destination set: wrld_cc124ed6-acec-4d55-9866-54ab66af172d`
		if err != nil {
			t.Error(err)
		}
		got, err := moved(freeze, log, loc)
		if err != nil {
			t.Error(err)
		}

		if expect != got {
			fmt.Printf("%v\n", expect)
			fmt.Printf("%v\n", got)
			t.FailNow()
		}
	})
}
