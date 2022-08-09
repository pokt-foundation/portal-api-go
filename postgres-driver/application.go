package postgresdriver

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/lib/pq"
	"github.com/pokt-foundation/portal-api-go/repository"
)

const (
	selectApplications = `
	SELECT a.application_id, a.contact_email, a.created_at, a.description, a.free_tier, a.max_relays, a.name, a.owner, a.status, a.updated_at, a.url, a.user_id, 
	fa.public_key AS fa_public_key, fa.signature AS fa_signature, fa.client_public_key AS fa_client_public_key, fa.version AS fa_version, 
	ga.public_key AS ga_public_key, ga.signature AS ga_signature, ga.client_public_key AS ga_client_public_key, ga.version AS ga_version,
	fac.public_key AS fac_public_key, fac.address AS fac_address, fac.private_key AS fac_private_key, fac.version AS fac_version,
	pa.public_key AS pa_public_key, pa.address AS pa_address,
	gs.secret_key, gs.secret_key_required, gs.whitelist_blockchains, gs.whitelist_contracts, gs.whitelist_methods, gs.whitelist_origins, gs.whitelist_user_agents,
	ns.signed_up, ns.on_quarter, ns.on_half, ns.on_three_quarters, ns.on_full,
	FROM applications AS a
	LEFT JOIN freetier_aat AS fa ON a.application_id=fa.application_id
	LEFT JOIN gateway_aat AS ga ON a.application_id=ga.application_id
	LEFT JOIN freetier_app_account AS fac ON a.application_id=fac.application_id
	LEFT JOIN public_pocket_account AS pa ON a.application_id=pa.application_id
	LEFT JOIN gateway_settings AS gs ON a.application_id=gs.application_id
	LEFT JOIN notification_settings AS ns ON a.application_id=ns.application_id
	`
)

type dbApplication struct {
	ApplicationID        string         `db:"application_id"`
	UserID               sql.NullString `db:"user_id"`
	Name                 sql.NullString `db:"name"`
	Status               sql.NullString `db:"status"`
	ContactEmail         sql.NullString `db:"contact_email"`
	Description          sql.NullString `db:"description"`
	FAPublicKey          sql.NullString `db:"fa_public_key"`
	FASignature          sql.NullString `db:"fa_signature"`
	FAClientPublicKey    sql.NullString `db:"fa_client_public_key"`
	FAVersion            sql.NullString `db:"fa_version"`
	FACPublicKey         sql.NullString `db:"fac_public_key"`
	FACAddress           sql.NullString `db:"fac_address"`
	FACPrivateKey        sql.NullString `db:"fac_private_key"`
	FACVersion           sql.NullString `db:"fac_version"`
	GAPublicKey          sql.NullString `db:"ga_public_key"`
	GASignature          sql.NullString `db:"ga_signature"`
	GAClientPublicKey    sql.NullString `db:"ga_client_public_key"`
	GAVersion            sql.NullString `db:"ga_version"`
	Owner                sql.NullString `db:"owner"`
	PAPublicKey          sql.NullString `db:"pa_public_key"`
	PAAdress             sql.NullString `db:"pa_address"`
	SecretKey            sql.NullString `db:"secret_key"`
	URL                  sql.NullString `db:"url"`
	WhitelistContracts   sql.NullString `db:"whitelist_contracts"`
	WhitelistMethods     sql.NullString `db:"whitelist_methods"`
	WhitelistOrigins     pq.StringArray `db:"whitelist_origins"`
	WhitelistUserAgents  pq.StringArray `db:"whitelist_user_agents"`
	WhitelistBlockchains pq.StringArray `db:"whitelist_blockchains"`
	MaxRelays            sql.NullInt32  `db:"max_relays"`
	Dummy                sql.NullBool   `db:"dummy"`
	FreeTier             sql.NullBool   `db:"free_tier"`
	SecretKeyRequired    sql.NullBool   `db:"secret_key_required"`
	SignedUp             sql.NullBool   `db:"signed_up"`
	Quarter              sql.NullBool   `db:"on_quarter"`
	Half                 sql.NullBool   `db:"on_half"`
	ThreeQuarters        sql.NullBool   `db:"on_three_quarters"`
	Full                 sql.NullBool   `db:"on_full"`
	CreatedAt            sql.NullTime   `db:"created_at"`
	UpdatedAt            sql.NullTime   `db:"updated_at"`
}

func stringToWhitelistContracts(rawContracts sql.NullString) []repository.WhitelistContract {
	contracts := []repository.WhitelistContract{}

	if !rawContracts.Valid {
		return contracts
	}

	_ = json.Unmarshal([]byte(rawContracts.String), &contracts)

	for i, contract := range contracts {
		for j, inContract := range contract.Contracts {
			contracts[i].Contracts[j] = strings.TrimSpace(inContract)
		}
	}

	return contracts
}

func stringToWhitelistMethods(rawMethods sql.NullString) []repository.WhitelistMethod {
	methods := []repository.WhitelistMethod{}

	if !rawMethods.Valid {
		return methods
	}

	_ = json.Unmarshal([]byte(rawMethods.String), &methods)

	for i, method := range methods {
		for j, inMethod := range method.Methods {
			methods[i].Methods[j] = strings.TrimSpace(inMethod)
		}
	}

	return methods
}

func (a *dbApplication) toApplication() *repository.Application {
	return &repository.Application{
		ID:           a.ApplicationID,
		UserID:       a.UserID.String,
		Name:         a.Name.String,
		Status:       a.Status.String,
		ContactEmail: a.ContactEmail.String,
		Description:  a.Description.String,
		Owner:        a.Owner.String,
		URL:          a.URL.String,
		MaxRelays:    int(a.MaxRelays.Int32),
		Dummy:        a.Dummy.Bool,
		FreeTier:     a.FreeTier.Bool,
		CreatedAt:    a.CreatedAt.Time,
		UpdatedAt:    a.UpdatedAt.Time,
		FreeTierAAT: repository.FreeTierAAT{
			ApplicationPublicKey: a.FAPublicKey.String,
			ApplicationSignature: a.FASignature.String,
			ClientPublicKey:      a.FAClientPublicKey.String,
			Version:              a.FAVersion.String,
		},
		FreeTierApplicationAccount: repository.FreeTierApplicationAccount{
			Address:    a.FACAddress.String,
			PublicKey:  a.FACPublicKey.String,
			PrivateKey: a.FACPrivateKey.String,
			Version:    a.FACVersion.String,
		},
		GatewayAAT: repository.GatewayAAT{
			ApplicationPublicKey: a.GAPublicKey.String,
			ApplicationSignature: a.GASignature.String,
			ClientPublicKey:      a.GAClientPublicKey.String,
			Version:              a.GAVersion.String,
		},
		GatewaySettings: repository.GatewaySettings{
			SecretKey:            a.SecretKey.String,
			SecretKeyRequired:    a.SecretKeyRequired.Bool,
			WhitelistBlockchains: a.WhitelistBlockchains,
			WhitelistContracts:   stringToWhitelistContracts(a.WhitelistContracts),
			WhitelistMethods:     stringToWhitelistMethods(a.WhitelistMethods),
			WhitelistOrigins:     a.WhitelistOrigins,
			WhitelistUserAgents:  a.WhitelistUserAgents,
		},
		NotificationSettings: repository.NotificationSettings{
			SignedUp:      a.SignedUp.Bool,
			Quarter:       a.Quarter.Bool,
			Half:          a.Half.Bool,
			ThreeQuarters: a.ThreeQuarters.Bool,
			Full:          a.Full.Bool,
		},
		PublicPocketAccount: repository.PublicPocketAccount{
			PublicKey: a.PAPublicKey.String,
			Address:   a.PAAdress.String,
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
