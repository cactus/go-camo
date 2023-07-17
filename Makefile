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
GOBUILD_DEPFLAGS  := -tags netgo
GOBUILD_LDFLAGS   ?= -s -w
GOBUILD_FLAGS     := ${GOBUILD_DEPFLAGS} ${GOBUILD_OPTIONS} -ldflags "${GOBUILD_LDFLAGS} -X ${VERSION_VAR}=${APP_VER}"

# cross compile defs
CC_BUILD_ARCHES    = darwin/amd64 darwin/arm64 freebsd/amd64 linux/amd64 linux/arm64 windows/amd64
CC_OUTPUT_TPL     := ${BUILDDIR}/bin/{{.Dir}}.{{.OS}}-{{.Arch}}

# some exported vars (pre-configure go build behavior)
export GO111MODULE=on
export CGO_ENABLED=0
## enable go 1.21 loopvar "experiment"
#export GOEXPERIMENT=loopvar

define HELP_OUTPUT
Available targets:
  help                this help
  clean               clean up
  all                 build binaries and man pages
  test                run tests
  cover               run tests with cover output
  build               build all binaries
  man                 build all man pages
  tar                 build release tarball
  cross-tar           cross compile and build release tarballs
endef
export HELP_OUTPUT

.PHONY: help clean build test cover man man-copy all tar cross-tar setup-check setup-gox

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

setup-gox: $GOPATH/bin/gox

$GOPATH/bin/gox:
	go install github.com/mitchellh/gox@latest

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
	@go test -count=1 -vet=off ${GOTEST_FLAGS} ./...

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

.PHONY: update-go-deps
update-go-deps:
	@echo ">> updating Go dependencies"
	@for m in $$(go list -mod=readonly -m -f '{{ if and (not .Indirect) (not .Main)}}{{.Path}}{{end}}' all); do \
		go get $$m; \
	done
	go mod tidy

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

cross-tar: man setup setup-gox
	@echo "Building (cross-compile: ${CC_BUILD_ARCHES})..."
	@(for x in go-camo url-tool; do \
		echo "...$${x}..."; \
		env GOFLAGS="${GOBUILD_OPTIONS}" gox \
			-gocmd="go" \
			-output="${CC_OUTPUT_TPL}" \
			-osarch="${CC_BUILD_ARCHES}" \
			-ldflags "${GOBUILD_LDFLAGS} -X ${VERSION_VAR}=${APP_VER}" \
		${GOBUILD_DEPFLAGS} ./cmd/$${x}; \
		echo; \
	done)

	@echo "...creating tar files..."
	@(for x in $(subst /,-,${CC_BUILD_ARCHES}); do \
		echo "making tar for ${APP_NAME}.$${x}"; \
		EXT=""; \
		if echo "$${x}" | grep -q 'windows-'; then EXT=".exe"; fi; \
		XDIR="${GOVER}.$${x}"; \
		ODIR="${TARBUILDDIR}/$${XDIR}/${APP_NAME}-${APP_VER}"; \
		mkdir -p "$${ODIR}/bin"; \
		mkdir -p "$${ODIR}/man"; \
		cp ${BUILDDIR}/bin/${APP_NAME}.$${x}$${EXT} $${ODIR}/bin/${APP_NAME}$${EXT}; \
		cp ${BUILDDIR}/bin/url-tool.$${x}$${EXT} $${ODIR}/bin/url-tool$${EXT}; \
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
