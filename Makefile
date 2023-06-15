tidy:
	go mod tidy

update:
	go get -u ./...

test:
	go test ./...

test-cover:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out