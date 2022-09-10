package postgresdriver

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
