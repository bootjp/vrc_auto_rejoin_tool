package vrcarjt

import (
	"bytes"
	"errors"
	"log"
	"regexp"
	"time"
)

type Instance struct {
	Time time.Time
	ID   string
}

var worldRegexp = regexp.MustCompile(`wrld_.+$`)
func NewInstanceByLog(logs string) (Instance, error) {
	lt, err := parseLogTime(logs)
	if err != nil {
		log.Println(err)
		return Instance{}, err
	}
	group := worldRegexp.FindSubmatch([]byte(logs))
	if len(group) > 0 {
		return Instance{ID: string(bytes.Trim(group[0], "\x00")), Time: lt}, nil
	}

	return Instance{}, errors.New("world log not found")
}

func parseLogTime(log string) (time.Time, error) {
	logTime, err := time.ParseInLocation(TimeFormat, log[:19], time.Local)
	if err != nil {
		return logTime, err
	}
	return logTime, nil
}
