package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/hpcloud/tail"
)

const WorldLogPrefix = "[VRCFlowManagerVRC] Destination set: wrld_"
const Location = "Asia/Tokyo"
const TimeFormat = "2006.01.02 15:04:05"
const vrcRelativeLogPath = `\AppData\LocalLow\VRChat\VRChat\`

type Instance struct {
	Time time.Time
	ID   string
}

var NotMoved = errors.New("this log not moved")

func moved(runAt time.Time, l string, loc *time.Location) (Instance, error) {
	if l == "" {
		return Instance{}, NotMoved
	}

	if !strings.Contains(l, WorldLogPrefix) {
		return Instance{}, NotMoved
	}

	i, err := NewInstanceByLog(l, loc)
	if err != nil {
		return i, NotMoved
	}
	if i.Time.Before(runAt) {
		return Instance{}, NotMoved
	}

	return i, nil
}

func launch(instance Instance) error {
	cmd := command(instance)
	return cmd.Run()
}

func parseLatestInstance(logs string, loc *time.Location) (Instance, error) {
	latestInstance := Instance{}

	for _, line := range strings.Split(logs, "\n") {
		if line == "" {
			continue
		}
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}

		if !strings.Contains(line, WorldLogPrefix) {
			continue
		}

		instance, err := NewInstanceByLog(line, loc)
		if err != nil {
			return instance, err
		}
		latestInstance = instance
	}
	return latestInstance, nil
}

func parseLogTime(log string, loc *time.Location) (time.Time, error) {
	logTime, err := time.ParseInLocation(TimeFormat, log[:19], loc)
	if err != nil {
		return logTime, err
	}
	return logTime, nil
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

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

var debug bool

func setupDebugMode(home string) {
	debug = os.Getenv("DEBUG") == "true"
	debug = debug || strings.Contains(home, "bootjp")
}

func main() {
	home := UserHomeDir()
	if home == "" {
		log.Fatal("home dir not detect.")
	}
	setupDebugMode(home)
	loc, err := time.LoadLocation(Location)
	if err != nil {
		loc = time.FixedZone(Location, 9*60*60)
	}

	path := home + vrcRelativeLogPath
	latestInstance := Instance{}
	lock := sync.Mutex{}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println(err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})
	var filtered []os.FileInfo
	for _, v := range files {
		if strings.Contains(v.Name(), "output_log") {
			filtered = append(filtered, v)
		}
	}

	if debug {
		for _, v := range filtered {
			fmt.Println(v.Name(), v.ModTime().Format(TimeFormat))
		}
	}
	latestLog := ""
	if len(filtered) > 0 {
		latestLog = filtered[0].Name()
	}

	startAt := time.Now().In(loc)
	fmt.Println("RUNNING START AT", startAt.Format(TimeFormat))

	fmt.Println(path + latestLog)
	t, err := tail.TailFile(path+latestLog, tail.Config{
		Follow:    true,
		MustExist: true,
		ReOpen:    true,
		Poll:      true,
	})

	if err != nil {
		log.Println(err)
	}

	content, err := ioutil.ReadFile(path + latestLog)
	if err != nil {
		log.Println(err)
	}

	i, err := parseLatestInstance(string(content), loc)
	if err != nil {
		log.Println(err)
	}
	latestInstance = i
	fmt.Println(i)
	for true {
		msg, ok := <-t.Lines
		if !ok {
			continue
		}

		text := msg.Text
		nInstance, err := moved(startAt, text, loc)
		if err == NotMoved {
			continue
		}
		if debug {
			fmt.Println(text)
		}
		if err != nil {
			log.Println(err)
		}

		lock.Lock()
		fmt.Println("instance move detect!!!")
		if latestInstance != nInstance {
			if debug {
				fmt.Println("latestInstance", latestInstance)
			}
			if err := launch(latestInstance); err != nil {
				log.Println(err)
			}
			latestInstance = nInstance
		}
		lock.Unlock()
	}
}
