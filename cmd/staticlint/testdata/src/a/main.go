package main

import "os"

// func main of package main: os.Exit allowed here.
func main() {
	os.Exit(1)
	panic("unreachable") // want "forbidden: built-in panic"
}
