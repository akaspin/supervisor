
# all go sources excluding vendor
GO_SOURCES = $(shell find . -type f \( -iname '*.go' \) -not -path "./vendor/*" 2>/dev/null)
.SUFFIXES:

# Required to use modules inside GOPATH with Go 1.11. Temporary.
export GO111MODULE ?= on

##
## Use $(eval $(call TOOL,<binary>,<package>)) to make .TOOL-<binary> target
## .TOOL-* targets can be used as dependencies for goals which requires
## specific binaries. All binaries will be installed to $GOBIN directory.
## Hint: to install tools to specific directory override GOBIN and PATH
## OS environment variables.
##

define TOOL
.TOOL-$1:
	test -x "$(shell which $1)" \
	|| (mkdir -p /tmp/.INSTALL-$1 && cd /tmp/.INSTALL-$1 && \
		echo "module toolchain" > go.mod && \
		go get -u $2 && \
		rm -rf /tmp/.INSTALL-$1)
.NOTPARALLEL: .TOOL-$1
endef

##
## Maintenance
##

go.mod: $(GO_SOURCES)
	go mod tidy

mod: $(GO_SOURCES)
	go get -u=patch
	go mod tidy

mod-vendor: go.mod
	go mod vendor

fmt: $(GO_SOURCES)
	go fmt ./...

##
## Testing and lint
##

.PHONY: test bench race

test: go.mod
	go test -run=Test ./...

coverage: go.mod
	go test -coverprofile=coverage.txt -covermode=atomic ./...

race: go.mod
	go test -run=Test -race ./...

examples: go.mod
	go test -run=Example ./...

bench: go.mod
	go test -run= -bench=. -benchmem ./...

lint: .ASSERT-fmt .ASSERT-vet .ASSERT-lint

.ASSERT-fmt: $(GO_SOURCES)
	@DIFF=`gofmt -s -d $^` && echo "$$DIFF" && test -z "$$DIFF"

.ASSERT-vet: $(GO_SOURCES) go.mod
	go vet ./...
.NOTPARALLEL: .ASSERT-vet

$(eval $(call TOOL,revive,github.com/mgechev/revive))
.ASSERT-lint: .TOOL-revive $(GO_SOURCES)
	revive -config revive.toml -formatter friendly -exclude ./vendor/... ./...

