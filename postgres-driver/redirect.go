package postgresdriver

import (
	"database/sql"

	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	selectRedirectsScript = `
	SELECT blockchain_id, alias, loadbalancer, domain 
	FROM redirects`
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
