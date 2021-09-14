package main

import (
	"errors"
	"fmt"
	"os/user"

	"github.com/zcalusic/sysinfo"
)

var installDir = map[string]string{
	"ubuntu_18.04": "/usr/lib/x86_64-linux-gnu",
	"ubuntu_20.04": "/usr/lib/x86_64-linux-gnu",
	"centos_7":     "/usr/lib64",
	"centos_8":     "/usr/lib64",
}

func getLinuxDistroAndInstallDir() (string, string, error) {
	current, err := user.Current()
	if err != nil {
		return "", "", fmt.Errorf("failed to obtain current user: %v", err)
	}

	if current.Uid != "0" {
		return "", "", errors.New("requires superuser privileges")
	}

	var si sysinfo.SysInfo
	si.GetSysInfo()

	distro := fmt.Sprintf("%s_%s", si.OS.Vendor, si.OS.Version)
	if dir, ok := installDir[distro]; ok {
		return distro, dir, nil
	}

	return "", "", fmt.Errorf("unsupported linux distro: %s", distro)
}
