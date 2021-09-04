package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Traceableai/goagent/filters/blocking/library"
)

type libraryInfo struct {
	Name    string
	OS      string
	Version string
}

var libtraceable = libraryInfo{
	Name:    "libtraceable",
	OS:      "ubuntu_18.04",
	Version: "0.1.94",
}

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

func main() {
	tmpFile, err := ioutil.TempFile(os.TempDir(), libtraceable.Name)
	if err != nil {
		log.Fatal("Failed to create temporary download file", err)
	}
	defer os.Remove(tmpFile.Name())

	file, err := os.Create(tmpFile.Name())
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	username, token := os.Getenv("TA_BASIC_AUTH_USER"), os.Getenv("TA_BASIC_AUTH_TOKEN")

	req, err := http.NewRequest(http.MethodGet, downloadURL(libtraceable), nil)
	if err != nil {
		log.Fatalf("Failed to build download request: %v", err)
	}
	req.SetBasicAuth(username, token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to download %q from %q: %v", libtraceable.Name, req.URL.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("failed to download %q from %q with status %d", libtraceable.Name, req.URL.String(), resp.StatusCode)
	}

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if size == 0 {
		log.Fatal("failed to copy the file: 0 bytes found.")
	}

	dstFolder := library.LDLibraryPath
	if len(os.Args) > 1 {
		dstFolder = os.Args[1]
	}

	count, err := unzip(tmpFile.Name(), "traceable", dstFolder)
	if count == 0 {
		log.Fatal("zero files were found in the zip file.")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func unzip(zipFile, subdir, dstFolder string) (int, error) {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	if subdir != "" {
		subdir = path.Clean(subdir) + "/"
	}

	count := 0

	var dstPath string
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		if subdir == "" {
			dstPath = f.Name
		} else if strings.HasPrefix(f.Name, subdir) {
			dstPath = f.Name[len(subdir):]
		} else {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return 0, fmt.Errorf("failed to open a zipped file: %v", err)
		}
		defer rc.Close()

		fpath := filepath.Join(dstFolder, dstPath)
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

		count++
	}

	return count, nil
}
