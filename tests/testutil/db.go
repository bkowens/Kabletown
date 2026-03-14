// Package testutil provides shared test helpers for the Kabletown test suite.
package testutil

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

// NewMockDB creates a go-sqlmock database connection wrapped in sqlx.
// The caller is responsible for verifying expectations via mock.ExpectationsWereMet().
func NewMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("testutil.NewMockDB: failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() {
		mockDB.Close()
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("testutil.NewMockDB: unmet sqlmock expectations: %v", err)
		}
	})
	return sqlx.NewDb(mockDB, "sqlmock"), mock
}

// NewRawMockDB creates a go-sqlmock database connection using raw database/sql.
// Used for services like session-service that use database/sql instead of sqlx.
func NewRawMockDB(t *testing.T) (*sqlmock.Sqlmock, sqlmock.Sqlmock) {
	t.Helper()
	_, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("testutil.NewRawMockDB: failed to create sqlmock: %v", err)
	}
	return &mock, mock
}

// ExpectUserQuery sets up sqlmock expectations for a typical user SELECT query.
// Returns the mock for chaining.
func ExpectUserQuery(mock sqlmock.Sqlmock, id, name, password string, isDisabled, isHidden bool) {
	rows := sqlmock.NewRows([]string{"Id", "Name", "Password", "IsDisabled", "IsHidden", "PrimaryImageTag"}).
		AddRow(id, name, password, isDisabled, isHidden, "")
	mock.ExpectQuery("SELECT .+ FROM\\s+Users").WillReturnRows(rows)
}

// ExpectPermissionQuery sets up sqlmock expectations for a Permissions table query.
func ExpectPermissionQuery(mock sqlmock.Sqlmock, userID string, isAdmin bool) {
	val := 0
	if isAdmin {
		val = 1
	}
	rows := sqlmock.NewRows([]string{"Value"}).AddRow(val)
	mock.ExpectQuery("SELECT Value FROM Permissions").WithArgs(userID).WillReturnRows(rows)
}
