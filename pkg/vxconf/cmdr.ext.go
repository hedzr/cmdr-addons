// Copyright © 2020 Hedzr Yeh.

package vxconf

import (
	"fmt"
	"github.com/hedzr/cmdr"
	"gopkg.in/yaml.v3"
)

// RunMode return running mode string: prod, devel, staging, ...
func RunMode(defaultVal ...string) (runMode string) {
	if len(defaultVal) == 0 {
		runMode = cmdr.GetStringR("runmode", "prod")
	} else {
		runMode = cmdr.GetStringR("runmode", defaultVal...)
	}
	return
}

// RunModeExt return running mode and position: prod-newyork, devel-paris, staging, ...
func RunModeExt(defaultVal ...string) (runMode string) {
	if len(defaultVal) == 0 {
		runMode = cmdr.GetStringR("runmode", "prod")
	} else {
		runMode = cmdr.GetStringR("runmode", defaultVal...)
	}

	pos := cmdr.GetStringR("pos")
	if pos != "" {
		runMode = fmt.Sprintf("%s-%s", runMode, pos)
	}
	return
}

// IsProd return true if app is in production mode.
func IsProd() bool {
	switch RunMode("devel") {
	case "prod", "production":
		return true
	}
	return false
}

// LoadSectionTo returns error while cannot yaml Marshal and Unmarshal
func LoadSectionTo(runMode, sectionKeyPath string, configHolder interface{}) (err error) {
	var b []byte

	aKey := fmt.Sprintf("%s.%s", sectionKeyPath, runMode)
	fObj := cmdr.GetMapR(aKey)
	if fObj == nil {
		fObj = cmdr.GetMapR(sectionKeyPath)
	}

	b, err = yaml.Marshal(fObj)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(b, configHolder)
	if err != nil {
		return
	}

	// logrus.Debugf("configuration section got: %v", configHolder)
	return
}
