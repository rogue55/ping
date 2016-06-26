all: deps fmt build test

db:
	make -C dbsetup
	dbsetup/dbsetup

fmt:
	go fmt ./...

build: deps
	go build .

testdeps:
	script/setup-test-database

test: deps testdeps db
	go test ./...

deps:
	go get github.com/go-sql-driver/mysql \
		github.com/jmoiron/sqlx \
		github.com/parkr/gossip/serializer
