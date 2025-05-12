
start-postgres:
	docker run --detach --name postgres-db-test --env POSTGRES_USER=testuser --env POSTGRES_PASSWORD=testpass --env POSTGRES_ROOT_PASSWORD=root --env POSTGRES_DB=test_db -p 5432:5432 postgres:latest

test:
	go test -v ./...

test-with-coverage:
	go test -v ./... -coverprofile=test-with-coverage.out
	go tool cover -html=test-with-coverage.out
