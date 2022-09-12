package postgresdriver

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	selectBlockchainsScript = `SELECT b.id, b.blockchain_id, b.altruist, b.blockchain, b.blockchain_aliases, b.chain_id, b.chain_id_check, b.description, 
	b._index, b.log_limit_blocks, b.network, b.network_id, b.node_count, b._path, b.request_timeout, b.ticker, b.active,
	s.synccheck as s_sync_check, s.opts_allowance as s_opts_allowance, s.opts_body as s_opts_body, s.opts_path as s_opts_path, s.opts_result_key as s_opts_result_key
	FROM blockchains as b
	LEFT JOIN sync_check_options AS s ON b.blockchain_id=s.blockchain_id`
	insertBlockchainScript = `
	INSERT into blockchains (id, blockchain_id, altruist, blockchain, blockchain_aliases, chain_id, chain_id_check, description, _index, log_limit_blocks, network, network_id, node_count, _path, request_timeout, ticker, active, created_at, updated_at)
	VALUES (:id, :blockchain_id, :altruist, :blockchain, :blockchain_aliases, :chain_id, :chain_id_check, :description, :_index, :log_limit_blocks, :network, :network_id, :node_count, :_path, :request_timeout, :ticker, :active, :created_at, :updated_at)`
	insertSyncCheckOptionsScript = `
	INSERT into sync_check_options (blockchain_id, synccheck, opts_allowance, opts_body, opts_path, opts_result_key)
	VALUES (:blockchain_id, :synccheck, :opts_allowance, :opts_body, :opts_path, :opts_result_key)`
)

// TO DO - Below fields may or may not need to be added back pending the meeting to discuss Postgres migration DB fields.

type dbBlockchain struct {
	ID           int            `db:"id"`
	BlockchainID string         `db:"blockchain_id"`
	Altruist     sql.NullString `db:"altruist"`
	Blockchain   sql.NullString `db:"blockchain"`
	ChainID      sql.NullString `db:"chain_id"`
	ChainIDCheck sql.NullString `db:"chain_id_check"`
	ChainPath    sql.NullString `db:"_path"`
	Description  sql.NullString `db:"description"`
	Ticker       sql.NullString `db:"ticker"`
	// EnforceResult     sql.NullString `db:"enforce_result"` // Add to Migration Script (?)
	Network           sql.NullString `db:"network"`
	NetworkID         sql.NullString `db:"network_id"`
	BlockchainAliases pq.StringArray `db:"blockchain_aliases"`
	LogLimitBlocks    sql.NullInt32  `db:"log_limit_blocks"`
	Index             sql.NullInt32  `db:"_index"`
	NodeCount         sql.NullInt32  `db:"node_count"` // What is this?
	RequestTimeout    sql.NullInt32  `db:"request_timeout"`
	// SyncAllowance  sql.NullInt32  `db:"sync_allowance"` // Add to Migration Script (?)
	Active    sql.NullBool   `db:"active"` // Add to Migration Script
	SyncCheck sql.NullString `db:"s_sync_check"`
	Allowance sql.NullInt32  `db:"s_opts_allowance"`
	Body      sql.NullString `db:"s_opts_body"`
	Path      sql.NullString `db:"s_opts_path"`
	ResultKey sql.NullString `db:"s_opts_result_key"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

func (b *dbBlockchain) toBlockchain() *repository.Blockchain {
	return &repository.Blockchain{
		ID:           b.BlockchainID,
		Altruist:     b.Altruist.String,
		Blockchain:   b.Blockchain.String,
		ChainID:      b.ChainID.String,
		ChainIDCheck: b.ChainIDCheck.String,
		Description:  b.Description.String,
		// EnforceResult:     b.EnforceResult.String,
		Network:           b.Network.String,
		NetworkID:         b.NetworkID.String,
		Path:              b.ChainPath.String,
		SyncCheck:         b.SyncCheck.String,
		Ticker:            b.Ticker.String,
		BlockchainAliases: b.BlockchainAliases,
		Index:             int(b.Index.Int32),
		LogLimitBlocks:    int(b.LogLimitBlocks.Int32),
		RequestTimeout:    int(b.RequestTimeout.Int32),
		// SyncAllowance:     int(b.SyncAllowance.Int32),
		Active: b.Active.Bool,
		SyncCheckOptions: repository.SyncCheckOptions{
			Body:      b.Body.String,
			ResultKey: b.ResultKey.String,
			Path:      b.Path.String,
			Allowance: int(b.Allowance.Int32),
		},
	}
}

type insertDBBlockchain struct {
	ID           int            `db:"id"`
	BlockchainID string         `db:"blockchain_id"`
	Altruist     sql.NullString `db:"altruist"`
	Blockchain   sql.NullString `db:"blockchain"`
	ChainID      sql.NullString `db:"chain_id"`
	ChainIDCheck sql.NullString `db:"chain_id_check"`
	ChainPath    sql.NullString `db:"_path"`
	Description  sql.NullString `db:"description"`
	Ticker       sql.NullString `db:"ticker"`
	// EnforceResult     sql.NullString `db:"enforce_result"` // Add to Migration Script (?)
	Network           sql.NullString `db:"network"`
	NetworkID         sql.NullString `db:"network_id"`
	BlockchainAliases pq.StringArray `db:"blockchain_aliases"`
	LogLimitBlocks    sql.NullInt32  `db:"log_limit_blocks"`
	Index             sql.NullInt32  `db:"_index"`
	NodeCount         sql.NullInt32  `db:"node_count"`
	RequestTimeout    sql.NullInt32  `db:"request_timeout"`
	// SyncAllowance  sql.NullInt32  `db:"sync_allowance"` // Add to Migration Script (?)
	Active    bool      `db:"active"` // Add to Migration Script
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
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

func extractInsertDBBlockchain(blockchain *repository.Blockchain) *insertDBBlockchain {
	return &insertDBBlockchain{
		BlockchainID:      blockchain.ChainID,
		Altruist:          newSQLNullString(blockchain.Altruist),
		Blockchain:        newSQLNullString(blockchain.Blockchain),
		ChainID:           newSQLNullString(blockchain.ChainID),
		ChainIDCheck:      newSQLNullString(blockchain.ChainIDCheck),
		ChainPath:         newSQLNullString(blockchain.Path),
		Description:       newSQLNullString(blockchain.Description),
		Ticker:            newSQLNullString(blockchain.Ticker),
		Network:           newSQLNullString(blockchain.Network),
		NetworkID:         newSQLNullString(blockchain.NetworkID),
		BlockchainAliases: blockchain.BlockchainAliases,
		LogLimitBlocks:    newSQLNullInt32(int32(blockchain.LogLimitBlocks)),
		Index:             newSQLNullInt32(int32(blockchain.Index)),
		RequestTimeout:    newSQLNullInt32(int32(blockchain.RequestTimeout)),
		Active:            blockchain.Active,
	}
}

type insertSyncCheckOptions struct {
	BlockchainID string         `db:"blockchain_id"`
	SyncCheck    sql.NullString `db:"synccheck"`
	Body         sql.NullString `db:"opts_body"`
	Path         sql.NullString `db:"opts_path"`
	ResultKey    sql.NullString `db:"opts_result_key"`
	Allowance    sql.NullInt32  `db:"opts_allowance"`
}

func (i *insertSyncCheckOptions) isNotNull() bool {
	return i.SyncCheck.Valid || i.Body.Valid || i.Path.Valid || i.ResultKey.Valid || i.Allowance.Valid
}

func extractInsertSyncCheckOptions(blockchain *repository.Blockchain) *insertSyncCheckOptions {
	return &insertSyncCheckOptions{
		BlockchainID: blockchain.ChainID,
		SyncCheck:    newSQLNullString(blockchain.SyncCheck),
		Body:         newSQLNullString(blockchain.SyncCheckOptions.Body),
		Path:         newSQLNullString(blockchain.SyncCheckOptions.Path),
		ResultKey:    newSQLNullString(blockchain.SyncCheckOptions.ResultKey),
		Allowance:    newSQLNullInt32(int32(blockchain.SyncCheckOptions.Allowance)),
	}
}

// WriteBlockchain saves input blockchain in the database
func (d *PostgresDriver) WriteBlockchain(blockchain *repository.Blockchain) (*repository.Blockchain, error) {
	id, err := generateRandomID()
	if err != nil {
		return nil, err
	}

	blockchain.ID = id
	blockchain.CreatedAt = time.Now()
	blockchain.UpdatedAt = time.Now()

	insertApp := extractInsertDBBlockchain(blockchain)

	nullables := []nullable{}
	nullablesScripts := []string{}

	nullables = append(nullables, extractInsertSyncCheckOptions(blockchain))
	nullablesScripts = append(nullablesScripts, insertSyncCheckOptionsScript)

	tx, err := d.Beginx()
	if err != nil {
		return nil, err
	}

	_, err = tx.NamedExec(insertBlockchainScript, insertApp)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(nullables); i++ {
		if nullables[i].isNotNull() {
			_, err = tx.NamedExec(nullablesScripts[i], nullables[i])
			if err != nil {
				return nil, err
			}
		}
	}

	return blockchain, tx.Commit()
}
