package c

import "os"

// package is not main, so os.Exit here must NOT be reported.
func main() {
	os.Exit(1)
}
