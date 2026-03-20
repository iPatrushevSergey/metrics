package main

import o "os"

// direct os.Exit via import alias in package main func main: must be reported.
func main() {
	o.Exit(2) // want "direct os.Exit call in main is forbidden"
}
