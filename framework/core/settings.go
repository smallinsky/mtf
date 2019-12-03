package core

import (
	"flag"
)

type ArgSettings struct {
	BuildBinary             bool
	StopComponentsAfterExit bool
	Wait                    bool
}

var Settings = ArgSettings{}

func init() {
	flag.BoolVar(&Settings.BuildBinary, "rebuild_binary", true,
		"Determin if SUT binary should be rebuilded before start execution started")

	flag.BoolVar(&Settings.StopComponentsAfterExit, "stop_components", false,
		"Don't stop components after test execution have been finished")

	flag.BoolVar(&Settings.Wait, "wait", false,
		"Don't kill container after test execution")
}
