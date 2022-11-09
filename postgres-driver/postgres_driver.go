package postgresdriver

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	psqlDateLayout = "2006-01-02T15:04:05.999999"
)

var (
	// ErrNoFieldsToUpdate error when there are no fields to update
	// TODO - remove once moved to update structs IsValid methods
	ErrNoFieldsToUpdate = errors.New("no fields to update")

	// ErrMissingID error when ID is missing
	ErrMissingID = errors.New("missing id")

	idLength = 24
)

// PostgresDriver struct handler for PostgresDB related functions
type PostgresDriver struct {
	notification chan *repository.Notification
	listener     Listener
	*sqlx.DB
}

// NewPostgresDriverFromConnectionString returns PostgresDriver instance from connection string
func NewPostgresDriverFromConnectionString(connectionString string, listener Listener) (*PostgresDriver, error) {
	db, err := sqlx.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	driver := &PostgresDriver{
		notification: make(chan *repository.Notification, 32),
		listener:     listener,
		DB:           db,
	}

	err = driver.listener.Listen("events")
	if err != nil {
		return nil, err
	}

	go Listen(driver.listener.NotificationChannel(), driver.notification)

	return driver, nil
}

// NewPostgresDriverFromSQLDBInstance returns PostgresDriver instance from sdl.DB instance
// mostly used for mocking tests
func NewPostgresDriverFromSQLDBInstance(db *sql.DB, listener Listener) *PostgresDriver {
	driver := &PostgresDriver{
		notification: make(chan *repository.Notification, 32),
		listener:     listener,
		DB:           sqlx.NewDb(db, "postgres"),
	}

	err := driver.listener.Listen("events")
	if err != nil {
		panic(err)
	}

	go Listen(driver.listener.NotificationChannel(), driver.notification)

	return driver
}

// NotificationChannel returns just receiver Notification channel
func (d *PostgresDriver) NotificationChannel() <-chan *repository.Notification {
	return d.notification
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

func newSQLNullTime(value time.Time) sql.NullTime {
	if value.IsZero() {
		return sql.NullTime{}
	}

	return sql.NullTime{
		Time:  value,
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

type nullable interface {
	isNotNull() bool
}

type updatable interface {
	isUpdatable() bool
	read(appID string, driver *PostgresDriver) (updatable, error)
}

type update struct {
	insertScript string
	updateScript string
	toUpdate     updatable
}

func (d *PostgresDriver) doUpdate(id string, update *update, tx *sqlx.Tx) error {
	if !update.toUpdate.isUpdatable() {
		return nil
	}

	_, err := update.toUpdate.read(id, d)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err = tx.NamedExec(update.insertScript, update.toUpdate)
			if err != nil {
				return err
			}

			return nil
		}

		return err
	}

	_, err = tx.NamedExec(update.updateScript, update.toUpdate)
	if err != nil {
		return err
	}

	return nil
}

func psqlDateToTime(rawDate string) time.Time {
	date, _ := time.Parse(psqlDateLayout, rawDate)
	return date
}
