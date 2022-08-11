package postgresdriver

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	// ErrNoFieldsToUpdate error when there are no fields to update
	ErrNoFieldsToUpdate = errors.New("no fields to update")
	// ErrMissingID error when ID is missing
	ErrMissingID = errors.New("missing id")

	idLength = 24
)

// PostgresDriver struct handler for PostgresDB related functions
type PostgresDriver struct {
	*sqlx.DB
}

// NewPostgresDriverFromConnectionString returns PostgresDriver instance from connection string
func NewPostgresDriverFromConnectionString(connectionString string) (*PostgresDriver, error) {
	db, err := sqlx.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	return &PostgresDriver{
		DB: db,
	}, nil
}

// NewPostgresDriverFromSQLDBInstance returns PostgresDriver instance from sdl.DB instance
// mostly used for mocking tests
func NewPostgresDriverFromSQLDBInstance(db *sql.DB) *PostgresDriver {
	return &PostgresDriver{
		DB: sqlx.NewDb(db, "postgres"),
	}
}

func newSQLNullString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}

	return sql.NullString{
		String: value,
		Valid:  true,
	}
}

func newSQLNullInt32(value int32) sql.NullInt32 {
	if value == 0 {
		return sql.NullInt32{}
	}

	return sql.NullInt32{
		Int32: value,
		Valid: true,
	}
}

func newSQLNullInt64(value int64) sql.NullInt64 {
	if value == 0 {
		return sql.NullInt64{}
	}

	return sql.NullInt64{
		Int64: value,
		Valid: true,
	}
}

func generateRandomID() (string, error) {
	bytes := make([]byte, idLength/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
