package vrcarjt

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

// Setting is vrc auto rejoin tool behavior setting
type Setting struct {
	EnableProcessCheck   bool     `yaml:"enable_process_check"`
	Debug                bool     `yaml:"debug"`
	EnableRadioExercises bool     `yaml:"enable_radio_exercises"`
	EnableRejoinNotice   bool     `yaml:"enable_rejoin_notice"`
	EnableDaemon         bool     `yaml:"enable_daemon"`
	EnableSleepDetector  bool     `yaml:"enable_sleep_detector"`
	SleepWorld           []string `yaml:"sleep_world"`
}

var defaultSetting = &Setting{
	EnableProcessCheck:   true,
	Debug:                false,
	EnableRadioExercises: false,
	EnableRejoinNotice:   true,
	EnableDaemon:         false,
	EnableSleepDetector:  false,
}

func LoadConf(path string) *Setting {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("invalid config yml fallback to default setting")
		log.Println(err)
		return defaultSetting
	}

	t := Setting{}
	err = yaml.Unmarshal(file, &t)
	if err != nil {
		log.Println("invalid config yml fallback to default setting")
		log.Println(err)
		return defaultSetting

	}
	return &t
}
