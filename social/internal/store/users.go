package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail    = errors.New("a user with that email already exists")
	ErrDuplicateUsername = errors.New("a user with that username already exists")
)

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	//Password  string   `json:"-"` ex 43 replacing type string of password with password
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	//ex 45
	IsActive bool `json:"is_active"`
	//ex 56 add RoleId and Role
	RoleID int64 `json:"role_id"`
	Role   Role  `json:"role"`
}

// ex 43 user registration Password type will have text which is pointer to string
// so that can be null and also hash which will be slice of bytes
type password struct {
	text *string
	hash []byte
}

// ex 43
func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	//ex 57
	query := `
	INSERT INTO users (username, password, email, role_id) VALUES 
	($1, $2, $3, (SELECT id FROM roles WHERE name = $4)) RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	//Ex 57 Making sure there is always one role attached by initializing it
	role := user.Role.Name
	if role == "" {
		role = "user"
	}

	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		//ex 45 use hash of the password to store in database
		user.Password.hash,
		user.Email,
		//ex 54 adding role name
		role,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
	//Good way to handle errors as methods CreateAndInvite implementation can use this to give proper errors to end users
	//users_email_key and users_username_key can be checked under User table constraints in postgres
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil

}

func (s *UserStore) GetByID(ctx context.Context, userID int64) (*User, error) {
	// query := `
	// SELECT id, username, email, password, created_at
	// FROM users
	// WHERE id = $1 AND is_active = true
	// `
	//ex 56 Precedence middleware joining roles table to get roles all rows output of roles.*
	query := `
	SELECT users.id, username, email, password, created_at, roles.*
	FROM users
	JOIN roles ON (users.role_id = roles.id)
	WHERE users.id = $1 AND is_active = true
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		//ex 56 returing row of roles for the user
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

// ex44
func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	// transaction wrapper
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		// create the user
		err := s.Create(ctx, tx, user)
		if err != nil {
			return err
		}

		// create the user invite
		err = s.createUserInvitation(ctx, tx, token, invitationExp, user.ID)
		if err != nil {
			return err
		}
		return nil
	})
}

// ex 45 user activation
func (s *UserStore) Activate(ctx context.Context, token string) error {
	//have token in param, need to check inside database if that token matches for that userID
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		//1. find the user that this token belongs to
		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		//2. update the user
		user.IsActive = true
		err = s.update(ctx, tx, user)
		if err != nil {
			return err
		}
		//3. Clean the invitations
		err = s.deleteUserInvitations(ctx, tx, user.ID)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
	SELECT u.id, u.username, u.email, u.created_at, u.is_active
	FROM users u
	JOIN user_invitations ui on u.id = ui.user_id
	WHERE ui.token = $1 AND ui.expiry > $2
`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	//Actually token from request is plaintext token, so we need to hash it to compare it with postgres
	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	user := &User{}
	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.IsActive,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil

}

// ex44
func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userID int64) error {
	query := `INSERT INTO user_invitations (token, user_id, expiry) VALUES ($1, $2, $3)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}

// ex 45
func (s *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `UPDATE users SET username = $1, email = $2, is_active= $3 WHERE id = $4`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}
	return nil
}

// ex 45
func (s *UserStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}

// ex 46 This function used for Deleting the user from users table and cleaning his invitation from user_ invitations
func (s *UserStore) Delete(ctx context.Context, userID int64) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.delete(ctx, tx, userID); err != nil {
			return err
		}

		if err := s.deleteUserInvitations(ctx, tx, userID); err != nil {
			return err
		}

		return nil
	})
}

// ex 46
func (s *UserStore) delete(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `DELETE FROM user WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

// ex 51 generating tokens
func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, username, email, password, created_at FROM users
    			WHERE email = $1 AND is_active = true
				`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password.hash, &user.CreatedAt,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}
