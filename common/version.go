package common

import (
	"fmt"
)

var (
	Version   string
	Commit    string
	BuildTime string
)

func PrintVersion() {
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Commit: %s\n", Commit)
	fmt.Printf("BuildTime: %s\n", BuildTime)
}
