package vrc_auto_rejoin_tool

import (
	"errors"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/hpcloud/tail"
	"github.com/jinzhu/now"
	"github.com/mitchellh/go-ps"
	"github.com/okzk/ticker"
	"github.com/shirou/gopsutil/process"
	"os/exec"
	"runtime"

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
		conf,
		"",
		Instance{},
		time.Local,
	}
}

type VRCAutoRejoinTool struct {
	Config         *Setting
	Args           string
	LatestInstance Instance
	loc            *time.Location
}

func (V *VRCAutoRejoinTool) getUserHome() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func (V *VRCAutoRejoinTool) Run() error {
	home := V.getUserHome()
	lock := NewDupRunLock(home + `\AppData\Local\Temp\` + lockfile)
	ok, err := lock.Try()
	if err != nil || !ok {
		log.Println("vrc_auto_rejoin_tool がすでに起動しています．")
		log.Println("多重起動は誤作動の原因となるため，このウィンドウのvrc_auto_rejoin_toolは動作を停止します．")
		return err
	}

	err = lock.Lock()
	if err != nil {
		log.Fatal(err)
	}
	defer lock.UnLock()
	V.SetupTimeLocation()

	go V.Play("start.wav")

	wg := &sync.WaitGroup{}
	wg.Add(1)

	V.Args, err = V.findProcessArgsByName("VRChat.exe")
	if err != nil {
		log.Println(err)
	}

	path := home + vrcRelativeLogPath

	latestLog, err := V.fetchLatestLogName(path)

	if err != nil {
		log.Fatalf("log file not found. %s", err)
	}

	start := time.Now().In(V.loc)
	fmt.Println("RUNNING START AT", start.Format(TimeFormat))

	V.LatestInstance, err = V.ParseLatestInstance(path + latestLog)
	if err != nil {
		log.Println(err)
	}

	if V.Config.EnableProcessCheck {
		go V.checkProcess(wg)
	}

	fmt.Println(path + latestLog)
	go V.checkMoveInstance(path, latestLog, start, wg)
	wg.Wait()

	return nil
}

func (V *VRCAutoRejoinTool) Rejoin(i Instance) error {
	params := strings.Split(V.Args, `VRChat.exe" `)
	exe := strings.Join(params[:1], "") + `VRChat.exe`
	exe = strings.Trim(exe, `"`)
	cmd := exec.Command(exe, strings.Split(strings.Join(params[1:], "")+` `+`vrchat://launch?id=`+i.ID, ` `)...)

	return cmd.Start()
}

func (V *VRCAutoRejoinTool) ParseLatestInstance(path string) (Instance, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return Instance{}, err
	}

	return V.parseLatestInstance(string(content))

}

var ErrProcessNotFound = errors.New("process not found")

func (V *VRCAutoRejoinTool) findProcessPIDByName(name string) (int32, error) {
	processes, err := ps.Processes()
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

func (V *VRCAutoRejoinTool) findProcessArgsByName(name string) (string, error) {
	pid, err := V.findProcessPIDByName(name)
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

func (V *VRCAutoRejoinTool) killProcessByName(name string) error {
	pid, err := V.findProcessPIDByName(name)
	if err != nil {
		return err
	}

	p, err := os.FindProcess(int(pid))
	if err != nil {
		return err
	}
	return p.Kill()
}

func (V *VRCAutoRejoinTool) inTimeRange(start time.Time, end time.Time, target time.Time) bool {
	// https://stackoverflow.com/questions/55093676/checking-if-current-time-is-in-a-given-interval-golang
	if start.Before(end) {
		return !target.Before(start) && !target.After(end)
	}
	if start.Equal(end) {
		return target.Equal(start)
	}
	return !start.After(target) || !end.Before(target)
}

type AutoRejoin interface {
	Run() error
	Rejoin(i Instance) error
	ParseLatestInstance(path string) (Instance, error)
	SetupTimeLocation()
	Play(path string)

	getUserHome() string
	findProcessPIDByName(name string) (int32, error)
	findProcessArgsByName(name string) (string, error)
	killProcessByName(name string) error
	inTimeRange(start time.Time, end time.Time, target time.Time) bool
}

func (V *VRCAutoRejoinTool) SetupTimeLocation() {
	var err error
	V.loc, err = time.LoadLocation(Location)
	if err != nil {
		time.Local = time.FixedZone(Location, 9*60*60)
		V.loc = time.Local
	}
}

func (V *VRCAutoRejoinTool) Play(path string) {
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

func (V *VRCAutoRejoinTool) parseLatestInstance(s string) (Instance, error) {
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

		instance, err := NewInstanceByLog(line, V.loc)
		if err != nil {
			return instance, err
		}
		latestInstance = instance
	}
	return latestInstance, nil
}
func (V *VRCAutoRejoinTool) fetchLatestLogName(path string) (string, error) {
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

func (V *VRCAutoRejoinTool) checkProcess(wg *sync.WaitGroup) {
	_ = ticker.New(1*time.Minute, func(_ time.Time) {
		_, err := V.findProcessPIDByName("VRChat.exe")
		if err == ErrProcessNotFound {
			if V.Config.EnableRejoinNotice {
				V.Play("rejoin_notice.wav")
				time.Sleep(1 * time.Minute)
			}
			err := V.Rejoin(V.LatestInstance)
			if err != nil {
				log.Println(err)
			}
			time.Sleep(30 * time.Second)
			wg.Done()
			return
		}
	})
}

func (V *VRCAutoRejoinTool) checkMoveInstance(path string, latestLog string, at time.Time, wg *sync.WaitGroup) {
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
		nInstance, err := V.moved(at, text)
		if err == ErrNotMoved {
			continue
		}

		if err != nil {
			log.Println(err)
		}

		if V.LatestInstance == nInstance {
			continue
		}

		if V.Config.EnableRadioExercises {
			start, err := now.ParseInLocation(V.loc, "05:45")
			if err != nil {
				log.Println(err)
				continue
			}

			end, err := now.ParseInLocation(V.loc, "08:00")
			if err != nil {
				log.Println(err)
				continue
			}

			if V.inTimeRange(start, end, time.Now().In(V.loc)) {
				continue
			}
		}

		if V.Config.EnableRejoinNotice {
			V.Play("rejoin_notice.wav")
			time.Sleep(1 * time.Minute)
		}

		err = V.killProcessByName("VRChat.exe")
		if err != nil {
			log.Println(err)
		}
		if err := V.Rejoin(V.LatestInstance); err != nil {
			log.Println(err)
		}
		time.Sleep(30 * time.Second)
		wg.Done()
		return
	}
}

var ErrNotMoved = errors.New("not moved")

func (V *VRCAutoRejoinTool) moved(at time.Time, l string) (Instance, error) {
	if l == "" {
		return Instance{}, ErrNotMoved
	}

	if !strings.Contains(l, WorldLogPrefix) {
		return Instance{}, ErrNotMoved
	}

	i, err := NewInstanceByLog(l, V.loc)
	if err != nil {
		return i, ErrNotMoved
	}
	if i.Time.Before(at) {
		return Instance{}, ErrNotMoved
	}

	return i, nil

}
