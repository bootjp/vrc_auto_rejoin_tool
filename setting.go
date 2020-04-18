package test

type Setting struct {
	EnableProcessCheck   bool `yaml:"enable_process_check"`
	Debug                bool `yaml:"debug"`
	EnableRadioExercises bool `yaml:"enable_radio_exercises"`
	EnableRejoinNotice   bool `yaml:"enable_rejoin_notice"`
}
