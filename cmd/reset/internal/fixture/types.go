// Package fixture holds sample types for cmd/reset generator tests.
package fixture

// MyString is a type alias to string.
type MyString = string

type addressData struct {
	ZipCode int
}

type sessionState struct {
	Attempts int
}

func (s *sessionState) Reset() {
	if s == nil {
		return
	}
	s.Attempts = 0
}

// AllCases is the fixture struct passed to the generator.
//
// generate:reset
type AllCases struct {
	UserName       string
	RequestCount   int
	IsEnabled      bool
	RawByte        byte
	UnicodeRune    rune
	AverageScore32 float32
	AverageScore64 float64
	AliasName      MyString

	IDsByOrder []int
	ScoreByKey map[string]int

	Address     addressData
	Session     sessionState
	addressData // embedded struct

	OptionalCount      *int
	OptionalTags       *[]string
	OptionalCodeToName *map[int]string
	OptionalAddress    *addressData
	OptionalSession    *sessionState
	OptionalPtrCount   **int

	Payload interface{}
}
