# environment
BUILDDIR          := ${CURDIR}/build
TARBUILDDIR       := ${BUILDDIR}/tar
ARCH              := $(shell go env GOHOSTARCH)
OS                := $(shell go env GOHOSTOS)
GOVER             := $(shell go version | awk '{print $$3}' | tr -d '.')
SIGN_KEY          ?= ${HOME}/.minisign/go-camo.key

# app specific info
APP_NAME          := go-camo
APP_VER           := $(shell git describe --always --tags|sed 's/^v//')
GOPATH            := $(shell go env GOPATH)
VERSION_VAR       := main.ServerVersion

# flags and build configuration
GOBUILD_OPTIONS   := -trimpath
GOTEST_FLAGS      := -cpu=1,2
GOTEST_BENCHFLAGS :=
GOBUILD_DEPFLAGS  := -tags netgo
GOBUILD_LDFLAGS   ?= -s -w
GOBUILD_FLAGS     := ${GOBUILD_DEPFLAGS} ${GOBUILD_OPTIONS} -ldflags "${GOBUILD_LDFLAGS} -X ${VERSION_VAR}=${APP_VER}"

# cross compile defs
CC_BUILD_TARGETS   = go-camo url-tool
CC_BUILD_ARCHES    = darwin/arm64 freebsd/amd64 linux/amd64 linux/arm64 windows/amd64
CC_OUTPUT_TPL     := ${BUILDDIR}/bin/{{.Dir}}.{{.OS}}-{{.Arch}}

# some exported vars (pre-configure go build behavior)
export GO111MODULE=on
export CGO_ENABLED=0
## enable go 1.21 loopvar "experiment"
export GOEXPERIMENT=loopvar

define HELP_OUTPUT
Available targets:
* help                this help (default target)
  clean               clean up
  check               run checks and validators
  test                run tests
  cover               run tests with cover output
  bench               run benchmarks
  build               build all binaries
  man                 build all man pages
  all                 build binaries and man pages
  tar                 build release tarball for host platform only
  cross-tar           cross compile and build release tarballs for all platforms
  release-sign        sign release tarballs with minisign
  release             build and sign release
  update-go-deps      updates go.mod and go.sum files
endef
export HELP_OUTPUT

.PHONY: help clean build test cover bench man man-copy all tar cross-tar setup-check

help:
	@echo "$$HELP_OUTPUT"

clean:
	@rm -rf "${BUILDDIR}"

setup:

setup-check: ${GOPATH}/bin/staticcheck ${GOPATH}/bin/gosec ${GOPATH}/bin/govulncheck

${GOPATH}/bin/staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest

${GOPATH}/bin/gosec:
	go install github.com/securego/gosec/v2/cmd/gosec@latest

${GOPATH}/bin/govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest

${GOPATH}/bin/nilness:
	go install golang.org/x/tools/go/analysis/passes/nilness/cmd/nilness@latest

build: setup
	@[ -d "${BUILDDIR}/bin" ] || mkdir -p "${BUILDDIR}/bin"
	@echo "Building..."
	@echo "...go-camo..."
	@go build ${GOBUILD_FLAGS} -o "${BUILDDIR}/bin/go-camo" ./cmd/go-camo
	@echo "...url-tool..."
	@go build ${GOBUILD_FLAGS} -o "${BUILDDIR}/bin/url-tool" ./cmd/url-tool
	@echo "done!"

test: setup
	@echo "Running tests..."
	@go test -count=1 -cpu=4 -vet=off ${GOTEST_FLAGS} ./...

bench: setup
	@echo "Running benchmarks..."
	@go test -bench="." -run="^$$" -test.benchmem=true ${GOTEST_BENCHFLAGS} ./...

cover: setup
	@echo "Running tests with coverage..."
	@go test -vet=off -cover ${GOTEST_FLAGS} ./...

check: setup setup-check
	@echo "Running checks and validators..."
	@echo "... staticcheck ..."
	@${GOPATH}/bin/staticcheck ./...
	@echo "... go-vet ..."
	@go vet ./...
	@echo "... gosec ..."
	@${GOPATH}/bin/gosec -quiet ./...
	@echo "... govulncheck ..."
	@${GOPATH}/bin/govulncheck ./...
	@echo "... nilness ..."
	@nilness ./...

.PHONY: update-go-deps
update-go-deps:
	@echo ">> updating Go dependencies"
	@go get -u all
	@go mod tidy

${BUILDDIR}/man/%: man/%.adoc
	@[ -d "${BUILDDIR}/man" ] || mkdir -p "${BUILDDIR}/man"
	@asciidoctor -b manpage -a release-version="${APP_VER}" -o $@ $<

man: $(patsubst man/%.adoc,${BUILDDIR}/man/%,$(wildcard man/*.adoc))

tar: all
	@echo "Building tar..."
	@mkdir -p ${TARBUILDDIR}/${APP_NAME}-${APP_VER}/bin
	@mkdir -p ${TARBUILDDIR}/${APP_NAME}-${APP_VER}/man
	@cp ${BUILDDIR}/bin/${APP_NAME} ${TARBUILDDIR}/${APP_NAME}-${APP_VER}/bin/${APP_NAME}
	@cp ${BUILDDIR}/bin/url-tool ${TARBUILDDIR}/${APP_NAME}-${APP_VER}/bin/url-tool
	@cp ${BUILDDIR}/man/*.[1-9] ${TARBUILDDIR}/${APP_NAME}-${APP_VER}/man/
	@tar -C ${TARBUILDDIR} -czf ${TARBUILDDIR}/${APP_NAME}-${APP_VER}.${GOVER}.${OS}-${ARCH}.tar.gz ${APP_NAME}-${APP_VER}
	@rm -rf "${TARBUILDDIR}/${APP_NAME}-${APP_VER}"

cross-tar: man setup
	@echo "Building (cross-compile: ${CC_BUILD_ARCHES})..."
	@(for x in ${CC_BUILD_TARGETS}; do \
		for y in $(subst /,-,${CC_BUILD_ARCHES}); do \
			printf -- "--> %15s: %s\n" "$${y}" "$${x}"; \
			GOOS="$${y%%-*}"; \
			GOARCH="$${y##*-}"; \
			EXT=""; \
			if echo "$${y}" | grep -q 'windows-'; then EXT=".exe"; fi; \
			env GOOS=$${GOOS} GOARCH=$${GOARCH} go build ${GOBUILD_FLAGS} -o "${BUILDDIR}/bin/$${x}.$${GOOS}-$${GOARCH}$${EXT}" ./cmd/$${x}; \
		done; \
	done)

	@echo "Creating tar archives..."
	@(for x in $(subst /,-,${CC_BUILD_ARCHES}); do \
		printf -- "--> %15s\n" "$${x}"; \
		EXT=""; \
		if echo "$${x}" | grep -q 'windows-'; then EXT=".exe"; fi; \
		XDIR="${GOVER}.$${x}"; \
		ODIR="${TARBUILDDIR}/$${XDIR}/${APP_NAME}-${APP_VER}"; \
		mkdir -p "$${ODIR}/bin"; \
		mkdir -p "$${ODIR}/man"; \
		for t in ${CC_BUILD_TARGETS}; do \
			cp ${BUILDDIR}/bin/$${t}.$${x}$${EXT} $${ODIR}/bin/$${t}$${EXT}; \
		done; \
		cp ${BUILDDIR}/man/*.[1-9] $${ODIR}/man/; \
		tar -C ${TARBUILDDIR}/$${XDIR} -czf ${TARBUILDDIR}/${APP_NAME}-${APP_VER}.$${XDIR}.tar.gz ${APP_NAME}-${APP_VER}; \
		rm -rf "${TARBUILDDIR}/$${XDIR}/"; \
	done)

	@echo "done!"

release-sign:
	@echo "signing release tarballs"
	@(cd build/tar; shasum -a 256 go-camo-*.tar.gz > SHA256; \
	  minisign -S -s ${SIGN_KEY} \
	    -c "go-camo-${APP_VER} SHA256" \
	    -t "go-camo-${APP_VER} SHA256" \
	    -x SHA256.sig -m SHA256; \
	 )

release: cross-tar release-sign
all: build man
