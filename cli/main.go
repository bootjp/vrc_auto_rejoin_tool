package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/jinzhu/now"
	"github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/process"
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
	yaml "gopkg.in/yaml.v2"
)

const WorldLogPrefix = "[VRCFlowManagerVRC] Destination set: wrld_"
const Location = "Local"
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

func playAudio(file string) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := wav.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}

var latestInstance Instance

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

type setting struct {
	EnableProcessCheck   bool `yaml:"enable_process_check"`
	Debug                bool `yaml:"debug"`
	EnableRadioExercises bool `yaml:"enable_radio_exercises"`
	EnableRejoinNotice   bool `yaml:"enable_rejoin_notice"`
}

// if setting file does not exits fallback to default setting.
func loadSetting() setting {
	file, err := ioutil.ReadFile("setting.yml")
	if err != nil {
		log.Println(err)
		return setting{}
	}

	fmt.Printf("%s\n", file)
	t := setting{}
	err = yaml.Unmarshal(file, &t)
	if err != nil {
		log.Println(err)
		return setting{}
	}

	return t
}

var conf setting

func debugLog(l ...interface{}) {
	if conf.Debug {
		log.Printf("%v", l)
	}
}

func loadLatestInstance(filepath string, location *time.Location) (Instance, error) {

	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Println(err)
	}

	return parseLatestInstance(string(content), location)
}

func fetchLatestLogName(path string) (string, error) {
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

	latestLog := ""
	if len(filtered) > 0 {
		latestLog = filtered[0].Name()
	}

	return latestLog, nil
}

func InTimeRange(start time.Time, end time.Time, target time.Time) bool {
	// https://stackoverflow.com/questions/55093676/checking-if-current-time-is-in-a-given-interval-golang
	if start.Before(end) {
		return !target.Before(start) && !target.After(end)
	}
	if start.Equal(end) {
		return target.Equal(start)
	}
	return !start.After(target) || !end.Before(target)
}

func checkMoveInstance(path string, latestLog string, startAt time.Time, loc *time.Location, wg *sync.WaitGroup) {
	t, err := tail.TailFile(path+latestLog, tail.Config{
		Follow:    true,
		MustExist: true,
		ReOpen:    true,
		Poll:      true,
	})
	if err != nil {
		log.Println(err)
	}
	for {
		msg, ok := <-t.Lines
		if !ok {
			continue
		}

		text := msg.Text
		nInstance, err := moved(startAt, text, loc)
		if err == NotMoved {
			continue
		}
		debugLog(text)

		if err != nil {
			log.Println(err)
		}

		if latestInstance == nInstance {
			continue
		}

		if conf.EnableRadioExercises {
			start, err := now.ParseInLocation(loc, "05:45")
			if err != nil {
				log.Println(err)
				continue
			}

			end, err := now.ParseInLocation(loc, "08:00")
			if err != nil {
				log.Println(err)
				continue
			}

			if InTimeRange(start, end, time.Now().In(loc)) {
				continue
			}
		}

		debugLog("detected instance move")
		debugLog("latestInstance", latestInstance)

		if conf.EnableRejoinNotice {
			playAudio("rejoin_notice.wav")
			time.Sleep(1 * time.Minute)
		}

		err = KillProcessByName("VRChat.exe")
		if err != nil {
			log.Println(err)
		}
		if err := launch(latestInstance); err != nil {
			log.Println(err)
		}
		wg.Done()
		return
	}
}
func KillProcessByName(name string) error {
	if exits, pid := findProcessByName(name); exits {
		process, err := os.FindProcess(pid)
		if err != nil {
			return err
		}
		err = process.Kill()
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

func main() {
	playAudio("start.wav")
	wg := &sync.WaitGroup{}
	wg.Add(1)
	home := UserHomeDir()
	runArgs, err := findProcessArgsByName("VRChat")
	if err != nil {
		log.Println(err)
	}

	debugLog(runArgs)

	if home == "" {
		log.Fatal("home dir not detect.")
	}

	conf = loadSetting()

	debugLog(conf)

	loc, err := time.LoadLocation(Location)
	if err != nil {
		time.Local = time.FixedZone(Location, 9*60*60)
	}

	path := home + vrcRelativeLogPath

	latestLog, err := fetchLatestLogName(path)

	if err != nil {
		log.Fatalf("log file not found. %s", err)
	}

	startAt := time.Now().In(loc)
	fmt.Println("RUNNING START AT", startAt.Format(TimeFormat))

	latestInstance, err = loadLatestInstance(path+latestLog, loc)
	if err != nil {
		log.Println(err)
	}

	if conf.EnableProcessCheck {
		go checkProcess(wg)
	}

	fmt.Println(path + latestLog)
	go checkMoveInstance(path, latestLog, startAt, loc, wg)
	wg.Wait()
}

func findProcessByName(name string) (bool, int) {
	processes, err := ps.Processes()
	if err != nil {
		return false, -1
	}

	for _, p := range processes {
		if strings.Contains(p.Executable(), name) {
			return true, p.Pid()
		}
	}

	return false, -1
}

func findProcessArgsByName(n string) ([]string, error) {
	ok, pid := findProcessByName(n)
	if !ok {
		return nil, errors.New("process does not exits")
	}

	p, err := process.NewProcess(int32(pid))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return p.CmdlineSlice()
}

func checkProcess(wg *sync.WaitGroup) {
	for range time.Tick(10 * time.Second) {

		debugLog("check process exits")
		exists, _ := findProcessByName("VRChat.exe")
		if !exists {
			debugLog("process does not exits")
			if conf.EnableRejoinNotice {
				playAudio("rejoin_notice.wav")
				time.Sleep(1 * time.Minute)
			}
			err := launch(latestInstance)
			if err != nil {
				log.Println(err)
			}
			wg.Done()
			return // throw checkProcess
		}
	}

}
