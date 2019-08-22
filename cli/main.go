// +build darwin

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hpcloud/tail"
)

const WolrdLogPrefix = "[VRCFlowManagerVRC] Destination set: wrld_"
const Location = "Asia/Tokyo"
const TimeFormat = "2006.01.02 15:04:05"

type Instance struct {
	Time time.Time
	ID   string
}
type Instances []Instance

func (in Instances) Len() int {
	return len(in)
}

func (in Instances) Less(i, j int) bool {
	return in[i].Time.Before(in[j].Time)
}

func (in Instances) Swap(i, j int) {
	in[i], in[j] = in[j], in[i]
}

var NotMoved = errors.New("this log not moved")

func moved(runAt time.Time, l string, loc *time.Location) (Instance, error) {
	if l == "" {
		return Instance{}, NotMoved
	}

	if !strings.Contains(l, WolrdLogPrefix) {
		return Instance{}, NotMoved
	}

	i := NewInstanceByLog(l, loc)

	if i.Time.Before(runAt) {
		return Instance{}, NotMoved
	}

	return i, nil
}

func lunch(instance Instance) error {
	cmd := &exec.Cmd{
		Path:   os.Getenv("COMSPEC"),
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		SysProcAttr: &syscall.SysProcAttr{
			CmdLine: `/S /C start vrchat://launch?id=` + instance.ID,
			// Foreground: true,
		}, // when run non windows env please comment out this line.
	}

	out, err := cmd.Output()
	fmt.Println(out)
	return err
}

func parseLatestInstance(logs string, loc *time.Location) (Instance, error) {
	i := Instance{}

	for _, line := range strings.Split(logs, "\n") {
		if line == "" {
			continue
		}

		if !strings.Contains(line, WolrdLogPrefix) {
			continue
		}

		i = NewInstanceByLog(line, loc)
	}
	return i, nil
}

func parseLogTime(log string, loc *time.Location) (time.Time, error) {
	logTime, err := time.ParseInLocation(TimeFormat, log[:19], loc)
	if err != nil {
		return logTime, err
	}
	return logTime, nil
}

func NewInstanceByLog(logs string, loc *time.Location) Instance {
	r := regexp.MustCompile(`wrld_.+`)

	lt, err := parseLogTime(logs, loc)
	if err != nil {
		log.Fatal(err)
	}
	group := r.FindSubmatch([]byte(logs))
	return Instance{ID: string(group[0]), Time: lt}
}

// todo 今の実装ではlucherを起動したあとにログのtailをしないので治す
func main() {
	loc, err := time.LoadLocation(Location)
	if err != nil {
		loc = time.FixedZone(Location, 9*60*60)
	}

	path := `C:\Users\bootjp\AppData\LocalLow\VRChat\VRChat\`
	fmt.Println(path)
	latestInstance := Instance{}
	lock := sync.Mutex{}
	var history = Instances{}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	// todo reverse
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})
	var filtered []os.FileInfo
	for _, v := range files {
		if strings.Contains(v.Name(), "output_log") {
			filtered = append(filtered, v)
		}
	}

	for _, v := range filtered {
		fmt.Println(v.Name())
	}

	max := len(filtered) - 1
	latestLog := filtered[max].Name()
	startAt := time.Now().In(loc)
	fmt.Println("RUNNING START AT", startAt.Format(TimeFormat))

	fmt.Println(path + latestLog)
	t, err := tail.TailFile(path+latestLog, tail.Config{
		Follow:    true,
		MustExist: true,
	})

	if err != nil {
		log.Fatal(err)
	}

	content, err := ioutil.ReadFile(path + latestLog)
	if err != nil {
		log.Fatal(err)
	}

	i, err := parseLatestInstance(string(content), loc)
	if err != nil {
		log.Fatal(err)
	}
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
		if err != nil {
			log.Fatal(err)
		}

		lock.Lock()
		fmt.Println("instance move detect!!!")
		if latestInstance != nInstance {
			latestInstance = nInstance
			history = append(history, nInstance)
			if err := lunch(history[0]); err != nil {
				log.Fatal(err)
			}
		}
		lock.Unlock()
	}
}
