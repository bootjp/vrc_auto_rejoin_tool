package vrc_auto_rejoin_tool

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

func NewInstanceByLog(logs string, loc *time.Location) (Instance, error) {
	r := regexp.MustCompile(`wrld_.+$`)

	lt, err := parseLogTime(logs, loc)
	if err != nil {
		log.Println(err)
	}
	group := r.FindSubmatch([]byte(logs))
	if len(group) > 0 {
		return Instance{ID: string(bytes.Trim(group[0], "\x00")), Time: lt}, nil
	}

	return Instance{}, errors.New("world log not found")
}

func parseLogTime(log string, loc *time.Location) (time.Time, error) {
	logTime, err := time.ParseInLocation(TimeFormat, log[:19], loc)
	if err != nil {
		return logTime, err
	}
	return logTime, nil
}
