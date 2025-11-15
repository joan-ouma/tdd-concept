.PHONY: build test coverage clean

build:
	go build -o curlem main.go

test:
	go test -v ./... -coverprofile=coverage.out

coverage: test
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -f curlem coverage.out coverage.html