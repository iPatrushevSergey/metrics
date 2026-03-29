package main

import (
	"log"
	o "os"
)

// Helpers in package main are checked.
func die() {
	o.Exit(1) // want "forbidden: os.Exit or log.Fatal outside func main of package main"
}

func stop() {
	log.Fatal("stop") // want "forbidden: os.Exit or log.Fatal outside func main of package main"
}

func main() {
	die()
	stop()
}
