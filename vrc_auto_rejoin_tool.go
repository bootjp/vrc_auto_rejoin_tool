package vrcarjt

import (
	"errors"
	"fmt"

	"os/exec"
	"runtime"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/hpcloud/tail"
	"github.com/jinzhu/now"
	gops "github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/process"

	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

const lockfile = "vrc_auto_rejoin_tool.lock"
const WorldLogPrefix = "[VRCFlowManagerVRC] Destination set: wrld_"
const Location = "Local"
const TimeFormat = "2006.01.02 15:04:05"
const vrcRelativeLogPath = `\AppData\LocalLow\VRChat\VRChat\`

func NewVRCAutoRejoinTool() *VRCAutoRejoinTool {
	conf := LoadConf("setting.yml")

	return &VRCAutoRejoinTool{
		Config:         conf,
		Args:           "",
		LatestInstance: Instance{},
		EnableRejoin:   !conf.EnableSleepDetector, // EnableSleepDetectorがOnのとき即座にインスタンス移動の検出をしないため
		InSleep:        false,
		lock:           &sync.Mutex{},
		wait:           &sync.WaitGroup{},
		running:        false,
		done:           make(chan bool),
	}
}

// VRCAutoRejoinTool
type VRCAutoRejoinTool struct {
	Config         *Setting
	Args           string
	LatestInstance Instance
	EnableRejoin   bool
	InSleep        bool
	lock           *sync.Mutex
	wait           *sync.WaitGroup
	running        bool
	done           chan bool
}

type AutoRejoin interface {
	Run() error
	IsRun() bool
	ParseLatestInstance(path string) (Instance, error)
	SleepStart()
	Stop() error

	sleepInstanceDetector() Instance
	setupTimeLocation()
	playAudioFile(path string)
	rejoin(i Instance) error
	logInspector(line *tail.Tail, at time.Time)
	getUserHome() string
	findProcessPIDByName(name string) (int32, error)
	findProcessArgsByName(name string) (string, error)
	killProcessByName(name string) error
	inTimeRange(start time.Time, end time.Time, target time.Time) bool
}

func (v *VRCAutoRejoinTool) IsRun() bool {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.running
}

func (v *VRCAutoRejoinTool) SleepStart() {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.InSleep = true
}

func (v *VRCAutoRejoinTool) Stop() error {
	if !v.running {
		return nil
	}
	v.lock.Lock()
	defer v.lock.Unlock()
	go v.playAudioFile("stop.wav")

	// FOR DEBUGGING
	select {
	case <-v.done:
	default:
		// OPEN CHANNEL
		close(v.done)
	}

	v.running = false

	return nil
}

func (v *VRCAutoRejoinTool) sleepInstanceDetector() Instance {
	return Instance{}
}

func (v *VRCAutoRejoinTool) getUserHome() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

var ErrDuplicateRun = errors.New("auto rejoin tool duplicate run error")

func (v *VRCAutoRejoinTool) Run() error {
	home := v.getUserHome()
	lock := NewDupRunLock(home + `\AppData\Local\Temp\` + lockfile)
	ok, err := lock.Try()

	if err != nil || !ok {
		v.lock.Lock()
		v.running = false
		v.lock.Unlock()
		return ErrDuplicateRun
	}

	if err = lock.Lock(); err != nil {
		v.lock.Lock()
		v.running = false
		v.lock.Unlock()
		return err
	}

	defer lock.UnLock()
	v.setupTimeLocation()

	v.Args, err = v.findProcessArgsByName("VRChat.exe")
	if err == ErrProcessNotFound {
		v.playAudioFile("start_vrc.wav")
		v.lock.Lock()
		v.running = false
		v.lock.Unlock()
		return nil
	}
	if err != nil {
		v.lock.Lock()
		v.running = false
		v.lock.Unlock()
		return err
	}

	v.lock.Lock()
	v.running = true
	v.lock.Unlock()

	go v.playAudioFile("start.wav")
	v.wait.Add(1)

	path := home + vrcRelativeLogPath
	latestLog, err := v.fetchLatestLogName(path)
	if err != nil {
		return fmt.Errorf("log file not found. %s", err)
	}

	start := time.Now().In(time.Local)
	fmt.Println("RUNNING START AT", start.Format(TimeFormat))

	// blocking this
	//for v.Config.EnableSleepDetector && !v.InSleep {
	//	t, err := tail.TailFile(path+latestLog, tail.Config{
	//		Follow:    true,
	//		MustExist: true,
	//		ReOpen:    true,
	//		Poll:      true,
	//	})
	//	if err != nil {
	//		return err
	//	}
	//
	//	for line := range t.Lines {
	//		instance, err := v.moved(start, line.Text)
	//		if err == ErrNotMoved {
	//			continue
	//		}
	//		if err != nil {
	//			log.Println(err)
	//		}
	//		v.LatestInstance = instance
	//
	//		// for sleep detector.
	//		//var _ struct {
	//		//	waitingWorld    bool
	//		//	WorldID         string
	//		//	startTime       time.Time
	//		//	endTime         time.Time
	//		//	waitingDuration time.Duration
	//		//}
	//		//for _, w := range v.Config.SleepWorld {
	//		//	if strings.Contains(instance.ID, w) {
	//		//
	//		//	}
	//		//}
	//	}
	//}

	v.LatestInstance, err = v.ParseLatestInstance(path + latestLog)
	if err != nil {
		return err
	}

	t, err := tail.TailFile(path+latestLog, tail.Config{
		Follow:    true,
		MustExist: true,
		ReOpen:    true,
		Poll:      true,
	})
	if err != nil {
		v.lock.Lock()
		v.running = false
		v.lock.Unlock()
		return err
	}
	if v.Config.EnableProcessCheck {
		go v.processWatcher()
	}
	go v.logInspector(t, start)

	return nil
}

func (v *VRCAutoRejoinTool) rejoin(i Instance) error {
	params := strings.Split(v.Args, `VRChat.exe" `)
	exe := strings.Join(params[:1], "") + `VRChat.exe`
	exe = strings.Trim(exe, `"`)
	cmd := exec.Command(exe, strings.Split(strings.Join(params[1:], "")+` `+`vrchat://launch?id=`+i.ID, ` `)...)

	return cmd.Start()
}

func (v *VRCAutoRejoinTool) ParseLatestInstance(path string) (Instance, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return Instance{}, err
	}

	return v.parseLatestInstance(string(content))

}

// ErrProcessNotFound is an error that is returned when the target process could not be found
var ErrProcessNotFound = errors.New("process not found")

func (v *VRCAutoRejoinTool) findProcessPIDByName(name string) (int32, error) {
	processes, err := gops.Processes()
	if err != nil {
		return -1, err
	}

	for _, p := range processes {
		if strings.Contains(p.Executable(), name) {
			return int32(p.Pid()), nil
		}
	}

	return -1, ErrProcessNotFound
}

func (v *VRCAutoRejoinTool) findProcessArgsByName(name string) (string, error) {
	pid, err := v.findProcessPIDByName(name)
	if err != nil {
		return "", ErrProcessNotFound
	}

	p, err := process.NewProcess(pid)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return p.Cmdline()
}

func (v *VRCAutoRejoinTool) killProcessByName(name string) error {
	pid, err := v.findProcessPIDByName(name)
	if err != nil {
		return err
	}

	p, err := os.FindProcess(int(pid))
	if err != nil {
		return err
	}
	return p.Kill()
}

func (v *VRCAutoRejoinTool) inTimeRange(start time.Time, end time.Time, target time.Time) bool {
	// https://stackoverflow.com/questions/55093676/checking-if-current-time-is-in-a-given-interval-golang
	if start.Before(end) {
		return !target.Before(start) && !target.After(end)
	}
	if start.Equal(end) {
		return target.Equal(start)
	}
	return !start.After(target) || !end.Before(target)
}

func (v *VRCAutoRejoinTool) setupTimeLocation() {
	var err error
	time.Local, err = time.LoadLocation(Location)
	if err != nil {
		time.Local = time.FixedZone(Location, 9*60*60)
	}
}

func (v *VRCAutoRejoinTool) playAudioFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := wav.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = streamer.Close()
	}()

	v.lock.Lock()
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))
	v.lock.Unlock()
	<-done

}

func (v *VRCAutoRejoinTool) parseLatestInstance(s string) (Instance, error) {
	latestInstance := Instance{}

	for _, line := range strings.Split(s, "\n") {
		if line == "" {
			continue
		}
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}

		if !strings.Contains(line, WorldLogPrefix) {
			continue

		}

		instance, err := NewInstanceByLog(line)
		if err != nil {
			return instance, err
		}
		latestInstance = instance
	}
	return latestInstance, nil
}
func (v *VRCAutoRejoinTool) fetchLatestLogName(path string) (string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println(err)
		return "", err
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

func (v *VRCAutoRejoinTool) processWatcher() {

	for range v.done {
		log.Println("process watcher available")
		_, err := v.findProcessPIDByName("VRChat.exe")
		if err == ErrProcessNotFound {
			if v.Config.EnableRejoinNotice {
				v.playAudioFile("rejoin_notice.wav")
				time.Sleep(1 * time.Minute)
			}
			v.lock.Lock()
			err := v.rejoin(v.LatestInstance)
			if err != nil {
				log.Println(err)
			}
			log.Println("process watcher cleanup")
			close(v.done)
			v.running = false
			v.lock.Unlock()
			return
		}
		time.Sleep(10 * time.Second)
	}

}

func (v *VRCAutoRejoinTool) logInspector(tail *tail.Tail, at time.Time) {

	for msg := range tail.Lines {
		text := msg.Text
		nInstance, err := v.moved(at, text)
		if err == ErrNotMoved {
			continue
		}

		if err != nil {
			log.Println(err)

		}

		if v.LatestInstance.ID == nInstance.ID {
			continue
		}

		if v.Config.EnableRadioExercises {
			start, err := now.ParseInLocation(time.Local, "05:45")
			if err != nil {
				log.Println(err)
				continue
			}

			end, err := now.ParseInLocation(time.Local, "08:00")
			if err != nil {
				log.Println(err)
				continue
			}

			if v.inTimeRange(start, end, time.Now().In(time.Local)) {
				continue
			}
		}

		if v.Config.EnableRejoinNotice {
			v.playAudioFile("rejoin_notice.wav")
			time.Sleep(1 * time.Minute)
		}

		v.lock.Lock()

		err = v.killProcessByName("VRChat.exe")
		if err != nil {
			log.Println(err)
		}

		err = v.rejoin(v.LatestInstance)
		if err != nil {
			log.Println(err)
		}

		log.Println("log Watcher clean up")
		v.running = false
		close(v.done)
		v.lock.Unlock()
		tail.Cleanup()
		return
	}
}

// ErrNotMoved is Error when a move cannot be detected in the log
var ErrNotMoved = errors.New("not moved")

func (v *VRCAutoRejoinTool) moved(at time.Time, l string) (Instance, error) {
	if l == "" {
		return Instance{}, ErrNotMoved
	}

	if !strings.Contains(l, WorldLogPrefix) {
		return Instance{}, ErrNotMoved
	}

	i, err := NewInstanceByLog(l)
	if err != nil {
		return i, ErrNotMoved
	}
	if i.Time.Before(at) {
		return Instance{}, ErrNotMoved
	}

	return i, nil

}
