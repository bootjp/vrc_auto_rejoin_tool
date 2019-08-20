package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hpcloud/tail"
)

const WolrdLogPrefix = "[VRCFlowManagerVRC] Destination set: wrld_"
const location = "Asia/Tokyo"

var worldReg = regexp.MustCompile(`(wrld.+?):(\d+)`)

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

func moved(runAt time.Time, l string) bool {
	if l == "" {
		return false
	}

	if !strings.Contains(l, WolrdLogPrefix) {
		return false
	}

	loc, err := time.LoadLocation(location)
	if err != nil {
		loc = time.FixedZone(location, 9*60*60)
	}

	logTime, err := time.ParseInLocation("2006.01.02 15:04:05", l[:19], loc)

	if err != nil {
		fmt.Println(l)
		panic(err)
	}

	if logTime.Before(runAt) {
		return false
	}

	return true
}

func lunch(instance Instance) {
	cmd := &exec.Cmd{
		Path:        os.Getenv("COMSPEC"),
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		SysProcAttr: &syscall.SysProcAttr{CmdLine: `/S /C start vrchat://launch?id=` + instance.ID},
	}
	cmd.Run()

}

func main() {
	path := `C:\Users\bootjp\AppData\LocalLow\VRChat\VRChat\`
	latestLog := ""
	lock := sync.Mutex{}
	var history = Instances{}

	files, err := filepath.Glob(`C:\Users\bootjp\AppData\LocalLow\VRChat\VRChat\output_log*.txt`)
	if err != nil {
		log.Fatal(err)
	}

	sort.Strings(files)
	length := len(files)
	latestLog = files[length]

	loc, err := time.LoadLocation(location)
	if err != nil {
		loc = time.FixedZone(location, 9*60*60)
	}

	startAt := time.Now().In(loc)
	fmt.Println("RUNNING START AT", startAt.Format("2006.01.02 15:04:05"))

	t, err := tail.TailFile(path+latestLog, tail.Config{
		Follow:    true,
		MustExist: true,
	})

	if err != nil {
		log.Fatal(err)
	}
	var msg *tail.Line
	var ok bool
	for true {
		msg, ok = <-t.Lines
		if !ok {
			continue
		}

		text := msg.Text
		// todo 起動時インスタンスの取得
		if moved(startAt, text) {
			lock.Lock()
			fmt.Println("instance move detect!!!")
			fmt.Println(text)
			if latestLog != text {
				latestLog = text
				history = append(history, Instance{Time: time.Time{}, ID: text})
				lunch(history[0])
			}
			lock.Unlock()
		}
	}
}
