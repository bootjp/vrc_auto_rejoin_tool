package detect

import (
	"time"

	vrcarjt "github.com/bootjp/vrc_auto_rejoin_tool"
)

var _ struct {
	waitingWorld    bool
	WorldID         string
	startTime       time.Time
	endTime         time.Time
	waitingDuration time.Duration
}

type SleepDetect struct {
	Instance vrcarjt.Instance
	time.Ticker
	After func()
}

func NewSleepDetector(i vrcarjt.Instance) *SleepDetect {
	return &SleepDetect{
		Instance: i,
		After: func() {

		},
	}

}

type Detector interface {
	IsSleep() bool
}
