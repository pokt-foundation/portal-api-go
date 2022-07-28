package postgresdriver

import "github.com/pokt-foundation/portal-api-go/repository"

const (
	selectUsers = "SELECT user_id FROM users"
)

type dbUser struct {
	UserID string `db:"user_id"`
}

func (u *dbUser) toUser() *repository.User {
	return &repository.User{
		ID: u.UserID,
	}
}

// ReadUsers returns all users saved in the database
func (d *PostgresDriver) ReadUsers() ([]*repository.User, error) {
	var dbUsers []*dbUser

	err := d.Select(&dbUsers, selectUsers)
	if err != nil {
		return nil, err
	}

	var users []*repository.User

	for _, dbUser := range dbUsers {
		users = append(users, dbUser.toUser())
	}

	return users, nil
}
