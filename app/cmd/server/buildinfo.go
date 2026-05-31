package main

import "fmt"

// Set at link time, e.g.:
// go build -ldflags "-X main.buildVersion=1.0 -X main.buildDate=... -X main.buildCommit=..."
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildNA(buildVersion))
	fmt.Printf("Build date: %s\n", buildNA(buildDate))
	fmt.Printf("Build commit: %s\n", buildNA(buildCommit))
}

func buildNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
