package postgresdriver

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestPostgresDriver_ReadLoadbalancers(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	rows := sqlmock.NewRows([]string{"lb_id", "name", "created_at", "updated_at",
		"request_timeout", "gigastake", "gigastake_redirect", "user_id"}).
		AddRow("60e517ea76cfec00352bcdad", "wawawa", time.Now(), time.Now(),
			2100, true, true, "6107ef92825e090034dce25e")

	mock.ExpectQuery("^SELECT (.+) FROM loadbalancers (.+)").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db)

	loadbalancer, err := driver.ReadLoadBalancers()
	c.NoError(err)
	c.Len(loadbalancer, 1)

	mock.ExpectQuery("^SELECT (.+) FROM loadbalancers (.+)").WillReturnError(errors.New("dummy error"))

	loadbalancer, err = driver.ReadLoadBalancers()
	c.EqualError(err, "dummy error")
	c.Empty(loadbalancer)
}
