package postgresdriver

import (
	"database/sql"

	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	selectRedirectsScript = `
	SELECT blockchain_id, alias, loadbalancer, domain 
	FROM redirects`
	insertRedirectScript = `
	INSERT into redirects (blockchain_id, alias, loadbalancer, domain)
	VALUES (:blockchain_id, :alias, :loadbalancer, :domain)`
)

type dbRedirect struct {
	BlockchainID   string         `db:"blockchain_id"`
	Alias          sql.NullString `db:"alias"`
	LoadBalancerID sql.NullString `db:"loadbalancer"`
	Domain         sql.NullString `db:"domain"`
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
