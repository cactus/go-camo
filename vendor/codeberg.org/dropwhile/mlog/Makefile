# environment
BUILDDIR            := ${CURDIR}/build
ARCH                := $(shell go env GOHOSTARCH)
OS                  := $(shell go env GOHOSTOS)
GOVER               := $(shell go version | awk '{print $$3}' | tr -d '.')

# app specific info
APP_VER             := v$(shell git describe --always --tags|sed 's/^v//')
GITHASH             := $(shell git rev-parse --short HEAD)
GOPATH              := $(shell go env GOPATH)
GOBIN               := ${CURDIR}/.tools
VERSION_VAR         := main.Version

# flags and build configuration
GOBUILD_OPTIONS     := -trimpath
GOTEST_FLAGS        :=
GOTEST_BENCHFLAGS   :=
GOBUILD_DEPFLAGS    := -tags netgo,production
GOBUILD_LDFLAGS     ?= -s -w
GOBUILD_FLAGS       := ${GOBUILD_DEPFLAGS} ${GOBUILD_OPTIONS} -ldflags "${GOBUILD_LDFLAGS} -X ${VERSION_VAR}=${APP_VER}"

# cross compile defs
CC_BUILD_TARGETS     = 
CC_BUILD_ARCHES      = darwin/amd64 darwin/arm64 freebsd/amd64 linux/amd64 linux/arm64 windows/amd64
CC_OUTPUT_TPL       := ${BUILDDIR}/bin/{{.Dir}}.{{.OS}}-{{.Arch}}

# misc
DOCKER_PREBUILD     ?=

# some exported vars (pre-configure go build behavior)
export GO111MODULE=on
#export CGO_ENABLED=0
## enable go 1.21 loopvar "experiment"
export GOEXPERIMENT=loopvar
export GOBIN
export PATH := ${GOBIN}:${PATH}

define HELP_OUTPUT
Available targets:
  help                this help
  clean               clean up
  all                 build binaries and man pages
  check               run checks and validators
  test                run tests
  cover               run tests with cover output
  bench               run benchmarks
  generate            run go:generate
  build               build all binaries
  update-go-deps      updates go.mod and go.sum files
endef
export HELP_OUTPUT


.PHONY: help
help:
	@echo "$$HELP_OUTPUT"

.PHONY: clean
clean:
	@rm -rf "${BUILDDIR}"

${GOBIN}/stringer:
	go install golang.org/x/tools/cmd/stringer@latest

${GOBIN}/staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest

${GOBIN}/gosec:
	go install github.com/securego/gosec/v2/cmd/gosec@latest

${GOBIN}/govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest

${GOBIN}/errcheck:
	go install github.com/kisielk/errcheck@latest

${GOBIN}/ineffassign:
	go install github.com/gordonklaus/ineffassign@latest

${GOBIN}/nilaway:
	go install go.uber.org/nilaway/cmd/nilaway@latest

BUILD_TOOLS := ${GOBIN}/stringer
CHECK_TOOLS := ${GOBIN}/staticcheck ${GOBIN}/gosec ${GOBIN}/govulncheck
CHECK_TOOLS += ${GOBIN}/errcheck ${GOBIN}/ineffassign ${GOBIN}/nilaway

.PHONY: setup
setup:

.PHONY: setup-build
setup-build: setup ${BUILD_TOOLS}

.PHONY: setup-check
setup-check: setup ${CHECK_TOOLS}

.PHONY: generate
generate: setup-build
	@echo ">> Generating..."
	@PATH="${PATH}" go generate ./...

.PHONY: build
build: setup-build
	@echo ">> Building..."
	@[ -d "${BUILDDIR}/bin" ] || mkdir -p "${BUILDDIR}/bin"
	@(for x in ${CC_BUILD_TARGETS}; do \
		echo "...$${x}..."; \
		go build ${GOBUILD_FLAGS} -o "${BUILDDIR}/bin/$${x}" ./cmd/$${x}; \
	done)
	@echo "done!"

.PHONY: test
test: setup
	@echo ">> Running tests..."
	@go test -count=1 -vet=off ${GOTEST_FLAGS} ./...

.PHONY: bench
bench: setup
	@echo ">> Running benchmarks..."
	@go test -bench="." -run="^$$" -test.benchmem=true ${GOTEST_BENCHFLAGS} ./...

.PHONY: cover
cover: setup
	@echo ">> Running tests with coverage..."
	@go test -vet=off -cover ${GOTEST_FLAGS} ./...

.PHONY: check
check: setup-check
	@echo ">> Running checks and validators..."
	@echo "... staticcheck ..."
	@${GOBIN}/staticcheck ./...
	@echo "... errcheck ..."
	@${GOBIN}/errcheck -ignoretests -exclude .errcheck-excludes.txt ./...
	@echo "... go-vet ..."
	@go vet ./...
	@echo "... gosec ..."
	@${GOBIN}/gosec -quiet -exclude-dir=tool -exclude G104 ./...
	@echo "... ineffassign ..."
	@${GOBIN}/ineffassign ./...
	@echo "... govulncheck ..."

.PHONY: update-go-deps
update-go-deps: setup
	@echo ">> updating Go dependencies..."
	@go get -u all
	@go mod tidy

.PHONY: all
all: build
