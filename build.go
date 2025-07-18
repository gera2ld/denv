package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type BuildAsset struct {
	Name        string
	Time        string
	DisplayTime string
	Size        int64
}

type BuildResult struct {
	Commit string
	Assets  []BuildAsset
}

const binDir = "bin"
const exeName = "denv"
const mainEntry = "./cmd/denv"

var args = [][]string{
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"windows", "amd64"},
	{"linux", "amd64"},
	{"linux", "arm64"},
}

func getCommit() string {
	cmd := exec.Command("git", "describe", "--always", "--tags")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Error getting commit: %s\n", err)
	}
	result := string(output)
	return strings.TrimSpace(result)
}

func getItem(name string) BuildAsset {
	fileInfo, err := os.Stat(binDir + "/" + name)
	if err != nil {
		log.Fatalf("Error reading file: %s\n", binDir+"/"+name)
	}
	return BuildAsset{
		Name:        name,
		Time:        fileInfo.ModTime().Format(time.RFC3339),
		DisplayTime: fileInfo.ModTime().Format(time.RFC1123),
		Size:        fileInfo.Size(),
	}
}

func buildItem(buildOs, buildArch, commit string) BuildAsset {
	suffix := ""
	if buildOs == "windows" {
		suffix = ".exe"
	}
	now := time.Now().Format(time.RFC3339)
	log.Printf("Build os=%s, arch=%s\n", buildOs, buildArch)
	name := exeName + "-" + buildOs + "-" + buildArch + suffix
	cmd := exec.Command(
		"go",
		"build",
		"-ldflags",
		"-s -w -X main.version="+commit+" -X main.builtAt="+now,
		"-trimpath",
		"-o",
		"bin/"+name,
		mainEntry,
	)
	env := os.Environ()
	cmd.Env = append(env, "CGO_ENABLED=0", "GOOS="+buildOs, "GOARCH="+buildArch)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed building os=%s, arch=%s\n", buildOs, buildArch)
	}
	return getItem(name)
}

func build() BuildResult {
	commit := getCommit()
	log.Printf("Build commit: %s\n", commit)
	result := BuildResult{
		Commit: commit,
	}
	for _, item := range args {
		buildOs := item[0]
		buildArch := item[1]
		item := buildItem(buildOs, buildArch, commit)
		result.Assets = append(result.Assets, item)
	}
	return result
}

func main() {
	build()
}
