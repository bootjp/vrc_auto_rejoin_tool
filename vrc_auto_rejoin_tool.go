package test

import (
	"errors"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/hpcloud/tail"
	"github.com/jinzhu/now"
	"github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/process"
	"gopkg.in/yaml.v2"
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

func NewVRCAutoRejoinTool(home string) *VRCAutoRejoinTool {
	return &VRCAutoRejoinTool{
		NewDupRunLock(home + `\AppData\Local\Temp\` + lockfile),
		&Setting{},
		"",
		"",
		Instance{},
		time.Local,
	}
}

type VRCAutoRejoinTool struct {
	DLock          *DupRunLock
	Config         *Setting
	Args           string
	VRChatPath     string
	LatestInstance Instance
	loc            *time.Location
}

func (V *VRCAutoRejoinTool) Run() error {
	home := V.getUserHome()

	lock := NewDupRunLock(home + lockfile)
	ok, err := lock.Try()
	if err != nil || !ok {
		log.Println("vrc_auto_rejoin_tool がすでに起動しています．")
		log.Println("多重起動は誤作動の原因となるため，こちらのvrc_auto_rejoin_toolは動作を停止します．")
	}

	err = lock.Lock()
	if err != nil {
		log.Fatal(err)
	}
	defer lock.UnLock()
	V.SetupTimeLocation()
	err = V.LoadConf("setting.yml")
	if err != nil {
		log.Println(err)
	}

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

func (V *VRCAutoRejoinTool) Rejoin(i Instance) (bool, error) {
	cmd := command(i)
	return true, cmd.Start() // todo fix
}

func (V *VRCAutoRejoinTool) ParseLatestInstance(path string) (Instance, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return Instance{}, err
	}

	return V.parseLatestInstance(string(content))

}

func (V *VRCAutoRejoinTool) LoadConf(path string) error {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		//return Setting{}
	}

	t := Setting{}
	err = yaml.Unmarshal(file, &t)
	if err != nil {
		log.Println(err)
		//return Setting{}
	}
	V.Config = &t

	return nil
}

func (V *VRCAutoRejoinTool) getUserHome() string {
	panic("implement me")
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
	if pid, err := V.findProcessPIDByName(name); err != nil {
		p, err := os.FindProcess(int(pid))
		if err != nil {
			return err
		}
		return p.Kill()
	}

	return ErrProcessNotFound
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
	Rejoin(i Instance) (bool, error)
	ParseLatestInstance(path string) (Instance, error)
	LoadConf(path string) error
	SetupTimeLocation()
	Play(path string) error

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

//func (V *VRCAutoRejoinTool) parseLogTime(log string) (time.Time, error) {
//	logTime, err := time.ParseInLocation(TimeFormat, log[:19], V.loc)
//	if err != nil {
//		return logTime, err
//	}
//	return logTime, nil
//}

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
	for range time.Tick(1 * time.Minute) {
		_, err := V.findProcessPIDByName("VRChat.exe")
		if err == ErrProcessNotFound {
			if V.Config.EnableRejoinNotice {
				V.Play("rejoin_notice.wav")
				time.Sleep(1 * time.Minute)
			}
			_, err := V.Rejoin(V.LatestInstance)
			if err != nil {
				log.Println(err)
			}
			time.Sleep(30 * time.Second)
			wg.Done()
			return
		}
	}
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
		if _, err := V.Rejoin(V.LatestInstance); err != nil {
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
