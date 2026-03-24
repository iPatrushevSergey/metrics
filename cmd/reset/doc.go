// Package main is the reset code generator binary.
//
// It scans Go packages under a module root (same layout as `go list ./...` from
// that directory using golang.org/x/tools/go/packages), finds struct types
// annotated with "// generate:reset" (in GenDecl or TypeSpec documentation),
// and writes reset.gen.go per package with Reset() methods on pointer receivers.
//
// Reset semantics (generated code):
//   - basic types: zero value;
//   - slices: expr = expr[:0] (capacity preserved);
//   - maps: clear(expr);
//   - pointers: if non-nil, reset pointed value following the same rules;
//   - struct values: call Reset() if defined on *T with signature func() (no args/results),
//     otherwise reset fields recursively;
//   - interface{} (empty): if value implements Reset(), call it via type assertion.
//
// Whole packages may be skipped when their path (relative to -root) contains
// certain segments (profiles, pkg, migrations, api) or a hidden directory
// (name longer than one rune and starting with '.').
//
// Regenerate fixture used by tests:
//
//	go generate ./cmd/reset
//
// Run on the module (from repo root):
//
//	go run ./cmd/reset -root .
//
// Run on a subtree only:
//
//	go run ./cmd/reset -root ./internal/model
//
// Build and run:
//
//	go build -o reset ./cmd/reset
//	./reset -root .
//
// Tests:
//
//	go test ./cmd/reset/...
//
//go:generate go run . -root ./internal/fixture
package main
