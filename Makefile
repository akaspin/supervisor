# required to use modules inside GOPATH
export GO111MODULE = on

# Use $(eval $(call TOOLCHAIN,<binary>,<package>)) to make pair of
# INSTALL-<package> and .CHECK-<package> targets. Use INSTALL and .CHECK
# targets to install or check all tools

define TOOLCHAIN
.PHONY: INSTALL-$1
INSTALL-$1:
	mkdir -p /tmp/.INSTALL-$1 && cd /tmp/.INSTALL-$1 && \
		echo "module toolchain" > go.mod && \
		go get -u $2
	rm -rf /tmp/.INSTALL-$1
INSTALL:: INSTALL-$1
.PHONY: .CHECK-$1
.CHECK-$1:
	@test -x "`which $1`" || (echo "$1 is not installed. run INSTALL-$1.")
.CHECK:: .CHECK-$1
endef

test-ci:
	go test -race -run=^Test -coverprofile=coverage.txt -covermode=atomic ./...

$(eval $(call TOOLCHAIN,revive,github.com/mgechev/revive))
lint: .CHECK-revive
	DIFF=`gofmt -s -d .` && echo "$$DIFF" && test -z "$$DIFF"
	go vet ./...
	revive -config revive.toml -formatter friendly -exclude ./vendor/... ./...

.PHONY: lint test-ci

