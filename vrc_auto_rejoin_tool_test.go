package vrcarjt

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/jinzhu/now"
)

func TestParseLatestInstance(t *testing.T) {
	loc, err := time.LoadLocation(Location)
	if err != nil {
		loc = time.FixedZone(Location, 9*60*60)
	}

	t.Run("parse log latest logInspector", func(t *testing.T) {
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

		res, err := NewVRCAutoRejoinTool().parseLatestInstance(string(content))
		if err != nil {
			t.Log(err)
		}
		if !reflect.DeepEqual(eq, res) {
			if !reflect.DeepEqual(eq.Time, res.Time) {
				t.Logf("%v", eq.Time)
				t.Logf("%v", res.Time)
			}

			if eq.ID != res.ID {
				fmt.Printf("%x ", eq.ID)
				fmt.Printf("%x ", res.ID)
				t.Logf("%s", eq.ID)
				t.Logf("%s", res.ID)
			}
			t.FailNow()
		}
	})

	t.Run("parse log latest logInspector in private", func(t *testing.T) {
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

		res, err := NewVRCAutoRejoinTool().parseLatestInstance(string(content))
		if err != nil {
			t.Log(err)
		}

		if err != nil {
			t.Log(err)
		}

		if !reflect.DeepEqual(eq, res) {
			if !reflect.DeepEqual(eq.Time, res.Time) {
				t.Logf("%v", eq.Time)
				t.Logf("%v", res.Time)
			}

			if eq.ID != res.ID {
				fmt.Printf("%x ", eq.ID)
				fmt.Printf("%x ", res.ID)

				t.Logf("%s", eq.ID)
				t.Logf("%s", res.ID)
			}
			t.FailNow()
		}
		//t.
	})

}

func TestNewInstanceByLog(t *testing.T) {
	var err error
	time.Local, err = time.LoadLocation(Location)
	if err != nil {
		time.Local = time.FixedZone(Location, 9*60*60)
	}

	log := `2019.08.18 21:02:38 Log        -  [VRCFlowManagerVRC] Destination set: wrld_cc124ed6-acec-4d55-9866-54ab66af172d`
	ti, err := time.ParseInLocation(TimeFormat, "2019.08.18 21:02:38", time.Local)

	if err != nil {
		t.Fatal(err)
	}
	expect := Instance{ID: "wrld_cc124ed6-acec-4d55-9866-54ab66af172d", Time: ti}
	got, err := NewInstanceByLog(log)
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
	var err error
	time.Local, err = time.LoadLocation(Location)
	if err != nil {
		time.Local = time.FixedZone(Location, 9*60*60)
	}

	freeze := time.Date(2018, 1, 1, 0, 0, 0, 0, time.Local)

	ti, err := time.ParseInLocation(TimeFormat, "2019.08.18 21:02:38", time.Local)
	expect := Instance{ID: "wrld_cc124ed6-acec-4d55-9866-54ab66af172d", Time: ti}

	t.Run("success case", func(t *testing.T) {
		log := `2019.08.18 21:02:38 Log        -  [VRCFlowManagerVRC] Destination set: wrld_cc124ed6-acec-4d55-9866-54ab66af172d`
		if err != nil {
			t.Error(err)
		}
		got, err := NewVRCAutoRejoinTool().moved(freeze, log)
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
		_, err := NewVRCAutoRejoinTool().moved(freeze, log)
		if err != ErrNotMoved {
			t.FailNow()
		}
	})

	t.Run("log has nonce", func(t *testing.T) {

		ti, err := time.ParseInLocation(TimeFormat, "2019.08.18 21:48:39", time.Local)
		expect := Instance{ID: "wrld_58260f57-0076-41d3-a617-c0d0bc8f3d6f:43710~private(usr_d97adcdc-718b-4361-9b75-2c97c0a4993d)~nonce(86CB2A7F4E4AC916CD5A1313F656863C1E80BD2ED63738EA789E2B4C25B48F39)", Time: ti}

		log := `2019.08.18 21:48:39 Log        -  [VRCFlowManagerVRC] Destination set: wrld_58260f57-0076-41d3-a617-c0d0bc8f3d6f:43710~private(usr_d97adcdc-718b-4361-9b75-2c97c0a4993d)~nonce(86CB2A7F4E4AC916CD5A1313F656863C1E80BD2ED63738EA789E2B4C25B48F39)`

		if err != nil {
			t.Error(err)
		}
		got, err := NewVRCAutoRejoinTool().moved(freeze, log)
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
		got, err := NewVRCAutoRejoinTool().moved(freeze, log)
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

func TestFindProcessByName(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}
	cmd := exec.Command("cmd", "/C", "timeout", "3")
	err := cmd.Start()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewVRCAutoRejoinTool().findProcessPIDByName("cmd.exe"); err != nil {
		t.Fatal("process not found")
	}
}

func TestInTimeRange(t *testing.T) {

	loc, err := time.LoadLocation(Location)
	if err != nil {
		loc = time.FixedZone(Location, 9*60*60)
	}
	tests := []struct {
		start   string
		end     string
		check   string
		inRange bool
	}{
		{"05:45", "08:00", "04:00", false},
		{"05:45", "08:00", "23:30", false},
		{"05:45", "08:00", "05:46", true},
		{"05:45", "08:00", "06:00", true},
		{"05:45", "08:00", "08:00", true},
		{"05:45", "08:00", "03:00", false},
	}
	newLayout := "15:04"
	for _, test := range tests {
		check, _ := time.ParseInLocation(newLayout, test.check, loc)
		start, _ := time.ParseInLocation(newLayout, test.start, loc)
		end, _ := time.ParseInLocation(newLayout, test.end, loc)
		if NewVRCAutoRejoinTool().inTimeRange(start, end, check) != test.inRange {
			t.Errorf("test is failed expect %v given %v", test.inRange, NewVRCAutoRejoinTool().inTimeRange(start, end, check))
		}
	}

}
func TestTime(t *testing.T) {
	loc, err := time.LoadLocation("local")
	if err != nil {
		loc = time.FixedZone(Location, 9*60*60)
	}
	tests := []struct {
		check   string
		inRange bool
	}{
		{"04:00", false},
		{"23:30", false},
		{"05:46", true},
		{"06:00", true},
		{"08:00", true},
		{"03:00", false},
	}

	current := time.Now().In(loc)

	for _, test := range tests {

		start, _ := now.ParseInLocation(loc, "05:45")
		end, _ := now.ParseInLocation(loc, "08:00")
		check, _ := now.ParseInLocation(loc, test.check)

		if check.Format("2006-01-02") != current.Format("2006-01-02") {
			t.Errorf("test logic error. check date and current must be equal")
		}

		if NewVRCAutoRejoinTool().inTimeRange(start, end, check) != test.inRange {
			t.Errorf("check %v test is failed expect %v given %v", test.check,
				test.inRange, NewVRCAutoRejoinTool().inTimeRange(start, end, check))
		}
	}
}
