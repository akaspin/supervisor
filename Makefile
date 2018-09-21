export GO111MODULE = on

test-race:
	go test -race ./...

test-coverage:
	go test -coverprofile=coverage.txt -covermode=atomic ./...

lint: CHECK-toolchain
	DIFF=`gofmt -s -d .` && echo "$$DIFF" && test -z "$$DIFF"
	go vet ./...
	revive -config revive.toml -formatter friendly -exclude ./vendor/... ./...

.PHONY: INSTALL-toolchain
INSTALL-toolchain:
	mkdir -p .tool && cd .tool && \
		echo "module toolchain" > go.mod && \
		go get -u github.com/mgechev/revive
	rm -rf .tool

.PHONY: CHECK-toolchain
CHECK-toolchain:
	which revive

.PHONY: lint test
