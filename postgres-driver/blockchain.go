package postgresdriver

import (
	"database/sql"

	"github.com/lib/pq"
	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	selectBlockchainsScript = `SELECT b.id, b.blockchain_id, b.altruist, b.blockchain, b.blockchain_aliases, b.chain_id, b.chain_id_check, b.description, 
	b.enforce_result, b._index, b.log_limit_blocks, b.network, b.network_id, b.node_count, b._path, b.request_timeOut, b.sync_allowance, b.ticker, b.active,
	s.syncCheck as s_syncCheck, s.opts_allowance as s_opts_allowance, s.opts_body as s_opts_body, s.opts_path as s_opts_path, s.opts_result_key as s_opts_result_key
	FROM blockchains as b
	LEFT JOIN sync_check_options AS s ON b.blockchain_id=s.blockchain_id`
)

type dbBlockchain struct {
	ID                int            `db:"id"`
	BlockchainID      string         `db:"blockchain_id"`
	Altruist          sql.NullString `db:"altruist"`
	Blockchain        sql.NullString `db:"blockchain"`
	BlockchainAliases pq.StringArray `db:"blockchain_aliases"`
	ChainID           sql.NullString `db:"chain_id"`
	ChainIDCheck      sql.NullString `db:"chain_id_check"`
	Description       sql.NullString `db:"description"`
	EnforceResult     sql.NullString `db:"enforce_result"`
	Index             sql.NullInt32  `db:"_index"`
	LogLimitBlocks    sql.NullInt32  `db:"log_limit_blocks"`
	Network           sql.NullString `db:"network"`
	NetworkID         sql.NullString `db:"network_id"`
	NodeCount         sql.NullInt32  `db:"node_count"`
	ChainPath         sql.NullString `db:"_path"`
	RequestTimeout    sql.NullInt32  `db:"request_timeOut"`
	SyncAllowance     sql.NullInt32  `db:"sync_allowance"` // Add to Script (?)
	Ticker            sql.NullString `db:"ticker"`
	Active            sql.NullBool   `db:"active"` // Add to Script
	SyncCheck         sql.NullString `db:"s_syncCheck"`
	Allowance         sql.NullInt32  `db:"s_opts_allowance"`
	Body              sql.NullString `db:"s_opts_body"`
	Path              sql.NullString `db:"s_opts_path"`
	ResultKey         sql.NullString `db:"s_opts_result_key"`
}

func (b *dbBlockchain) toBlockchain() *repository.Blockchain {
	return &repository.Blockchain{
		ID:                b.BlockchainID,
		Altruist:          b.Altruist.String,
		Blockchain:        b.Blockchain.String,
		ChainID:           b.ChainID.String,
		ChainIDCheck:      b.ChainIDCheck.String,
		Description:       b.Description.String,
		EnforceResult:     b.EnforceResult.String,
		Network:           b.Network.String,
		NetworkID:         b.NetworkID.String,
		Path:              b.ChainPath.String,
		SyncCheck:         b.SyncCheck.String,
		Ticker:            b.Ticker.String,
		BlockchainAliases: b.BlockchainAliases,
		Index:             int(b.Index.Int32),
		LogLimitBlocks:    int(b.LogLimitBlocks.Int32),
		RequestTimeout:    int(b.RequestTimeout.Int32),
		SyncAllowance:     int(b.SyncAllowance.Int32),
		Active:            b.Active.Bool,
		SyncCheckOptions: repository.SyncCheckOptions{
			Body:      b.Body.String,
			ResultKey: b.ResultKey.String,
			Path:      b.Path.String,
			Allowance: int(b.Allowance.Int32),
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
