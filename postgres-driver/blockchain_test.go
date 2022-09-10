package postgresdriver

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestPostgresDriver_ReadBlockchains(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "blockchain_id", "altruist", "blockchain", "blockchain_aliases", "chain_id", "chain_id_check",
		"description", "_index", "log_limit_blocks", "network", "network_id", "node_count", "_path", "request_timeout", "ticker"}).
		AddRow(1, "0021", "https://dummy.com:18546", "eth-mainnet", pq.StringArray{"eth-mainnet"}, 21, sql.NullString{},
			"Ethereum Mainnet", 21, 212121, "ETH-1", sql.NullString{}, 21, sql.NullString{}, sql.NullString{}, "ETH").
		AddRow(2, "0021", "https://dummy.com:18546", "eth-mainnet", pq.StringArray{"eth-mainnet"}, 21, sql.NullString{},
			"Ethereum Mainnet", 21, 212121, "ETH-1", sql.NullString{}, 21, sql.NullString{}, sql.NullString{}, "ETH")

	mock.ExpectQuery("^SELECT (.+) FROM blockchains (.+)").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db)

	blockchains, err := driver.ReadBlockchains()
	c.NoError(err)
	c.Len(blockchains, 2)

	mock.ExpectQuery("^SELECT (.+) FROM blockchains (.+)").WillReturnError(errors.New("dummy error"))

	blockchains, err = driver.ReadBlockchains()
	c.EqualError(err, "dummy error")
	c.Empty(blockchains)
}
