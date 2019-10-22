help:                           ## Display this help message.
	@echo "Please use \`make <target>\` where <target> is one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
		awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'

# extra flags like -v
REFORM_TEST_FLAGS ?=

# SHELL = go run .github/shell.go

deps:                           ## Install dependencies.
	go get -u github.com/lib/pq
	go get -u github.com/jackc/pgx/stdlib
	go get -u github.com/go-sql-driver/mysql
	go get -u github.com/mattn/go-sqlite3
	go get -u github.com/denisenkom/go-mssqldb

	go get -u github.com/AlekSi/gocoverutil
	go get -u github.com/AlekSi/pointer
	go get -u github.com/brianvoe/gofakeit
	go get -u github.com/stretchr/testify/...

deps-check:                     ## Install linters.
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(shell go env GOPATH)/bin

check:                          ## Run linters.
	$(shell go env GOPATH)/bin/golangci-lint run ./...

test:                           ## Run unit tests, generate models, install tools.
	rm -f *.cover coverage.txt
	rm -f internal/test/models/*_reform.go
	rm -f reform-db/*_reform.go

	go install -v github.com/mc2soft/reform/reform
	go test $(REFORM_TEST_FLAGS) -covermode=count -coverprofile=parse.cover github.com/mc2soft/reform/parse
	go generate -v -x github.com/mc2soft/reform/internal/test/models
	go install -v github.com/mc2soft/reform/internal/test/models

	go generate -v -x github.com/mc2soft/reform/reform-db
	go install -v github.com/mc2soft/reform/reform-db


test-db:                        ## Initialize database and run integration tests.
	-reform-db -db-driver="$(REFORM_DRIVER)" -db-source="$(REFORM_ROOT_SOURCE)" -db-wait=15s exec \
		internal/test/sql/$(REFORM_DATABASE)_drop.sql
	reform-db -db-driver="$(REFORM_DRIVER)" -db-source="$(REFORM_ROOT_SOURCE)" exec \
		internal/test/sql/$(REFORM_DATABASE)_create.sql

	# TODO remove that hack in reform 1.4
	# https://github.com/go-reform/reform/issues/151
	# https://github.com/go-reform/reform/issues/157
	cat \
		internal/test/sql/$(REFORM_DATABASE)_init.sql \
		internal/test/sql/data.sql \
		internal/test/sql/$(REFORM_DATABASE)_data.sql \
		internal/test/sql/$(REFORM_DATABASE)_set.sql \
		> internal/test/sql/$(REFORM_DATABASE)_combined.tmp.sql
	reform-db -db-driver="$(REFORM_DRIVER)" -db-source="$(REFORM_INIT_SOURCE)" exec \
		internal/test/sql/$(REFORM_DATABASE)_combined.tmp.sql

	go test $(REFORM_TEST_FLAGS) -covermode=count -coverprofile=reform-db.cover github.com/mc2soft/reform/reform-db
	go test $(REFORM_TEST_FLAGS) -covermode=count -coverprofile=reform.cover
	gocoverutil -coverprofile=coverage.txt merge *.cover
	rm -f *.cover

test-dc:                        ## Run all integration tests with Docker Compose.
	go run .github/test-dc.go test

slaves: export REFORM_DRIVER = postgres
slaves: export REFORM_TEST_SOURCE = postgres://localhost:5432/reform-test?sslmode=disable&TimeZone=America/New_York
slaves: export REFORM_TEST_SOURCE_SLAVE = postgres://localhost:5432/reform-test-slave?sslmode=disable&TimeZone=America/New_York
slaves: export REFORM_TEST_SOURCE_MASTER = postgres://localhost:5432/reform-test-master?sslmode=disable&TimeZone=America/New_York
slaves:
	-dropdb reform-test
	createdb reform-test
	-dropdb reform-test-slave
	createdb reform-test-slave
	-dropdb reform-test-master
	createdb reform-test-master
	env PGTZ=UTC psql -v ON_ERROR_STOP=1 -q -d reform-test < internal/test/sql/postgres_init.sql
	env PGTZ=UTC psql -v ON_ERROR_STOP=1 -q -d reform-test < internal/test/sql/data.sql
	env PGTZ=UTC psql -v ON_ERROR_STOP=1 -q -d reform-test < internal/test/sql/postgres_data.sql
	env PGTZ=UTC psql -v ON_ERROR_STOP=1 -q -d reform-test < internal/test/sql/postgres_set.sql
	env PGTZ=UTC psql -v ON_ERROR_STOP=1 -q -d reform-test-slave < internal/test/sql/postgres_init.sql
	env PGTZ=UTC psql -v ON_ERROR_STOP=1 -q -d reform-test-master < internal/test/sql/postgres_init.sql
	go test -coverprofile=test_lib_pq.cover

# run unit tests and integration tests for PostgreSQL (postgres driver)
postgres: export REFORM_DATABASE = postgres
postgres: export DATABASE = postgres
postgres: export REFORM_DRIVER = postgres
postgres: export REFORM_ROOT_SOURCE = postgres://postgres@127.0.0.1/template1?sslmode=disable
postgres: export REFORM_INIT_SOURCE = postgres://postgres@127.0.0.1/reform-database?sslmode=disable&TimeZone=UTC
postgres: export REFORM_TEST_SOURCE = postgres://postgres@127.0.0.1/reform-database?sslmode=disable&TimeZone=America/New_York
postgres: test
	make test-db

# run unit tests and integration tests for PostgreSQL (pgx driver)
pgx: export REFORM_DATABASE = postgres
pgx: export REFORM_DRIVER = pgx
pgx: export REFORM_ROOT_SOURCE = postgres://postgres@127.0.0.1/template1?sslmode=disable
pgx: export REFORM_INIT_SOURCE = postgres://postgres@127.0.0.1/reform-database?sslmode=disable&TimeZone=UTC
pgx: export REFORM_TEST_SOURCE = postgres://postgres@127.0.0.1/reform-database?sslmode=disable&TimeZone=America/New_York
pgx: test
	make test-db

# run unit tests and integration tests for MySQL (ANSI SQL mode)
mysql: export REFORM_DATABASE = mysql
mysql: export REFORM_DRIVER = mysql
mysql: export REFORM_ROOT_SOURCE = root@/mysql
mysql: export REFORM_INIT_SOURCE = root@/reform-database?parseTime=true&clientFoundRows=true&time_zone='UTC'&sql_mode='ANSI'&multiStatements=true
mysql: export REFORM_TEST_SOURCE = root@/reform-database?parseTime=true&clientFoundRows=true&time_zone='America%2FNew_York'&sql_mode='ANSI'
mysql: test
	make test-db

# run unit tests and integration tests for MySQL (traditional SQL mode + interpolateParams)
mysql-traditional: export REFORM_DATABASE = mysql
mysql-traditional: export REFORM_DRIVER = mysql
mysql-traditional: export REFORM_ROOT_SOURCE = root@/mysql
mysql-traditional: export REFORM_INIT_SOURCE = root@/reform-database?parseTime=true&clientFoundRows=true&time_zone='UTC'&sql_mode='ANSI'&multiStatements=true
mysql-traditional: export REFORM_TEST_SOURCE = root@/reform-database?parseTime=true&clientFoundRows=true&time_zone='America%2FNew_York'&sql_mode='TRADITIONAL'&interpolateParams=true
mysql-traditional: test
	make test-db

# run unit tests and integration tests for SQLite3
sqlite3: export REFORM_DATABASE = sqlite3
sqlite3: export REFORM_DRIVER = sqlite3
sqlite3: export REFORM_ROOT_SOURCE = $(CURDIR)/reform-database.sqlite3
sqlite3: export REFORM_INIT_SOURCE = $(CURDIR)/reform-database.sqlite3
sqlite3: export REFORM_TEST_SOURCE = $(CURDIR)/reform-database.sqlite3
sqlite3: test
	rm -f $(CURDIR)/reform-database.sqlite3
	make test-db

# run unit tests and integration tests for SQL Server (mssql driver)
mssql: export REFORM_DATABASE = mssql
mssql: export REFORM_DRIVER = mssql
mssql: export REFORM_ROOT_SOURCE = server=localhost;user id=sa;password=reform-password123
mssql: export REFORM_INIT_SOURCE = server=localhost;user id=sa;password=reform-password123;database=reform-database
mssql: export REFORM_TEST_SOURCE = server=localhost;user id=sa;password=reform-password123;database=reform-database
mssql: test
	make test-db

# run unit tests and integration tests for SQL Server (sqlserver driver)
sqlserver: export REFORM_DATABASE = mssql
sqlserver: export REFORM_DRIVER = sqlserver
sqlserver: export REFORM_ROOT_SOURCE = server=localhost;user id=sa;password=reform-password123
sqlserver: export REFORM_INIT_SOURCE = server=localhost;user id=sa;password=reform-password123;database=reform-database
sqlserver: export REFORM_TEST_SOURCE = server=localhost;user id=sa;password=reform-password123;database=reform-database
sqlserver: test
	make test-db

# Windows: run unit tests and integration tests for SQL Server (mssql driver)
win-mssql: REFORM_SQL_HOST ?= 127.0.0.1
win-mssql: REFORM_SQL_INSTANCE ?= SQLEXPRESS
win-mssql: export REFORM_DATABASE = mssql
win-mssql: export REFORM_DRIVER = mssql
win-mssql: export REFORM_ROOT_SOURCE = server=$(REFORM_SQL_HOST)\$(REFORM_SQL_INSTANCE)
win-mssql: export REFORM_INIT_SOURCE = server=$(REFORM_SQL_HOST)\$(REFORM_SQL_INSTANCE);database=reform-database
win-mssql: export REFORM_TEST_SOURCE = server=$(REFORM_SQL_HOST)\$(REFORM_SQL_INSTANCE);database=reform-database
win-mssql: test
	make test-db

# Windows: run unit tests and integration tests for SQL Server (sqlserver driver)
win-sqlserver: REFORM_SQL_HOST ?= 127.0.0.1
win-sqlserver: REFORM_SQL_INSTANCE ?= SQLEXPRESS
win-sqlserver: export REFORM_DATABASE = mssql
win-sqlserver: export REFORM_DRIVER = sqlserver
win-sqlserver: export REFORM_ROOT_SOURCE = sqlserver://$(REFORM_SQL_HOST)/$(REFORM_SQL_INSTANCE)
win-sqlserver: export REFORM_INIT_SOURCE = sqlserver://$(REFORM_SQL_HOST)/$(REFORM_SQL_INSTANCE)?database=reform-database
win-sqlserver: export REFORM_TEST_SOURCE = sqlserver://$(REFORM_SQL_HOST)/$(REFORM_SQL_INSTANCE)?database=reform-database
win-sqlserver: test
	make test-db

.PHONY: docs parse reform reform-db
