package postgresdriver

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/stretchr/testify/require"
)

func TestPostgresDriver_ReadRedirects(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	rows := sqlmock.NewRows([]string{"blockchain_id", "alias", "loadbalancer", "domain"}).
		AddRow("0021", "pokt-mainnet", "12345", "pokt-mainnet.gateway.network").
		AddRow("0021", "pokt-mainnet", "12345", "pokt-mainnet.gateway.network")

	mock.ExpectQuery("^SELECT (.+) FROM redirects$").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db)

	blockchains, err := driver.ReadRedirects()
	c.NoError(err)
	c.Len(blockchains, 2)

	mock.ExpectQuery("^SELECT (.+) FROM redirects$").WillReturnError(errors.New("dummy error"))

	blockchains, err = driver.ReadRedirects()
	c.EqualError(err, "dummy error")
	c.Empty(blockchains)
}

func TestPostgresDriver_WriteRedirect(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	driver := NewPostgresDriverFromSQLDBInstance(db)

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into redirects").WithArgs("0021", "pokt-mainnet", "12345",
		"pokt-mainnet.gateway.network", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	redirectToSend := repository.Redirect{
		ID:             "1",
		BlockchainID:   "0021",
		Alias:          "pokt-mainnet",
		LoadBalancerID: "12345",
		Domain:         "pokt-mainnet.gateway.network",
	}

	app, err := driver.WriteRedirect(redirectToSend)
	c.NoError(err)
	c.NotEmpty(app.ID)

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into redirects").WithArgs("0021", "pokt-mainnet", "12345",
		"pokt-mainnet.gateway.network", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("error in redirects"))

	app, err = driver.WriteRedirect(redirectToSend)
	c.EqualError(err, "error in redirects")
	c.Empty(app)
}
