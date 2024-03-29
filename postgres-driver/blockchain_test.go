package postgresdriver

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/stretchr/testify/require"
)

func TestPostgresDriver_ReadBlockchains(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	rows := sqlmock.NewRows([]string{"blockchain_id", "altruist", "blockchain", "blockchain_aliases", "chain_id", "chain_id_check",
		"description", "log_limit_blocks", "network", "path", "request_timeout", "ticker"}).
		AddRow("0021", "https://dummy.com:18546", "eth-mainnet", pq.StringArray{"eth-mainnet"}, 21, sql.NullString{},
			"Ethereum Mainnet", 212121, "ETH-1", sql.NullString{}, sql.NullString{}, "ETH").
		AddRow("0021", "https://dummy.com:18546", "eth-mainnet", pq.StringArray{"eth-mainnet"}, 21, sql.NullString{},
			"Ethereum Mainnet", 212121, "ETH-1", sql.NullString{}, sql.NullString{}, "ETH")

	mock.ExpectQuery("^SELECT (.+) FROM blockchains (.+)").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db, &ListenerMock{})

	blockchains, err := driver.ReadBlockchains()
	c.NoError(err)
	c.Len(blockchains, 2)

	mock.ExpectQuery("^SELECT (.+) FROM blockchains (.+)").WillReturnError(errors.New("dummy error"))

	blockchains, err = driver.ReadBlockchains()
	c.EqualError(err, "dummy error")
	c.Empty(blockchains)
}

func TestPostgresDriver_WriteBlockchain(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	driver := NewPostgresDriverFromSQLDBInstance(db, &ListenerMock{})

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into blockchains").WithArgs(
		"0062", true, "https://testaltruist.com", "ethereum-mainnet", pq.StringArray([]string{"ethereum-mainnet-high-gas"}), "34", `{\""method\"":\""eth_chainId\"",\""id\"":1,\""jsonrpc\"":\""2.0\""`, "Ethereum Mainnet", "JSON", 10000, "ETH-34", "/etc/", 10, "ETH", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into sync_check_options").WithArgs(
		"0062", "synccheck", 3, `{\""method\"":\""eth_blockNumber\"",\""id\"":1,\""jsonrpc\"":\""2.0\""}`, "/ext/bc/C/rpc", "result").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	blockchainToSend := &repository.Blockchain{
		ID:                "0062",
		Altruist:          "https://testaltruist.com",
		Blockchain:        "ethereum-mainnet",
		ChainID:           "34",
		ChainIDCheck:      `{\""method\"":\""eth_chainId\"",\""id\"":1,\""jsonrpc\"":\""2.0\""`,
		Description:       "Ethereum Mainnet",
		EnforceResult:     "JSON",
		Network:           "ETH-34",
		Path:              "/etc/",
		SyncCheck:         "synccheck",
		Ticker:            "ETH",
		BlockchainAliases: []string{"ethereum-mainnet-high-gas"},
		RequestTimeout:    10,
		LogLimitBlocks:    10000,
		Active:            true,
		SyncCheckOptions: repository.SyncCheckOptions{
			BlockchainID: "0062",
			Body:         `{\""method\"":\""eth_blockNumber\"",\""id\"":1,\""jsonrpc\"":\""2.0\""}`,
			ResultKey:    "result",
			Path:         "/ext/bc/C/rpc",
			Allowance:    3,
		},
	}

	blockchain, err := driver.WriteBlockchain(blockchainToSend)
	c.NoError(err)
	c.NotEmpty(blockchain.ID)

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into blockchains").WithArgs(
		"0062", true, "https://testaltruist.com", "ethereum-mainnet", pq.StringArray([]string{"ethereum-mainnet-high-gas"}), "34", `{\""method\"":\""eth_chainId\"",\""id\"":1,\""jsonrpc\"":\""2.0\""`, "Ethereum Mainnet", "JSON", 10000, "ETH-34", "/etc/", 10, "ETH", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("error in blockchains"))

	blockchain, err = driver.WriteBlockchain(blockchainToSend)
	c.EqualError(err, "error in blockchains")
	c.Empty(blockchain)

	mock.ExpectBegin()

	mock.ExpectExec("INSERT into blockchains").WithArgs(
		"0062", true, "https://testaltruist.com", "ethereum-mainnet", pq.StringArray([]string{"ethereum-mainnet-high-gas"}), "34", `{\""method\"":\""eth_chainId\"",\""id\"":1,\""jsonrpc\"":\""2.0\""`, "Ethereum Mainnet", "JSON", 10000, "ETH-34", "/etc/", 10, "ETH", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT into sync_check_options").WithArgs(
		"0062", "synccheck", 3, `{\""method\"":\""eth_blockNumber\"",\""id\"":1,\""jsonrpc\"":\""2.0\""}`, "/ext/bc/C/rpc", "result").
		WillReturnError(errors.New("error in sync_check_options"))

	blockchain, err = driver.WriteBlockchain(blockchainToSend)
	c.EqualError(err, "error in sync_check_options")
	c.Empty(blockchain)
}
