package postgresdriver

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/stretchr/testify/require"
)

func TestPostgresDriver_ReadLoadbalancers(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	rows := sqlmock.NewRows([]string{"lb_id", "name", "created_at", "updated_at",
		"request_timeout", "gigastake", "gigastake_redirect", "user_id", "app_ids"}).
		AddRow("60e517ea76cfec00352bcdad", "wawawa", time.Now(), time.Now(),
			2100, true, true, "6107ef92825e090034dce25e", "{$oid:6107ef92825e090034dce25f}")

	mock.ExpectQuery("^SELECT (.+) FROM loadbalancers (.+)").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db, &ListenerMock{})

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

	driver := NewPostgresDriverFromSQLDBInstance(db, &ListenerMock{})

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into loadbalancers").WithArgs(sqlmock.AnyArg(),
		"yes", "60e85042bf95f5003559b791", sql.NullInt32{}, false, false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into stickiness_options").WithArgs(sqlmock.AnyArg(),
		"21", 21, true, pq.StringArray([]string{"pjog"})).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into lb_apps").WithArgs(sqlmock.AnyArg(), "61eae7640ae317bbc6c36dbb").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into lb_apps").WithArgs(sqlmock.AnyArg(), "61eae7640ae317bbc6c36dba").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	loadBalancer, err := driver.WriteLoadBalancer(&repository.LoadBalancer{
		ID:             "60ddc61b6e29c3003378361D",
		Name:           "yes",
		UserID:         "60e85042bf95f5003559b791",
		ApplicationIDs: []string{"61eae7640ae317bbc6c36dbb", "61eae7640ae317bbc6c36dba"},
		StickyOptions: repository.StickyOptions{
			Duration:      "21",
			StickyOrigins: []string{"pjog"},
			StickyMax:     21,
			Stickiness:    true,
		},
	})
	c.NoError(err)
	c.NotEmpty(loadBalancer)

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into loadbalancers").WithArgs(sqlmock.AnyArg(),
		"yes", "60e85042bf95f5003559b791", sql.NullInt32{}, false, false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("error in loadbalancers"))

	loadBalancer, err = driver.WriteLoadBalancer(&repository.LoadBalancer{
		ID:             "60ddc61b6e29c3003378361D",
		Name:           "yes",
		UserID:         "60e85042bf95f5003559b791",
		ApplicationIDs: []string{"61eae7640ae317bbc6c36dbb", "61eae7640ae317bbc6c36dba"},
	})
	c.EqualError(err, "error in loadbalancers")
	c.Empty(loadBalancer)

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into loadbalancers").WithArgs(sqlmock.AnyArg(),
		"yes", "60e85042bf95f5003559b791", sql.NullInt32{}, false, false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into stickiness_options").WithArgs(sqlmock.AnyArg(),
		"21", 21, true, pq.StringArray([]string{"pjog"})).
		WillReturnError(errors.New("error in stickiness options"))

	loadBalancer, err = driver.WriteLoadBalancer(&repository.LoadBalancer{
		ID:             "60ddc61b6e29c3003378361D",
		Name:           "yes",
		UserID:         "60e85042bf95f5003559b791",
		ApplicationIDs: []string{"61eae7640ae317bbc6c36dbb", "61eae7640ae317bbc6c36dba"},
		StickyOptions: repository.StickyOptions{
			Duration:      "21",
			StickyOrigins: []string{"pjog"},
			StickyMax:     21,
			Stickiness:    true,
		},
	})
	c.EqualError(err, "error in stickiness options")
	c.Empty(loadBalancer)

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into loadbalancers").WithArgs(sqlmock.AnyArg(),
		"yes", "60e85042bf95f5003559b791", sql.NullInt32{}, false, false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into lb_apps").WithArgs(sqlmock.AnyArg(), "61eae7640ae317bbc6c36dbb").
		WillReturnError(errors.New("error in lb_apps"))

	loadBalancer, err = driver.WriteLoadBalancer(&repository.LoadBalancer{
		ID:             "60ddc61b6e29c3003378361D",
		Name:           "yes",
		UserID:         "60e85042bf95f5003559b791",
		ApplicationIDs: []string{"61eae7640ae317bbc6c36dbb", "61eae7640ae317bbc6c36dba"},
	})
	c.EqualError(err, "error in lb_apps")
	c.Empty(loadBalancer)
}

func TestPostgresDriver_UpdateLoadBalancer(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	driver := NewPostgresDriverFromSQLDBInstance(db, &ListenerMock{})

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE loadbalancers").WithArgs("rochy", sqlmock.AnyArg(),
		"60ddc61b6e29c3003378361D").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM stickiness_options (.+)").WillReturnRows(sqlmock.NewRows([]string{"lb_id", "duration", "sticky_max", "stickiness", "origins"}).
		AddRow("60ddc61b6e29c3003378361D",
			"22", 22, true, pq.StringArray([]string{"rvn"})))

	mock.ExpectExec("UPDATE stickiness_options").WithArgs("21", 21, true,
		pq.StringArray([]string{"pjog"}), "60ddc61b6e29c3003378361D").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = driver.UpdateLoadBalancer("60ddc61b6e29c3003378361D", &repository.UpdateLoadBalancer{
		Name: "rochy",
		StickyOptions: &repository.StickyOptions{
			Duration:      "21",
			StickyOrigins: []string{"pjog"},
			StickyMax:     21,
			Stickiness:    true,
		},
	})
	c.NoError(err)

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE loadbalancers").WithArgs("rochy", sqlmock.AnyArg(),
		"60ddc61b6e29c3003378361D").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM stickiness_options (.+)").WillReturnRows(sqlmock.NewRows(nil))

	mock.ExpectExec("INSERT into stickiness_options").WithArgs("60ddc61b6e29c3003378361D",
		"21", 21, true, pq.StringArray([]string{"pjog"})).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = driver.UpdateLoadBalancer("60ddc61b6e29c3003378361D", &repository.UpdateLoadBalancer{
		Name: "rochy",
		StickyOptions: &repository.StickyOptions{
			Duration:      "21",
			StickyOrigins: []string{"pjog"},
			StickyMax:     21,
			Stickiness:    true,
		},
	})
	c.NoError(err)

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE loadbalancers").WithArgs("rochy", sqlmock.AnyArg(),
		"60ddc61b6e29c3003378361D").
		WillReturnError(errors.New("error load balancers"))

	err = driver.UpdateLoadBalancer("60ddc61b6e29c3003378361D", &repository.UpdateLoadBalancer{
		Name: "rochy",
	})
	c.EqualError(err, "error load balancers")

	mock.ExpectBegin()

	mock.ExpectExec("UPDATE loadbalancers").WithArgs("rochy", sqlmock.AnyArg(),
		"60ddc61b6e29c3003378361D").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("^SELECT (.+) FROM stickiness_options (.+)").WillReturnError(errors.New("error reading options"))

	err = driver.UpdateLoadBalancer("60ddc61b6e29c3003378361D", &repository.UpdateLoadBalancer{
		Name: "rochy",
		StickyOptions: &repository.StickyOptions{
			Duration:      "21",
			StickyOrigins: []string{"pjog"},
			StickyMax:     21,
			Stickiness:    true,
		},
	})
	c.EqualError(err, "error reading options")

	err = driver.UpdateLoadBalancer("60ddc61b6e29c3003378361D", nil)
	c.Equal(ErrNoFieldsToUpdate, err)

	err = driver.UpdateLoadBalancer("", nil)
	c.Equal(ErrMissingID, err)
}

func TestPostgresDriver_RemoveLoadBalancer(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	driver := NewPostgresDriverFromSQLDBInstance(db, &ListenerMock{})

	mock.ExpectExec("UPDATE loadbalancers").WithArgs(sqlmock.AnyArg(), "60ddc61b6e29c3003378361D").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = driver.RemoveLoadBalancer("60ddc61b6e29c3003378361D")
	c.NoError(err)

	mock.ExpectExec("UPDATE loadbalancers").WithArgs(sqlmock.AnyArg(), "not-an-id").
		WillReturnError(errors.New("dummy error"))

	err = driver.RemoveLoadBalancer("not-an-id")
	c.EqualError(err, "dummy error")

	err = driver.RemoveLoadBalancer("")
	c.Equal(ErrMissingID, err)
}
