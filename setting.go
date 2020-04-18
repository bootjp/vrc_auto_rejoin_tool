package vrc_auto_rejoin_tool

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Setting struct {
	EnableProcessCheck   bool `yaml:"enable_process_check"`
	Debug                bool `yaml:"debug"`
	EnableRadioExercises bool `yaml:"enable_radio_exercises"`
	EnableRejoinNotice   bool `yaml:"enable_rejoin_notice"`
}

func LoadConf(path string) *Setting {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return &Setting{
			EnableRejoinNotice: true,
		}
	}

	t := Setting{}
	err = yaml.Unmarshal(file, &t)
	if err != nil {
		log.Println(err)
		return &Setting{
			EnableRejoinNotice: true,
		}
	}
	return &t
}
