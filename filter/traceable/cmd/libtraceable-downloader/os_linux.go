package main

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/zcalusic/sysinfo"
)

type supportedVersions struct {
	minVersion    string
	targetLibrary string
}

var resolveLib = map[string]supportedVersions{
	"ubuntu": {"18", "ubuntu_18.04"},
	"debian": {"10", "ubuntu_18.04"},
	"centos": {"7", "centos_7"},
	"alpine": {"3.9", "alpine_3.9"},
	"amzn":   {"2", "centos_7"},
}

func getLinuxDistro() (string, error) {
	var si sysinfo.SysInfo
	si.GetSysInfo()

	if s, ok := resolveLib[si.OS.Vendor]; ok {
		mv, err := version.NewVersion(s.minVersion)
		if err != nil {
			return "", errors.New("failed to parse min version")
		}

		cv, err := version.NewVersion(si.OS.Version)
		if err != nil {
			return "", errors.New("failed to parse current version")
		}

		if mv.LessThanOrEqual(cv) {
			return s.targetLibrary, nil
		}
	}

	return "", fmt.Errorf("unsupported linux distro: %s_%s", si.OS.Vendor, si.OS.Version)
}
