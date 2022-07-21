package postgresdriver

import (
	"database/sql"
	"time"
)

const (
	selectApplicationByID = "SELECT * FROM applications WHERE application_id = $1"
)

type dbApplication struct {
	ID            int            `db:"id"`
	ApplicationID string         `db:"application_id"`
	ContactEmail  sql.NullString `db:"contact_email"`
	CreatedAt     sql.NullTime   `db:"created_at"`
	Description   sql.NullString `db:"description"`
	Name          sql.NullString `db:"name"`
	Owner         sql.NullString `db:"owner"`
	UpdatedAt     sql.NullTime   `db:"updated_at"`
	URL           sql.NullString `db:"url"`
	UserID        sql.NullString `db:"user_id"`
}

func (a *dbApplication) toApplication() *Application {
	return &Application{
		ApplicationID: a.ApplicationID,
		ContactEmail:  a.ContactEmail.String,
		CreatedAt:     &a.CreatedAt.Time,
		Description:   a.Description.String,
		Name:          a.Name.String,
		Owner:         a.Owner.String,
		UpdatedAt:     &a.UpdatedAt.Time,
		URL:           a.URL.String,
		UserID:        a.UserID.String,
	}
}

// Application struct handler representing an Application
type Application struct {
	ApplicationID string
	ContactEmail  string
	CreatedAt     *time.Time
	Description   string
	Name          string
	Owner         string
	UpdatedAt     *time.Time
	URL           string
	UserID        string
}

// ReadApplicationByID returns application in the database with given id
func (d *PostgresDriver) ReadApplicationByID(id string) (*Application, error) {
	var dbApplication dbApplication

	err := d.Get(&dbApplication, selectApplicationByID, id)
	if err != nil {
		return nil, err
	}

	return dbApplication.toApplication(), nil
}
