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
	Body              sql.NullString `db:"body"`
	ChainID           sql.NullString `db:"chain_id"`
	ChaindIDCheck     sql.NullString `db:"chain_id_check"`
	Description       sql.NullString `db:"description"`
	EnforceResult     sql.NullString `db:"enforce_result"`
	Network           sql.NullString `db:"network"`
	NetworkID         sql.NullString `db:"network_id"`
	Path              sql.NullString `db:"_path"`
	ResultKey         sql.NullString `db:"result_key"`
	SyncCheck         sql.NullString `db:"sync_check"`
	Ticker            sql.NullString `db:"ticker"`
	BlockchainAliases pq.StringArray `db:"blockchain_aliases"`
	Index             sql.NullInt64  `db:"_index"`
	LogLimitBlocks    sql.NullInt64  `db:"log_limit_blocks"`
	RequestTimeout    sql.NullInt64  `db:"request_timeout"`
	SyncAllowance     sql.NullInt64  `db:"sync_allowance"`
	NodeCount         sql.NullInt64  `db:"node_count"`
	Active            sql.NullBool   `db:"active"`
	Allowance         sql.NullInt64  `db:"active"`
}

func (b *dbBlockchain) toBlockchain() *repository.Blockchain {
	return &repository.Blockchain{
		ID:                b.BlockchainID,
		Altruist:          b.Altruist.String,
		Blockchain:        b.Blockchain.String,
		ChainID:           b.ChainID.String,
		ChaindIDCheck:     b.ChaindIDCheck.String,
		Description:       b.Description.String,
		EnforceResult:     b.EnforceResult.String,
		Network:           b.Network.String,
		NetworkID:         b.NetworkID.String,
		Path:              b.Path.String,
		SyncCheck:         b.SyncCheck.String,
		Ticker:            b.Ticker.String,
		BlockchainAliases: b.BlockchainAliases,
		Index:             b.Index.Int64,
		LogLimitBlocks:    b.LogLimitBlocks.Int64,
		RequestTimeout:    b.RequestTimeout.Int64,
		SyncAllowance:     b.SyncAllowance.Int64,
		Active:            b.Active.Bool,
		SyncCheckOptions: repository.SyncCheckOptions{
			Body:      b.Body.String,
			ResultKey: b.ResultKey.String,
			Path:      b.Path.String,
			Allowance: b.Allowance.Int64,
		},
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
