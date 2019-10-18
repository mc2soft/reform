all: test_postgresql

prepare:
	find . -name '*_reform.go' -delete
	go test -v github.com/mc2soft/reform/internal/...
	go install -v github.com/mc2soft/reform/reform
	#cd reform-schema && $(GOPATH)/bin/reform -v
	cd test && $(GOPATH)/bin/reform -v

test_postgresql: prepare
	psql -q -d reform-test -f test/drop.sql
	psql -q -d reform-test -f test/create.sql
	cd test && env REFORM_TEST_DRIVER=postgres REFORM_TEST_SOURCE="dbname=reform-test sslmode=disable" go test -v -check.v
