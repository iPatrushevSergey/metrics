// Package main provides the staticlint multichecker binary.
//
// The binary is built from cmd/staticlint and runs a combined set of analyzers:
//   - standard analyzers from golang.org/x/tools/go/analysis/passes;
//   - all SA analyzers from staticcheck (honnef.co/go/tools/staticcheck);
//   - at least one analyzer from other staticcheck classes (S and ST);
//   - custom analyzer exitcheck.
//
// exitcheck walks the whole module. It reports calls to os.Exit or
// log.Fatal / Fatalf / Fatalln everywhere except inside func main of
// package main, and reports the built-in panic everywhere with no exceptions.
//
// Run:
//
//	go run ./cmd/staticlint ./...
//
// or build once and run:
//
//	go build -o staticlint ./cmd/staticlint
//	./staticlint ./...
package main
