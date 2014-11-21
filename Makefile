all: build

travis: tidy test

install:
	 go install

forego:
	go get github.com/ddollar/forego

test: install
	go test

test-aws: install forego
	forego run go test

tidy:
	go get github.com/golang/lint/golint
	go get code.google.com/p/go.tools/cmd/goimports
	test -z "$$(goimports -l -d . | tee /dev/stderr)"
	test -z "$$(golint . | tee /dev/stderr)"

imports:
	goimports -w .

fmt:
	go fmt ./...

ready: fmt imports tidy

