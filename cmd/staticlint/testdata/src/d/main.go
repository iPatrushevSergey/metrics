package main

import "os"

// indirect call from main (main -> exit -> os.Exit): must NOT be reported.
func main() {
	exit(1)
}

func exit(code int) {
	os.Exit(code)
}
