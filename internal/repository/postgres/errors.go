package postgres

import (
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrorClassification type for error classification
type ErrorClassification int

const (
	// NonRetriable - the operation should not be repeated
	NonRetriable ErrorClassification = iota

	// Retriable - the operation can be repeated
	Retriable
)

// String returns a string representation of the classification
func (c ErrorClassification) String() string {
	switch c {
	case Retriable:
		return "Retriable"
	case NonRetriable:
		return "NonRetriable"
	default:
		return "Unknown"
	}
}

// PostgresErrorClassifier PostgreSQL Error Classifier
type PostgresErrorClassifier struct{}

// NewPostgresErrorClassifier creates a new instance of the error classifier
func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

// Classify classifies the error and returns an ErrorClassification
func (c *PostgresErrorClassifier) Classify(err error) ErrorClassification {
	if err == nil {
		return NonRetriable
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return c.classifyPgError(pgErr)
	}

	return NonRetriable
}

// Code Reference: https://www.postgresql.org/docs/current/errcodes-appendix.html
func (c *PostgresErrorClassifier) classifyPgError(pgErr *pgconn.PgError) ErrorClassification {
	code := pgErr.Code

	// Class 08 - Connection Exception (транспортные ошибки)
	if pgerrcode.IsConnectionException(code) {
		return Retriable
	}

	// Class 25 - Invalid Transaction State
	if pgerrcode.IsInvalidTransactionState(code) {
		return Retriable
	}

	// Class 40 - Transaction Rollback
	if pgerrcode.IsTransactionRollback(code) {
		return Retriable
	}

	// Class 53 - Insufficient Resources
	if pgerrcode.IsInsufficientResources(code) {
		return Retriable
	}

	// Class 57 - Operator Intervention (только определенные коды)
	if pgerrcode.IsOperatorIntervention(code) {
		switch code {
		case pgerrcode.CannotConnectNow, // 57P03
			pgerrcode.DatabaseDropped,                 // 57P04
			pgerrcode.IdleInTransactionSessionTimeout: // 57P05
			return Retriable
		}
	}

	return NonRetriable
}

// IsRetriable reports whether the error is classified as retriable.
func (c *PostgresErrorClassifier) IsRetriable(err error) bool {
	return c.Classify(err) == Retriable
}

// IsNonRetriable reports whether the error is classified as non-retriable.
func (c *PostgresErrorClassifier) IsNonRetriable(err error) bool {
	return c.Classify(err) == NonRetriable
}

// GetErrorCode returns the PostgreSQL error code from err, or an error if err is not a PgError.
func (c *PostgresErrorClassifier) GetErrorCode(err error) (string, error) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code, nil
	}
	return "", fmt.Errorf("error is not a PostgreSQL error: %w", err)
}
