package postgres

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorClassification_String(t *testing.T) {
	assert.Equal(t, "Retriable", Retriable.String())
	assert.Equal(t, "NonRetriable", NonRetriable.String())
	assert.Equal(t, "Unknown", ErrorClassification(99).String())
}

func TestNewPostgresErrorClassifier(t *testing.T) {
	c := NewPostgresErrorClassifier()
	require.NotNil(t, c)
}

func TestPostgresErrorClassifier_Classify_nil(t *testing.T) {
	c := NewPostgresErrorClassifier()
	assert.Equal(t, NonRetriable, c.Classify(nil))
}

func TestPostgresErrorClassifier_Classify_ordinaryError(t *testing.T) {
	c := NewPostgresErrorClassifier()
	err := errors.New("not pg")
	assert.Equal(t, NonRetriable, c.Classify(err))
}

func TestPostgresErrorClassifier_IsRetriable_IsNonRetriable(t *testing.T) {
	c := NewPostgresErrorClassifier()
	err := errors.New("x")
	assert.False(t, c.IsRetriable(err))
	assert.True(t, c.IsNonRetriable(err))
}

func TestPostgresErrorClassifier_GetErrorCode(t *testing.T) {
	c := NewPostgresErrorClassifier()
	_, err := c.GetErrorCode(errors.New("x"))
	assert.Error(t, err)
}

func TestPostgresErrorClassifier_Classify_pgError(t *testing.T) {
	c := NewPostgresErrorClassifier()
	// Connection exception — retriable
	pgErr := &pgconn.PgError{Code: "08006"}
	assert.Equal(t, Retriable, c.Classify(pgErr))
}

func TestPostgresErrorClassifier_GetErrorCode_success(t *testing.T) {
	c := NewPostgresErrorClassifier()
	pgErr := &pgconn.PgError{Code: "42P01"}
	code, err := c.GetErrorCode(pgErr)
	require.NoError(t, err)
	assert.Equal(t, "42P01", code)
}
