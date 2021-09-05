package main

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

type libraryInfo struct {
	Name    string
	OS      string
	Version string
}

//go:embed version.txt
var libVersion string

var downloadURLTmpl = template.Must(template.New("download_url").Parse(
	"https://traceableai.jfrog.io/artifactory/gradle-local/ai/traceable/agent/" +
		"{{ .Name}}_{{ .OS}}_x86_64/{{ .Version}}/{{ .Name}}_{{ .OS}}_x86_64-{{ .Version}}.zip",
))

func downloadURL(info libraryInfo) string {
	ub := new(bytes.Buffer)
	if err := downloadURLTmpl.Execute(ub, info); err != nil {
		log.Fatalf("failed to generate download URL: %v", err)
	}
	return ub.String()
}

func writeStringf(out io.Writer, msg string, args ...interface{}) {
	out.Write([]byte(fmt.Sprintf(msg, args...) + "\n"))
}

func main() {
	cmd := "help"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	var (
		out        = os.Stdout
		cmdArgs    = os.Args[2:]
		statusCode = 0
	)

	switch cmd {
	case "help":
		statusCode = helpCmd(out, cmdArgs)
	case "pull-library":
		statusCode = pullLibraryCmd(out, cmdArgs)
	case "install-library":
		statusCode = installLibraryCmd(out, cmdArgs)
	default:
		out.WriteString(fmt.Sprintf("Unknown command %q.", cmd))
		statusCode = 1
	}

	os.Exit(statusCode)
}

func helpCmd(out io.Writer, args []string) int {
	writeStringf(out, `
Usage:

	%s <command> [args...]

The commands are:

	help              displays the help
	pull-library      pulls the library into the repository
	install-library   installs the library in the host

	`, os.Args[0])
	return 0
}

func pullLibraryCmd(out io.Writer, args []string) int {
	return downloadFile(out, libraryInfo{
		Name:    "libtraceable",
		OS:      "ubuntu_20.04",
		Version: libVersion,
	}, "traceable/blocking.h", "../../library")
}

func installLibraryCmd(out io.Writer, args []string) int {
	if len(args) != 2 {
		writeStringf(out, `
Usage: %s %s <os> <dst_folder>
			`, os.Args[0], os.Args[1])
		return 1
	}

	return downloadFile(out, libraryInfo{
		Name:    "libtraceable",
		OS:      args[0],
		Version: libVersion,
	}, "traceable/libtraceable.so", args[1])
}

func downloadFile(out io.Writer, lib libraryInfo, filepath, dstDir string) int {
	tmpFile, err := ioutil.TempFile(os.TempDir(), lib.Name)
	if err != nil {
		writeStringf(out, "Failed to create temporary download file: %v", err)
		return 1
	}
	defer os.Remove(tmpFile.Name())

	file, err := os.Create(tmpFile.Name())
	if err != nil {
		writeStringf(out, "Failed to open temporary download file: %v", err)
		return 1
	}
	defer file.Close()

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	username, token := os.Getenv("TA_BASIC_AUTH_USER"), os.Getenv("TA_BASIC_AUTH_TOKEN")

	req, err := http.NewRequest(http.MethodGet, downloadURL(lib), nil)
	if err != nil {
		writeStringf(out, "Failed to build download request: %v", err)
		return 1
	}
	req.SetBasicAuth(username, token)

	resp, err := client.Do(req)
	if err != nil {
		writeStringf(out, "Failed to download %q from %q: %v", lib.Name, req.URL.String(), err)
		return 1
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		writeStringf(out, "Failed to download %q from %q with status %d", lib.Name, req.URL.String(), resp.StatusCode)
		return 1
	}

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		writeStringf(out, "Failed to copy download file: %v", err)
		return 1
	}
	if size == 0 {
		writeStringf(out, "Failed to copy download file: zero byes copied")
		return 1
	}

	count, err := unzipFile(tmpFile.Name(), filepath, dstDir)
	if err != nil {
		writeStringf(out, "Failed to unzip downloaded file: %v", err)
		return 1
	}
	if count == 0 {
		writeStringf(out, "No files were found in the downloaded zip file.")
		return 1
	}

	return 0
}

func unzipFile(zipFile, haystackFile, dstFolder string) (int, error) {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		if f.Name != haystackFile {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return 0, fmt.Errorf("failed to open a zipped file: %v", err)
		}
		defer rc.Close()

		fpath := filepath.Join(dstFolder, path.Base(f.Name))
		os.MkdirAll(path.Dir(fpath), os.ModePerm)

		dstFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return 0, fmt.Errorf("failed to create a target unzipped file: %v", err)
		}

		_, err = io.Copy(dstFile, rc)
		if err != nil {
			return 0, err
		}
		dstFile.Close()

		return 1, nil
	}

	return 0, nil
}
