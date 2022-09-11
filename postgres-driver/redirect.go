package postgresdriver

import (
	"database/sql"
	"time"

	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	selectRedirectsScript = `
	SELECT blockchain_id, alias, loadbalancer, domain 
	FROM redirects`
	insertRedirectScript = `
	INSERT into redirects (redirect_id, blockchain_id, alias, loadbalancer, domain, created_at, updated_at)
	VALUES (:redirect_id, :blockchain_id, :alias, :loadbalancer, :domain, :created_at, :updated_at)`
)

type dbRedirect struct {
	RedirectID     string         `db:"redirect_id"`
	BlockchainID   string         `db:"blockchain_id"`
	Alias          sql.NullString `db:"alias"`
	LoadBalancerID sql.NullString `db:"loadbalancer"`
	Domain         sql.NullString `db:"domain"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

func (r *dbRedirect) toRedirect() *repository.Redirect {
	return &repository.Redirect{
		BlockchainID:   r.BlockchainID,
		Alias:          r.Alias.String,
		LoadBalancerID: r.LoadBalancerID.String,
		Domain:         r.Domain.String,
	}
}

// ReadRedirects returns all redirects on the database
func (d *PostgresDriver) ReadRedirects() ([]*repository.Redirect, error) {
	var dbRedirects []*dbRedirect

	err := d.Select(&dbRedirects, selectRedirectsScript)
	if err != nil {
		return nil, err
	}

	var redirects []*repository.Redirect

	for _, dbRedirect := range dbRedirects {
		redirects = append(redirects, dbRedirect.toRedirect())
	}

	return redirects, nil
}

func extractDBRedirect(redirect *repository.Redirect) *dbRedirect {
	return &dbRedirect{
		RedirectID:     redirect.ID,
		BlockchainID:   redirect.BlockchainID,
		Alias:          newSQLNullString(redirect.Alias),
		LoadBalancerID: newSQLNullString(redirect.LoadBalancerID),
		Domain:         newSQLNullString(redirect.Domain),
	}
}

// WriteRedirect saves input redirect in the database
func (d *PostgresDriver) WriteRedirect(redirect *repository.Redirect) (*repository.Redirect, error) {
	id, err := generateRandomID()
	if err != nil {
		return nil, err
	}

	redirect.ID = id
	redirect.CreatedAt = time.Now()
	redirect.UpdatedAt = time.Now()

	insertApp := extractDBRedirect(redirect)

	tx, err := d.Beginx()
	if err != nil {
		return nil, err
	}

	_, err = tx.NamedExec(insertRedirectScript, insertApp)
	if err != nil {
		return nil, err
	}

	return redirect, tx.Commit()
}
