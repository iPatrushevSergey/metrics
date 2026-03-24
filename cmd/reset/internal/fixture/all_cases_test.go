package fixture

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type resetSpy struct {
	calls int
}

func (s *resetSpy) Reset() {
	s.calls++
}

func TestAllCases_Reset(t *testing.T) {
	n := 7
	tags := []string{"a", "b"}
	m := map[int]string{1: "one"}
	addr := addressData{ZipCode: 100}
	sess := sessionState{Attempts: 3}
	inner := 42
	pInner := &inner
	pp := &pInner

	spy := &resetSpy{}

	v := &AllCases{
		UserName:           "user",
		RequestCount:       10,
		IsEnabled:          true,
		RawByte:            9,
		UnicodeRune:        'я',
		AverageScore32:     1.5,
		AverageScore64:     2.5,
		AliasName:          "alias",
		IDsByOrder:         []int{1, 2, 3},
		ScoreByKey:         map[string]int{"k": 1},
		Address:            addressData{ZipCode: 200},
		Session:            sessionState{Attempts: 5},
		addressData:        addressData{ZipCode: 300},
		OptionalCount:      &n,
		OptionalTags:       &tags,
		OptionalCodeToName: &m,
		OptionalAddress:    &addr,
		OptionalSession:    &sess,
		OptionalPtrCount:   pp,
		Payload:            spy,
	}

	v.Reset()

	require.Equal(t, "", v.UserName)
	require.Equal(t, 0, v.RequestCount)
	require.False(t, v.IsEnabled)
	require.Equal(t, byte(0), v.RawByte)
	require.Equal(t, rune(0), v.UnicodeRune)
	require.Equal(t, float32(0), v.AverageScore32)
	require.Equal(t, float64(0), v.AverageScore64)
	require.Equal(t, MyString(""), v.AliasName)

	require.NotNil(t, v.IDsByOrder)
	require.Empty(t, v.IDsByOrder)
	require.Greater(t, cap(v.IDsByOrder), 0)

	require.NotNil(t, v.ScoreByKey)
	require.Empty(t, v.ScoreByKey)

	require.Zero(t, v.Address.ZipCode)
	require.Zero(t, v.Session.Attempts)
	require.Zero(t, v.addressData.ZipCode)

	require.NotNil(t, v.OptionalCount)
	require.Zero(t, *v.OptionalCount)

	require.NotNil(t, v.OptionalTags)
	require.Empty(t, *v.OptionalTags)

	require.NotNil(t, v.OptionalCodeToName)
	require.Empty(t, *v.OptionalCodeToName)

	require.NotNil(t, v.OptionalAddress)
	require.Zero(t, v.OptionalAddress.ZipCode)

	require.NotNil(t, v.OptionalSession)
	require.Zero(t, v.OptionalSession.Attempts)

	require.NotNil(t, v.OptionalPtrCount)
	require.NotNil(t, *v.OptionalPtrCount)
	require.Zero(t, **v.OptionalPtrCount)

	require.Equal(t, 1, spy.calls)
}
