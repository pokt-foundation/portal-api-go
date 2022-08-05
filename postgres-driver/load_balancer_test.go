package postgresdriver

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pokt-foundation/portal-api-go/repository"
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

func TestPostgresDriver_WriteLoadBalancer(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	driver := NewPostgresDriverFromSQLDBInstance(db)

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into loadbalancers").WithArgs(sqlmock.AnyArg(),
		"yes", "60e85042bf95f5003559b791", sql.NullInt32{}, false, false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into lb_apps").WithArgs(sqlmock.AnyArg(), "61eae7640ae317bbc6c36dbb").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into lb_apps").WithArgs(sqlmock.AnyArg(), "61eae7640ae317bbc6c36dba").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = driver.WriteLoadBalancer(&repository.LoadBalancer{
		ID:             "60ddc61b6e29c3003378361D",
		Name:           "yes",
		UserID:         "60e85042bf95f5003559b791",
		ApplicationIDs: []string{"61eae7640ae317bbc6c36dbb", "61eae7640ae317bbc6c36dba"},
	})
	c.NoError(err)

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into loadbalancers").WithArgs(sqlmock.AnyArg(),
		"yes", "60e85042bf95f5003559b791", sql.NullInt32{}, false, false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("error in loadbalancers"))

	err = driver.WriteLoadBalancer(&repository.LoadBalancer{
		ID:             "60ddc61b6e29c3003378361D",
		Name:           "yes",
		UserID:         "60e85042bf95f5003559b791",
		ApplicationIDs: []string{"61eae7640ae317bbc6c36dbb", "61eae7640ae317bbc6c36dba"},
	})
	c.EqualError(err, "error in loadbalancers")

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into loadbalancers").WithArgs(sqlmock.AnyArg(),
		"yes", "60e85042bf95f5003559b791", sql.NullInt32{}, false, false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into lb_apps").WithArgs(sqlmock.AnyArg(), "61eae7640ae317bbc6c36dbb").
		WillReturnError(errors.New("error in lb_apps"))

	err = driver.WriteLoadBalancer(&repository.LoadBalancer{
		ID:             "60ddc61b6e29c3003378361D",
		Name:           "yes",
		UserID:         "60e85042bf95f5003559b791",
		ApplicationIDs: []string{"61eae7640ae317bbc6c36dbb", "61eae7640ae317bbc6c36dba"},
	})
	c.EqualError(err, "error in lb_apps")
}

func TestPostgresDriver_UpdateLoadBalancer(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	driver := NewPostgresDriverFromSQLDBInstance(db)

	mock.ExpectExec("UPDATE loadbalancers").WithArgs("rochy", "60e85042bf95f5003559b791", sqlmock.AnyArg(),
		"60ddc61b6e29c3003378361D").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = driver.UpdateLoadBalancer("60ddc61b6e29c3003378361D", &UpdateLoadBalancerOptions{
		Name:   "rochy",
		UserID: "60e85042bf95f5003559b791",
	})
	c.NoError(err)

	mock.ExpectExec("UPDATE loadbalancers").WithArgs("rochy", "60e85042bf95f5003559b791", sqlmock.AnyArg(),
		"60ddc61b6e29c3003378361D").
		WillReturnError(errors.New("dummy error"))

	err = driver.UpdateLoadBalancer("60ddc61b6e29c3003378361D", &UpdateLoadBalancerOptions{
		Name:   "rochy",
		UserID: "60e85042bf95f5003559b791",
	})
	c.EqualError(err, "dummy error")

	err = driver.UpdateLoadBalancer("60ddc61b6e29c3003378361D", nil)
	c.Equal(ErrNoFieldsToUpdate, err)
}
