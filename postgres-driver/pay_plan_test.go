package postgresdriver

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestPostgresDriver_ReadPayPlans(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	rows := sqlmock.NewRows([]string{"plan_type", "daily_limit"}).
		AddRow("FREETIER_V0", 250000).
		AddRow("PAY_AS_YOU_GO_V0", 0)

	mock.ExpectQuery("^SELECT (.+) FROM pay_plans$").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db, &ListenerMock{})

	payPlans, err := driver.ReadPayPlans()
	c.NoError(err)
	c.Len(payPlans, 2)

	mock.ExpectQuery("^SELECT (.+) FROM pay_plans$").WillReturnError(errors.New("dummy error"))

	payPlans, err = driver.ReadPayPlans()
	c.EqualError(err, "dummy error")
	c.Empty(payPlans)
}
