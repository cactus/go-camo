# environment
BUILDDIR          := ${CURDIR}/build
TARBUILDDIR       := ${BUILDDIR}/tar
ARCH              := $(shell go env GOHOSTARCH)
OS                := $(shell go env GOHOSTOS)
GOVER             := $(shell go version | awk '{print $$3}' | tr -d '.')
SIGN_KEY          ?= ${HOME}/.signify/go-camo.sec

# app specific info
APP_NAME          := go-camo
APP_VER           := $(shell git describe --always --tags|sed 's/^v//')
VERSION_VAR       := main.ServerVersion

# flags and build configuration
GOTEST_FLAGS      := -cpu=1,2
GOBUILD_DEPFLAGS  := -tags netgo
GOBUILD_LDFLAGS   ?= -s -w
GOBUILD_FLAGS     := ${GOBUILD_DEPFLAGS} -ldflags "${GOBUILD_LDFLAGS} -X ${VERSION_VAR}=${APP_VER}"

# cross compile defs
CC_BUILD_ARCHES    = darwin/amd64 freebsd/amd64 linux/amd64
CC_OUTPUT_TPL     := ${BUILDDIR}/bin/{{.Dir}}.{{.OS}}-{{.Arch}}

# some exported vars (pre-configure go build behavior)
export GO111MODULE=on
export CGO_ENABLED=0

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

.PHONY: help clean build test cover man man-copy all tar cross-tar

help:
	@echo "$$HELP_OUTPUT"

clean:
	@rm -rf "${BUILDDIR}"

setup:

setup-gox:
	@if [ -z "$(shell which gox)" ]; then \
		echo "* 'gox' command not found."; \
		echo "  install (or otherwise ensure presence in PATH)"; \
		echo "  go get github.com/mitchellh/gox"; \
		exit 1;\
	fi

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
	@env go test ${GOTEST_FLAGS} ./...

cover: setup
	@echo "Running tests with coverage..."
	@env go test -cover ${GOTEST_FLAGS} ./...

${BUILDDIR}/man/%: man/%.mdoc
	@[ -d "${BUILDDIR}/man" ] || mkdir -p "${BUILDDIR}/man"
	@cat $< | sed -E "s#.Os (.*) VERSION#.Os \1 ${APP_VER}#" > $@

man: $(patsubst man/%.mdoc,${BUILDDIR}/man/%,$(wildcard man/*.1.mdoc))

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
	@echo "...go-camo..."
	@gox -gocmd="go" -output="${CC_OUTPUT_TPL}" -osarch="${CC_BUILD_ARCHES}" ${GOBUILD_FLAGS} ./cmd/go-camo
	@echo

	@echo "...url-tool..."
	@gox -gocmd="go" -output="${CC_OUTPUT_TPL}" -osarch="${CC_BUILD_ARCHES}" ${GOBUILD_FLAGS} ./cmd/url-tool
	@echo

	@echo "...creating tar files..."
	@(for x in $(subst /,-,${CC_BUILD_ARCHES}); do \
		echo "making tar for ${APP_NAME}.$${x}"; \
		XDIR="${GOVER}.$${x}"; \
		ODIR="${TARBUILDDIR}/$${XDIR}/${APP_NAME}-${APP_VER}"; \
		mkdir -p $${ODIR}/{bin,man}/; \
		cp ${BUILDDIR}/bin/${APP_NAME}.$${x} $${ODIR}/bin/${APP_NAME}; \
		cp ${BUILDDIR}/bin/url-tool.$${x} $${ODIR}/bin/url-tool; \
		cp ${BUILDDIR}/man/*.[1-9] $${ODIR}/man/; \
		tar -C ${TARBUILDDIR}/$${XDIR} -czf ${TARBUILDDIR}/${APP_NAME}-${APP_VER}.$${XDIR}.tar.gz ${APP_NAME}-${APP_VER}; \
		rm -rf "${TARBUILDDIR}/$${XDIR}/"; \
	done)

	@echo "done!"

release-sign:
	@echo "signing release tarballs"
	@(cd build/tar; shasum -a 256 go-camo-*.tar.gz > SHA256; \
	  signify -S -s ${SIGN_KEY} -m SHA256; \
	  sed -i.bak -E 's#^(.*:).*#\1 go-camo-${APP_VER} SHA256#' SHA256.sig; \
	  rm -f SHA256.sig.bak; \
	 )

all: build man
