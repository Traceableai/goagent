package main

import (
	"errors"
	"fmt"
	"os/user"
	"strings"

	"github.com/zcalusic/sysinfo"
)

type installSettings struct {
	distro     string
	installDir string
}

var installDir = map[string]installSettings{
	"ubuntu_18":   installSettings{"ubuntu_18.04", "/usr/lib/x86_64-linux-gnu"},
	"ubuntu_20":   installSettings{"ubuntu_18.04", "/usr/lib/x86_64-linux-gnu"},
	"debian_10":   installSettings{"ubuntu_18.04", "/usr/lib/x86_64-linux-gnu"},
	"debian_11":   installSettings{"ubuntu_18.04", "/usr/lib/x86_64-linux-gnu"},
	"centos_7":    installSettings{"centos_7", "/usr/lib64"},
	"centos_8":    installSettings{"centos_8", "/usr/lib64"},
	"alpine_3.9":  installSettings{"alpine_3.9", "/usr/local/lib"},
	"alpine_3.10": installSettings{"alpine_3.9", "/usr/local/lib"},
	"alpine_3.11": installSettings{"alpine_3.9", "/usr/local/lib"},
	"alpine_3.12": installSettings{"alpine_3.9", "/usr/local/lib"},
	"alpine_3.13": installSettings{"alpine_3.9", "/usr/local/lib"},
	"alpine_3.14": installSettings{"alpine_3.9", "/usr/local/lib"},
	"amzn_2":      installSettings{"centos_7", "/usr/lib64"},
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

	version := si.OS.Version
	if idx := strings.LastIndex(version, "."); idx != -1 {
		version = version[:idx]
	}

	distro := fmt.Sprintf("%s_%s", si.OS.Vendor, version)
	if s, ok := installDir[distro]; ok {
		return s.distro, s.installDir, nil
	}

	return "", "", fmt.Errorf("unsupported linux distro: %s", distro)
}
