package main

import (
	"errors"
	"fmt"
	"os/user"
	"strings"

	"github.com/zcalusic/sysinfo"
)

var installDir = map[string]string{
	"ubuntu_18":   "ubuntu_18.04",
	"ubuntu_20":   "ubuntu_18.04",
	"debian_10":   "ubuntu_18.04",
	"debian_11":   "ubuntu_18.04",
	"centos_7":    "centos_7",
	"centos_8":    "centos_8",
	"alpine_3.9":  "alpine_3.9",
	"alpine_3.10": "alpine_3.9",
	"alpine_3.11": "alpine_3.9",
	"alpine_3.12": "alpine_3.9",
	"alpine_3.13": "alpine_3.9",
	"alpine_3.14": "alpine_3.9",
	"amzn_2":      "centos_7",
}

func getLinuxDistro() (string, error) {
	current, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to obtain current user: %v", err)
	}

	if current.Uid != "0" {
		return "", errors.New("requires superuser privileges")
	}

	var si sysinfo.SysInfo
	si.GetSysInfo()

	version := si.OS.Version
	if idx := strings.LastIndex(version, "."); idx != -1 {
		version = version[:idx]
	}

	distro := fmt.Sprintf("%s_%s", si.OS.Vendor, version)
	if s, ok := installDir[distro]; ok {
		return s, nil
	}

	return "", fmt.Errorf("unsupported linux distro: %s", distro)
}
