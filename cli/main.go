package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/hpcloud/tail"
)

const WolrdLogPrefix = "[VRCFlowManagerVRC] Destination set: wrld_"
const location = "Asia/Tokyo"

var worldReg = regexp.MustCompile(`(wrld.+?):(\d+)`)

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

func lunch() {
	cmd := &exec.Cmd{
		Path:        os.Getenv("COMSPEC"),
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		SysProcAttr: &syscall.SysProcAttr{CmdLine: `/S /C start vrchat://launch?id=wrld_dc124ed6-acec-4d55-9866-54ab66af172d:13345~friends(usr_d97adcdc-718b-4361-9b75-2c97c0a4993d)`},
	}
	cmd.Run()

}
func main() {
	path := `C:\Users\bootjp\AppData\LocalLow\VRChat\VRChat\`
	latestLog := ""
	oldLogFile := ""

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("modified file:", event.Name)

					loc, err := time.LoadLocation(location)
					if err != nil {
						loc = time.FixedZone(location, 9*60*60)
					}

					startAt := time.Now().In(loc)
					fmt.Println(startAt.Format("2006.01.02 15:04:05"))

					if !strings.Contains(event.Name, "output_log") {
						return
					}
					if oldLogFile != "" && oldLogFile != event.Name {
						oldLogFile = event.Name
					}
					t, err := tail.TailFile(event.Name, tail.Config{
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
						if moved(startAt, text) {
							fmt.Println("instance move detect!!!")
							fmt.Println(text)
							if latestLog != text {
								latestLog = text
								lunch()
							}

						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		log.Println("watcher")
		log.Fatal(err)
	}
	<-done

}
