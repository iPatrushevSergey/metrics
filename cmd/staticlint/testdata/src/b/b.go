package b

import "os"

// Not package main: func main here is a normal function, not exempt.
func main() {
	os.Exit(1) // want "forbidden: os.Exit or log.Fatal outside func main of package main"
	panic("x") // want "forbidden: built-in panic"
}
