package main

import "os"

// direct os.Exit in package main func main: must be reported.
func main() {
	os.Exit(1) // want "direct os.Exit call in main is forbidden"
}
