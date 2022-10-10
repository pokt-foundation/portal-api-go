package postgresdriver

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	selectBlockchainsScript = `SELECT b.blockchain_id, b.altruist, b.blockchain, b.blockchain_aliases, b.chain_id, b.chain_id_check, b.description, b.enforce_result, b.log_limit_blocks, b.network, b.path, b.request_timeout, b.ticker, b.active,
	s.synccheck as s_sync_check, s.allowance as s_allowance, s.body as s_body, s.path as s_path, s.result_key as s_result_key
	FROM blockchains as b
	LEFT JOIN sync_check_options AS s ON b.blockchain_id=s.blockchain_id`
	insertBlockchainScript = `
	INSERT into blockchains (blockchain_id, active, altruist, blockchain, blockchain_aliases, chain_id, chain_id_check, description, enforce_result, log_limit_blocks, network, path, request_timeout, ticker, created_at, updated_at)
	VALUES (:blockchain_id, :active, :altruist, :blockchain, :blockchain_aliases, :chain_id, :chain_id_check, :description, :enforce_result, :log_limit_blocks, :network, :path, :request_timeout, :ticker, :created_at, :updated_at)`
	insertSyncCheckOptionsScript = `
	INSERT into sync_check_options (blockchain_id, synccheck, allowance, body, path, result_key)
	VALUES (:blockchain_id, :synccheck, :allowance, :body, :path, :result_key)`
	activateBlockchain = `
	UPDATE blockchains
	SET active = :active, updated_at = :updated_at
	WHERE blockchain_id = :blockchain_id`
)

type dbBlockchain struct {
	BlockchainID      string         `db:"blockchain_id"`
	Altruist          sql.NullString `db:"altruist"`
	Blockchain        sql.NullString `db:"blockchain"`
	ChainID           sql.NullString `db:"chain_id"`
	ChainIDCheck      sql.NullString `db:"chain_id_check"`
	ChainPath         sql.NullString `db:"path"`
	Description       sql.NullString `db:"description"`
	EnforceResult     sql.NullString `db:"enforce_result"`
	Network           sql.NullString `db:"network"`
	NetworkID         sql.NullString `db:"network_id"`
	Ticker            sql.NullString `db:"ticker"`
	BlockchainAliases pq.StringArray `db:"blockchain_aliases"`
	LogLimitBlocks    sql.NullInt32  `db:"log_limit_blocks"`
	RequestTimeout    sql.NullInt32  `db:"request_timeout"`
	Active            sql.NullBool   `db:"active"`
	SyncCheck         sql.NullString `db:"s_sync_check"`
	Allowance         sql.NullInt32  `db:"s_allowance"`
	Body              sql.NullString `db:"s_body"`
	Path              sql.NullString `db:"s_path"`
	ResultKey         sql.NullString `db:"s_result_key"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
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
		Path:              b.ChainPath.String,
		SyncCheck:         b.SyncCheck.String,
		Ticker:            b.Ticker.String,
		BlockchainAliases: b.BlockchainAliases,
		LogLimitBlocks:    int(b.LogLimitBlocks.Int32),
		RequestTimeout:    int(b.RequestTimeout.Int32),
		Active:            b.Active.Bool,
		SyncCheckOptions: repository.SyncCheckOptions{
			Body:      b.Body.String,
			ResultKey: b.ResultKey.String,
			Path:      b.Path.String,
			Allowance: int(b.Allowance.Int32),
		},
	}
}

type dbBlockchainJSON struct {
	BlockchainID      string    `json:"blockchain_id"`
	Altruist          string    `json:"altruist"`
	Blockchain        string    `json:"blockchain"`
	ChainID           string    `json:"chain_id"`
	ChainIDCheck      string    `json:"chain_id_check"`
	ChainPath         string    `json:"path"`
	Description       string    `json:"description"`
	EnforceResult     string    `json:"enforce_result"`
	Network           string    `json:"network"`
	Ticker            string    `json:"ticker"`
	BlockchainAliases []string  `json:"blockchain_aliases"`
	LogLimitBlocks    int       `json:"log_limit_blocks"`
	RequestTimeout    int       `json:"request_timeout"`
	Active            bool      `json:"active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (j dbBlockchainJSON) toOutput() any {
	return &repository.Blockchain{
		ID:                j.BlockchainID,
		Altruist:          j.Altruist,
		Blockchain:        j.Blockchain,
		ChainID:           j.ChainID,
		ChainIDCheck:      j.ChainIDCheck,
		Path:              j.ChainPath,
		Description:       j.Description,
		EnforceResult:     j.EnforceResult,
		Network:           j.Network,
		Ticker:            j.Ticker,
		BlockchainAliases: j.BlockchainAliases,
		LogLimitBlocks:    j.LogLimitBlocks,
		RequestTimeout:    j.RequestTimeout,
		Active:            j.Active,
		CreatedAt:         j.CreatedAt,
		UpdatedAt:         j.UpdatedAt,
	}
}

type insertDBBlockchain struct {
	BlockchainID      string         `db:"blockchain_id"`
	Altruist          sql.NullString `db:"altruist"`
	Blockchain        sql.NullString `db:"blockchain"`
	ChainID           sql.NullString `db:"chain_id"`
	ChainIDCheck      sql.NullString `db:"chain_id_check"`
	ChainPath         sql.NullString `db:"path"`
	Description       sql.NullString `db:"description"`
	EnforceResult     sql.NullString `db:"enforce_result"`
	Network           sql.NullString `db:"network"`
	Ticker            sql.NullString `db:"ticker"`
	BlockchainAliases pq.StringArray `db:"blockchain_aliases"`
	LogLimitBlocks    sql.NullInt32  `db:"log_limit_blocks"`
	RequestTimeout    sql.NullInt32  `db:"request_timeout"`
	Active            bool           `db:"active"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
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
		BlockchainID:      blockchain.ID,
		Altruist:          newSQLNullString(blockchain.Altruist),
		Blockchain:        newSQLNullString(blockchain.Blockchain),
		ChainID:           newSQLNullString(blockchain.ChainID),
		ChainIDCheck:      newSQLNullString(blockchain.ChainIDCheck),
		ChainPath:         newSQLNullString(blockchain.Path),
		Description:       newSQLNullString(blockchain.Description),
		EnforceResult:     newSQLNullString(blockchain.EnforceResult),
		Network:           newSQLNullString(blockchain.Network),
		Ticker:            newSQLNullString(blockchain.Ticker),
		BlockchainAliases: blockchain.BlockchainAliases,
		LogLimitBlocks:    newSQLNullInt32(int32(blockchain.LogLimitBlocks)),
		RequestTimeout:    newSQLNullInt32(int32(blockchain.RequestTimeout)),
		Active:            blockchain.Active,
	}
}

type dbSyncCheckOptionsJSON struct {
	BlockchainID string `json:"blockchain_id"`
	SyncCheck    string `json:"synccheck"`
	Body         string `json:"body"`
	Path         string `json:"path"`
	ResultKey    string `json:"result_key"`
	Allowance    int    `json:"allowance"`
}

func (j dbSyncCheckOptionsJSON) toOutput() any {
	return &repository.SyncCheckOptions{
		BlockchainID: j.BlockchainID,
		Body:         j.Body,
		Path:         j.Path,
		ResultKey:    j.ResultKey,
		Allowance:    j.Allowance,
	}
}

type insertSyncCheckOptions struct {
	BlockchainID string         `db:"blockchain_id"`
	SyncCheck    sql.NullString `db:"synccheck"`
	Body         sql.NullString `db:"body"`
	Path         sql.NullString `db:"path"`
	ResultKey    sql.NullString `db:"result_key"`
	Allowance    sql.NullInt32  `db:"allowance"`
}

func (i *insertSyncCheckOptions) isNotNull() bool {
	return i.SyncCheck.Valid || i.Body.Valid || i.Path.Valid || i.ResultKey.Valid || i.Allowance.Valid
}

func extractInsertSyncCheckOptions(blockchain *repository.Blockchain) *insertSyncCheckOptions {
	return &insertSyncCheckOptions{
		BlockchainID: blockchain.ID,
		SyncCheck:    newSQLNullString(blockchain.SyncCheck),
		Body:         newSQLNullString(blockchain.SyncCheckOptions.Body),
		Path:         newSQLNullString(blockchain.SyncCheckOptions.Path),
		ResultKey:    newSQLNullString(blockchain.SyncCheckOptions.ResultKey),
		Allowance:    newSQLNullInt32(int32(blockchain.SyncCheckOptions.Allowance)),
	}
}

// WriteBlockchain saves input blockchain in the database
func (d *PostgresDriver) WriteBlockchain(blockchain *repository.Blockchain) (*repository.Blockchain, error) {
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

type activateDBBlockchain struct {
	BlockchainID string    `db:"blockchain_id"`
	Active       bool      `db:"active"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (d *PostgresDriver) ActivateBlockchain(id string, active bool) error {
	if id == "" {
		return ErrMissingID
	}

	tx, err := d.Beginx()
	if err != nil {
		return err
	}

	update := &activateDBBlockchain{
		BlockchainID: id,
		Active:       active,
		UpdatedAt:    time.Now(),
	}

	_, err = tx.NamedExec(activateBlockchain, update)
	if err != nil {
		return err
	}

	return tx.Commit()
}
