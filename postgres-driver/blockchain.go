package postgresdriver

import (
	"database/sql"

	"github.com/lib/pq"
)

const (
	selectAllBlockchainsScript = "SELECT * FROM blockchains"
	selectBlockchainByID       = "SELECT * FROM blockchains WHERE blockchain_id = $1"
)

type dbBlockchain struct {
	ID                int            `db:"id"`
	BlockchainID      string         `db:"blockchain_id"`
	Altruist          sql.NullString `db:"altruist"`
	Blockchain        sql.NullString `db:"blockchain"`
	BlockchainAliases pq.StringArray `db:"blockchain_aliases"`
	ChainID           sql.NullString `db:"chain_id"`
	ChaindIDCheck     sql.NullString `db:"chain_id_check"`
	Description       sql.NullString `db:"description"`
	Index             sql.NullInt64  `db:"_index"`
	LogLimitBlocks    sql.NullInt64  `db:"log_limit_blocks"`
	Network           sql.NullString `db:"network"`
	NetworkID         sql.NullString `db:"network_id"`
	NodeCount         sql.NullInt64  `db:"node_count"`
	Path              sql.NullString `db:"_path"`
	RequestTimeout    sql.NullInt64  `db:"request_timeout"`
	Ticker            sql.NullString `db:"ticker"`
}

func (b *dbBlockchain) toBlockchain() *Blockchain {
	return &Blockchain{
		BlockchainID:      b.BlockchainID,
		Altruist:          b.Altruist.String,
		Blockchain:        b.Blockchain.String,
		BlockchainAliases: b.BlockchainAliases,
		ChainID:           b.ChainID.String,
		ChaindIDCheck:     b.ChaindIDCheck.String,
		Description:       b.Description.String,
		Index:             b.Index.Int64,
		LogLimitBlocks:    b.LogLimitBlocks.Int64,
		Network:           b.Network.String,
		NetworkID:         b.NetworkID.String,
		NodeCount:         b.NodeCount.Int64,
		Path:              b.Path.String,
		RequestTimeout:    b.RequestTimeout.Int64,
		Ticker:            b.Ticker.String,
	}
}

// Blockchain struct handler representing a Blockchain
type Blockchain struct {
	BlockchainID      string
	Altruist          string
	Blockchain        string
	BlockchainAliases []string
	ChainID           string
	ChaindIDCheck     string
	Description       string
	Index             int64
	LogLimitBlocks    int64
	Network           string
	NetworkID         string
	NodeCount         int64
	Path              string
	RequestTimeout    int64
	Ticker            string
}

// ReadBlockchains returns all blockchains on the database
func (d *PostgresDriver) ReadBlockchains() ([]*Blockchain, error) {
	var dbBlockchains []*dbBlockchain

	err := d.Select(&dbBlockchains, selectAllBlockchainsScript)
	if err != nil {
		return nil, err
	}

	var blockchains []*Blockchain

	for _, dbBlockchain := range dbBlockchains {
		blockchains = append(blockchains, dbBlockchain.toBlockchain())
	}

	return blockchains, nil
}

// ReadBlockchainByID returns blockchain in the database with given id
func (d *PostgresDriver) ReadBlockchainByID(id string) (*Blockchain, error) {
	var dbBlockchain dbBlockchain

	err := d.Get(&dbBlockchain, selectBlockchainByID, id)
	if err != nil {
		return nil, err
	}

	return dbBlockchain.toBlockchain(), nil
}
