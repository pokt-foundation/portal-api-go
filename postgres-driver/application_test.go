package postgresdriver

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestPostgresDriver_ReadApplicationByID(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "application_id", "contact_email", "created_at", "description",
		"name", "owner", "updated_at", "url", "user_id"}).
		AddRow(1, "5f62b7d8be3591c4dea85661", "dummy@ocampoent.com", time.Now(), "Wawawa gateway",
			"Wawawa", "ohana", time.Now(), "https://dummy.com", "6068da279aab4900333ec6dd")

	mock.ExpectQuery("^SELECT (.+) FROM applications (.+)").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db)

	application, err := driver.ReadApplicationByID("5f62b7d8be3591c4dea85661")
	c.NoError(err)
	c.NotEmpty(application)

	mock.ExpectQuery("^SELECT (.+) FROM applications (.+)").WillReturnError(errors.New("dummy error"))

	application, err = driver.ReadApplicationByID("5f62b7d8be3591c4dea85661")
	c.EqualError(err, "dummy error")
	c.Empty(application)
}
