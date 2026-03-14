.PHONY: test test-shared test-auth test-user test-library test-session test-system test-integration test-verbose test-cover

# Run all tests across all services
test: test-shared test-auth test-user test-library test-session test-system test-integration

# Shared packages
test-shared:
	@echo "==> Testing shared..."
	cd shared && go test ./... -count=1

# Auth service
test-auth:
	@echo "==> Testing auth-service..."
	cd auth-service && go test ./... -count=1

# User service
test-user:
	@echo "==> Testing user-service..."
	cd user-service && go test ./... -count=1

# Library service
test-library:
	@echo "==> Testing library-service..."
	cd library-service && go test ./... -count=1

# Session service
test-session:
	@echo "==> Testing session-service..."
	cd session-service && go test ./... -count=1

# System service
test-system:
	@echo "==> Testing system-service..."
	cd system-service && go test ./... -count=1

# Integration / compatibility tests
test-integration:
	@echo "==> Testing integration (nginx + compat)..."
	cd tests && go test ./... -count=1

# Verbose output for all tests
test-verbose:
	cd shared && go test ./... -count=1 -v
	cd auth-service && go test ./... -count=1 -v
	cd user-service && go test ./... -count=1 -v
	cd library-service && go test ./... -count=1 -v
	cd session-service && go test ./... -count=1 -v
	cd system-service && go test ./... -count=1 -v
	cd tests && go test ./... -count=1 -v

# Coverage report
test-cover:
	@mkdir -p coverage
	cd shared && go test ./... -count=1 -coverprofile=../coverage/shared.out
	cd auth-service && go test ./... -count=1 -coverprofile=../coverage/auth.out
	cd user-service && go test ./... -count=1 -coverprofile=../coverage/user.out
	cd library-service && go test ./... -count=1 -coverprofile=../coverage/library.out
	cd session-service && go test ./... -count=1 -coverprofile=../coverage/session.out
	cd system-service && go test ./... -count=1 -coverprofile=../coverage/system.out
	@echo "Coverage files written to coverage/"
