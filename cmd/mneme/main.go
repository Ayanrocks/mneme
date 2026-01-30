package main

import (
	"mneme/internal/cli"
	"mneme/internal/logger"
	"os"
)

func main() {
	if err := cli.Execute(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
