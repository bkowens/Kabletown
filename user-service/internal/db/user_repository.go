package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// User represents a row from the Users table, augmented with the IsAdmin flag
// resolved from the Permissions table.
type User struct {
	Id              string `db:"Id"`
	Name            string `db:"Name"`
	Password        string `db:"Password"`
	IsAdmin         bool   // populated from Permissions table
	IsDisabled      bool   `db:"IsDisabled"`
	IsHidden        bool   `db:"IsHidden"`
	PrimaryImageTag string `db:"PrimaryImageTag"`
	Configuration   string `db:"Configuration"`
	Policy          string `db:"Policy"`
}

// UserRepository provides data access operations for the Users table.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new UserRepository backed by the given database.
func NewUserRepository(database *sqlx.DB) *UserRepository {
	return &UserRepository{db: database}
}

// GetUserByID fetches a single user by primary key and populates IsAdmin.
func (r *UserRepository) GetUserByID(id string) (*User, error) {
	var u User
	err := r.db.QueryRowx(
		`SELECT Id, Name, COALESCE(Password,'') AS Password,
		        IsDisabled, IsHidden,
		        COALESCE(PrimaryImageTag,'') AS PrimaryImageTag,
		        COALESCE(Configuration,'') AS Configuration,
		        COALESCE(Policy,'') AS Policy
		 FROM Users WHERE Id = ? LIMIT 1`,
		id,
	).StructScan(&u)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user_repository.GetUserByID: user %s not found", id)
		}
		return nil, fmt.Errorf("user_repository.GetUserByID: %w", err)
	}
	admin, err := r.isAdmin(id)
	if err != nil {
		return nil, err
	}
	u.IsAdmin = admin
	return &u, nil
}

// ListUsers fetches all users and populates IsAdmin for each.
func (r *UserRepository) ListUsers() ([]User, error) {
	rows, err := r.db.Queryx(
		`SELECT Id, Name, COALESCE(Password,'') AS Password,
		        IsDisabled, IsHidden,
		        COALESCE(PrimaryImageTag,'') AS PrimaryImageTag,
		        COALESCE(Configuration,'') AS Configuration,
		        COALESCE(Policy,'') AS Policy
		 FROM Users ORDER BY Name`,
	)
	if err != nil {
		return nil, fmt.Errorf("user_repository.ListUsers: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.StructScan(&u); err != nil {
			return nil, fmt.Errorf("user_repository.ListUsers scan: %w", err)
		}
		admin, err := r.isAdmin(u.Id)
		if err != nil {
			return nil, err
		}
		u.IsAdmin = admin
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user_repository.ListUsers rows: %w", err)
	}
	return users, nil
}

// CreateUser inserts a new user with a bcrypt-hashed password (cost=11).
// It also inserts a Permissions row with Kind=0, Value=0 (not admin by default).
func (r *UserRepository) CreateUser(name, password string) (*User, error) {
	id := uuid.NewString()

	var hash []byte
	if password != "" {
		var err error
		hash, err = bcrypt.GenerateFromPassword([]byte(password), 11)
		if err != nil {
			return nil, fmt.Errorf("user_repository.CreateUser bcrypt: %w", err)
		}
	}

	_, err := r.db.Exec(
		`INSERT INTO Users (Id, Name, Password, IsDisabled, IsHidden, PrimaryImageTag, Configuration, Policy)
		 VALUES (?, ?, ?, 0, 0, '', '', '')`,
		id, name, string(hash),
	)
	if err != nil {
		return nil, fmt.Errorf("user_repository.CreateUser insert: %w", err)
	}

	_, err = r.db.Exec(
		`INSERT INTO Permissions (UserId, Kind, Value) VALUES (?, 0, 0)`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("user_repository.CreateUser permissions: %w", err)
	}

	return r.GetUserByID(id)
}

// UpdateUser updates the Name and Configuration columns for the given user.
func (r *UserRepository) UpdateUser(id, name, configuration, policy string) error {
	_, err := r.db.Exec(
		`UPDATE Users SET Name = ?, Configuration = ?, Policy = ? WHERE Id = ?`,
		name, configuration, policy, id,
	)
	if err != nil {
		return fmt.Errorf("user_repository.UpdateUser: %w", err)
	}
	return nil
}

// DeleteUser removes a user and all their associated Permissions, UserData, and DisplayPreferences.
func (r *UserRepository) DeleteUser(id string) error {
	for _, stmt := range []struct {
		q    string
		desc string
	}{
		{`DELETE FROM UserData WHERE UserId = ?`, "UserData"},
		{`DELETE FROM DisplayPreferences WHERE UserId = ?`, "DisplayPreferences"},
		{`DELETE FROM Permissions WHERE UserId = ?`, "Permissions"},
		{`DELETE FROM Users WHERE Id = ?`, "Users"},
	} {
		if _, err := r.db.Exec(stmt.q, id); err != nil {
			return fmt.Errorf("user_repository.DeleteUser %s: %w", stmt.desc, err)
		}
	}
	return nil
}

// UpdatePassword hashes newPassword with bcrypt cost=11 and stores it.
func (r *UserRepository) UpdatePassword(id, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 11)
	if err != nil {
		return fmt.Errorf("user_repository.UpdatePassword bcrypt: %w", err)
	}
	_, err = r.db.Exec(
		`UPDATE Users SET Password = ? WHERE Id = ?`,
		string(hash), id,
	)
	if err != nil {
		return fmt.Errorf("user_repository.UpdatePassword: %w", err)
	}
	return nil
}

// isAdmin queries the Permissions table for Kind=0 (IsAdministrator).
func (r *UserRepository) isAdmin(userID string) (bool, error) {
	var value int
	err := r.db.QueryRow(
		`SELECT Value FROM Permissions WHERE UserId = ? AND Kind = 0 LIMIT 1`,
		userID,
	).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("user_repository.isAdmin: %w", err)
	}
	return value == 1, nil
}
