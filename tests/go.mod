module github.com/jellyfinhanced/tests

go 1.22

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/google/uuid v1.6.0
	github.com/jellyfinhanced/shared v0.0.0
	github.com/jmoiron/sqlx v1.4.0
)

replace (
	github.com/jellyfinhanced/shared => ../shared
	github.com/jellyfinhanced/system-service => ../system-service
)
