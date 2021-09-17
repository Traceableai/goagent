package main

import (
	"fmt"

	"github.com/zcalusic/sysinfo"
)

type supportedVersions struct {
	minVersion    string
	targetLibrary string
}

var resolveLib = map[string]supportedVersions{
	"ubuntu": supportedVersions{"18", "ubuntu_18.04"},
	"debian": supportedVersions{"10", "ubuntu_18.04"},
	"centos": supportedVersions{"7", "centos_7"},
	"alpine": supportedVersions{"3.9", "alpine_3.9"},
	"amzn":   supportedVersions{"2", "centos_7"},
}

func getLinuxDistro() (string, error) {
	var si sysinfo.SysInfo
	si.GetSysInfo()

	if s, ok := resolveLib[si.OS.Vendor]; ok && checkMinVersion(s.minVersion, si.OS.Version) {
		return s.targetLibrary, nil
	}

	return "", fmt.Errorf("unsupported linux distro: %s_%s", si.OS.Vendor, si.OS.Version)
}

func checkMinVersion(minVersion string, version string) bool {
	return version > minVersion
}
