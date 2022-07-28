package postgresdriver

import (
	"database/sql"

	"github.com/lib/pq"
	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	selectBlockchainsScript = "SELECT * FROM blockchains"
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

func (b *dbBlockchain) toBlockchain() *repository.Blockchain {
	return &repository.Blockchain{
		ID:                b.BlockchainID,
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
		Path:              b.Path.String,
		RequestTimeout:    b.RequestTimeout.Int64,
		Ticker:            b.Ticker.String,
	}
}

// ReadBlockchains returns all blockchains on the database
func (d *PostgresDriver) ReadBlockchains() ([]*repository.Blockchain, error) {
	var dbBlockchains []*dbBlockchain

	err := d.Select(&dbBlockchains, selectBlockchainsScript)
	if err != nil {
		return nil, err
	}

	var blockchains []*repository.Blockchain

	for _, dbBlockchain := range dbBlockchains {
		blockchains = append(blockchains, dbBlockchain.toBlockchain())
	}

	return blockchains, nil
}
