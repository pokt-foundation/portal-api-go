package postgresdriver

import (
	"database/sql"

	"github.com/lib/pq"
	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	selectApplications = `
	SELECT a.application_id, a.contact_email, a.created_at, a.description, a.name, a.owner, a.updated_at, a.url, a.user_id, 
	fa.public_key AS fa_public_key, fa.signature AS fa_signature, fa.client_public_key AS fa_client_public_key, fa.version AS fa_version, 
	ga.public_key AS ga_public_key, ga.signature AS ga_signature, ga.client_public_key AS ga_client_public_key, ga.version AS ga_version,
	fac.public_key AS fac_public_key, fac.address AS fac_address, fac.private_key AS fac_private_key, fac.version AS fac_version,
	pa.public_key AS pa_public_key, pa.address AS pa_address,
	gs.secret_key, gs.secret_key_required, gs.whitelist_blockchains, gs.whitelist_contracts, gs.whitelist_methods, gs.whitelist_origins, gs.whitelist_user_agents
	FROM applications AS a
	LEFT JOIN freetier_aat AS fa ON a.application_id=fa.application_id
	LEFT JOIN gateway_aat AS ga ON a.application_id=ga.application_id
	LEFT JOIN freetier_app_account AS fac ON a.application_id=fac.application_id
	LEFT JOIN public_pocket_account AS pa ON a.application_id=pa.application_id
	LEFT JOIN gateway_settings AS gs ON a.application_id=gs.application_id
	`
)

type dbApplication struct {
	ApplicationID        string         `db:"application_id"`
	ContactEmail         sql.NullString `db:"contact_email"`
	CreatedAt            sql.NullTime   `db:"created_at"`
	Description          sql.NullString `db:"description"`
	Name                 sql.NullString `db:"name"`
	Owner                sql.NullString `db:"owner"`
	UpdatedAt            sql.NullTime   `db:"updated_at"`
	URL                  sql.NullString `db:"url"`
	UserID               sql.NullString `db:"user_id"`
	FAPublicKey          sql.NullString `db:"fa_public_key"`
	FASignature          sql.NullString `db:"fa_signature"`
	FAClientPublicKey    sql.NullString `db:"fa_client_public_key"`
	FAVersion            sql.NullString `db:"fa_version"`
	GAPublicKey          sql.NullString `db:"ga_public_key"`
	GASignature          sql.NullString `db:"ga_signature"`
	GAClientPublicKey    sql.NullString `db:"ga_client_public_key"`
	GAVersion            sql.NullString `db:"ga_version"`
	FACPublicKey         sql.NullString `db:"fac_public_key"`
	FACAdress            sql.NullString `db:"fac_address"`
	FACPrivateKey        sql.NullString `db:"fac_private_key"`
	FACVersion           sql.NullString `db:"fac_version"`
	PAPublicKey          sql.NullString `db:"pa_public_key"`
	PAAdress             sql.NullString `db:"pa_address"`
	SecretKey            sql.NullString `db:"secret_key"`
	SecretKeyRequired    sql.NullBool   `db:"secret_key_required"`
	WhitelistBlockchains pq.StringArray `db:"whitelist_blockchains"`
	WhitelistContracts   pq.StringArray `db:"whitelist_contracts"`
	WhitelistMethods     pq.StringArray `db:"whitelist_methods"`
	WhitelistOrigins     pq.StringArray `db:"whitelist_origins"`
	WhitelistUserAgents  pq.StringArray `db:"whitelist_user_agents"`
}

func (a *dbApplication) toApplication() *repository.Application {
	return &repository.Application{
		ID:           a.ApplicationID,
		ContactEmail: a.ContactEmail.String,
		CreatedAt:    &a.CreatedAt.Time,
		Description:  a.Description.String,
		Name:         a.Name.String,
		Owner:        a.Owner.String,
		UpdatedAt:    &a.UpdatedAt.Time,
		URL:          a.URL.String,
		UserID:       a.UserID.String,
		PublicPocketAccount: repository.PublicPocketAccount{
			PublicKey: a.PAPublicKey.String,
			Address:   a.PAAdress.String,
		},
		FreeTierApplicationAccount: repository.FreeTierApplicationAccount{
			Address:    a.FACAdress.String,
			PublicKey:  a.FACPublicKey.String,
			PrivateKey: a.FACPrivateKey.String,
			Version:    a.FACVersion.String,
		},
		FreeTierAAT: repository.FreeTierAAT{
			Version:              a.FAVersion.String,
			ApplicationPublicKey: a.FAPublicKey.String,
			ClientPublicKey:      a.FAClientPublicKey.String,
			ApplicationSignature: a.FASignature.String,
		},
		GatewayAAT: repository.GatewayAAT{
			Version:              a.GAVersion.String,
			ApplicationPublicKey: a.GAPublicKey.String,
			ClientPublicKey:      a.GAClientPublicKey.String,
			ApplicationSignature: a.GASignature.String,
		},
		GatewaySettings: repository.GatewaySettings{
			SecretKey:            a.SecretKey.String,
			SecretKeyRequired:    a.SecretKeyRequired.Bool,
			WhitelistBlockchains: a.WhitelistBlockchains,
			WhitelistContracts:   a.WhitelistContracts,
			WhitelistMethods:     a.WhitelistMethods,
			WhitelistOrigins:     a.WhitelistOrigins,
			WhitelistUserAgents:  a.WhitelistUserAgents,
		},
	}
}

// ReadApplications returns all applications on the database
func (d *PostgresDriver) ReadApplications() ([]*repository.Application, error) {
	var dbApplications []*dbApplication

	err := d.Select(&dbApplications, selectApplications)
	if err != nil {
		return nil, err
	}

	var applications []*repository.Application

	for _, dbApplication := range dbApplications {
		applications = append(applications, dbApplication.toApplication())
	}

	return applications, nil
}
