package postgresdriver

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestPostgresDriver_ReadUsers(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id"}).
		AddRow("6025be31e1261e00308bfa3a").
		AddRow("6025be31e1261e00308bfa33")

	mock.ExpectQuery("^SELECT (.+) FROM users$").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db)

	users, err := driver.ReadUsers()
	c.NoError(err)
	c.Len(users, 2)

	mock.ExpectQuery("^SELECT (.+) FROM users$").WillReturnError(errors.New("dummy error"))

	users, err = driver.ReadUsers()
	c.EqualError(err, "dummy error")
	c.Empty(users)
}
