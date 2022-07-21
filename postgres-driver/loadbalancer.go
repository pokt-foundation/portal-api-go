package postgresdriver

import (
	"database/sql"
	"time"
)

const (
	selectLoadbalancerByID = "SELECT * FROM loadbalancers WHERE lb_id = $1"
)

type dbLoadbalancer struct {
	ID                int            `db:"id"`
	LbID              string         `db:"lb_id"`
	Name              sql.NullString `db:"name"`
	CreatedAt         sql.NullTime   `db:"created_at"`
	UpdatedAt         sql.NullTime   `db:"updated_at"`
	RequestTimeout    sql.NullInt64  `db:"request_timeout"`
	Gigastake         sql.NullBool   `db:"gigastake"`
	GigastakeRedirect sql.NullBool   `db:"gigastake_redirect"`
	UserID            sql.NullString `db:"user_id"`
}

func (lb *dbLoadbalancer) toLoadBalancer() *Loadbalancer {
	return &Loadbalancer{
		LbID:              lb.LbID,
		Name:              lb.Name.String,
		CreatedAt:         &lb.CreatedAt.Time,
		UpdatedAt:         &lb.UpdatedAt.Time,
		RequestTimeout:    lb.RequestTimeout.Int64,
		Gigastake:         lb.Gigastake.Bool,
		GigastakeRedirect: lb.GigastakeRedirect.Bool,
		UserID:            lb.UserID.String,
	}
}

// Loadbalancer struct handler representing a Loadbalancer
type Loadbalancer struct {
	LbID              string
	Name              string
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
	RequestTimeout    int64
	Gigastake         bool
	GigastakeRedirect bool
	UserID            string
}

// ReadLoadbalancerByID returns loadbalancer in the database with given id
func (d *PostgresDriver) ReadLoadbalancerByID(id string) (*Loadbalancer, error) {
	var dbLoadbalancer dbLoadbalancer

	err := d.Get(&dbLoadbalancer, selectLoadbalancerByID, id)
	if err != nil {
		return nil, err
	}

	return dbLoadbalancer.toLoadBalancer(), nil
}
