package main

import (
	"denv/internal/cli"
	"denv/internal/config"
	"denv/internal/env"
	"denv/internal/filehandler"
	"fmt"
	"os"
)

func main() {
	globalConfig := config.NewConfig()
	filehandler := filehandler.NewFileHandler(globalConfig.RootDir, globalConfig.Debug)
	userConfig := config.NewUserConfig(globalConfig, filehandler)
	envManager := env.NewDynamicEnv(globalConfig, userConfig, filehandler)
	rootCmd := cli.NewRootCommand(envManager)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
