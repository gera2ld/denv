package config

import (
	"log"
	"os"
	"path/filepath"
)

type ConfigType struct {
	RootDir    string
	Identities string
	DataDir    string
	EnvSuffix  string
	IndexFile  string
	ConfigFile string
	Debug      bool
}

func NewConfig() *ConfigType {
	rootDir := os.Getenv("DENV_ROOT")
	if rootDir == "" {
		rootDir = filepath.Join(os.Getenv("HOME"), ".config", "denv")
	}
	identities := os.Getenv("DENV_IDENTITIES")
	if identities == "" {
		identities = filepath.Join(os.Getenv("HOME"), ".keys", "identities")
	}
	debug := os.Getenv("DENV_DEBUG") == "true"
	if debug {
		log.Printf("rootDir: %s", rootDir)
		log.Printf("identities: %s", identities)
	}
	return &ConfigType{
		RootDir:    rootDir,
		Identities: identities,
		DataDir:    "env",
		EnvSuffix:  ".age",
		IndexFile:  "temp/index.yml",
		ConfigFile: "config.yml",
		Debug:      debug,
	}
}
