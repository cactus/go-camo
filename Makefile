
BUILDDIR          := ${CURDIR}/build
GOPATH            := ${BUILDDIR}
RPMBUILDDIR       := ${BUILDDIR}/rpm
ARCH              := $(shell uname -m)
FPM_VERSION       := $(shell gem list fpm|grep fpm|sed -E 's/fpm \((.*)\)/\1/g')
FPM_OPTIONS       ?=
ITERATION         ?= 1

GOCAMO_VER        := $(shell git describe --always --dirty --tags|sed 's/^v//')
VERSION_VAR       := main.ServerVersion
GOBUILD_LDFLAGS   := -ldflags "-X $(VERSION_VAR) $(GOCAMO_VER)"
GOBUILD_FLAGS     ?= -tags netgo

.PHONY: help clean build test cover man rpm all

help:
	@echo "Available targets:"
	@echo "  help                this help"
	@echo "  clean               clean up"
	@echo "  all                 build binaries and man pages"
	@echo "  build               build all"
	@echo "  build-go-camo       build go-camo"
	@echo "  build-url-tool      build url tool"
	@echo "  build-simple-server build simple server"
	@echo "  test                run tests"
	@echo "  cover               run tests with cover output"
	@echo "  man                 build all man pages"
	@echo "  man-go-camo         build go-camo man pages"
	@echo "  man-url-tool        build url-tool man pages"
	@echo "  man-simple-server   build simple-server man pages"
	@echo "  rpm                 build rpm"

clean:
	-rm -rf "${BUILDDIR}"

${GOPATH}/bin/godep:
	@mkdir -p "${GOPATH}/src"
	@echo "Building godep..."
	@env GOPATH="${GOPATH}" go get ${GOBUILD_FLAGS} github.com/kr/godep

build-setup: ${GOPATH}/bin/godep
	@echo "Restoring deps with godep..."
	@env GOPATH="${GOPATH}" ${GOPATH}/bin/godep restore
	@mkdir -p "${GOPATH}/src/github.com/cactus"
	@test -d "${GOPATH}/src/github.com/cactus/go-camo" || ln -s "${CURDIR}" "${GOPATH}/src/github.com/cactus/go-camo"

build-go-camo: build-setup
	@echo "Building go-camo..."
	@env GOPATH="${GOPATH}" go install ${GOBUILD_FLAGS} ${GOBUILD_LDFLAGS} github.com/cactus/go-camo

build-url-tool: build-setup
	@echo "Building url-tool..."
	@env GOPATH="${GOPATH}" go install ${GOBUILD_FLAGS} ${GOBUILD_LDFLAGS} github.com/cactus/go-camo/url-tool

build-simple-server: build-setup
	@echo "Building simple-server..."
	@env GOPATH="${GOPATH}" go install ${GOBUILD_FLAGS} ${GOBUILD_LDFLAGS} github.com/cactus/go-camo/simple-server

test: build-setup
	@echo "Running tests..."
	@env GOPATH="${GOPATH}" go test ./camo/...

cover: build-setup
	@echo "Running tests with coverage..."
	@env GOPATH="${GOPATH}" go test -cover ./camo/...

man-setup:
	@mkdir -p "${BUILDDIR}/man/man1"

man-camo: man-setup
	@echo "Building go-camo manpage..."
	@pod2man -s 1 -r "go-camo ${GOCAMO_VER}" -n go-camo \
		--center="go-camo manual" man/go-camo.pod | \
		gzip > build/man/man1/go-camo.1.gz

man-url-tool: man-setup
	@echo "Building url-tool manpage..."
	@pod2man -s 1 -r "url-tool ${GOCAMO_VER}" -n url-tool \
		--center="go-camo manual" man/url-tool.pod | \
		gzip > build/man/man1/url-tool.1.gz

man-simple-server: man-setup
	@echo "Building simple-server manpage..."
	@pod2man -s 1 -r "simple-server ${GOCAMO_VER}" -n simple-server \
		--center="go-camo manual" man/simple-server.pod | \
		gzip > build/man/man1/simple-server.1.gz

rpm: all
	@echo "Building rpm..."
	@mkdir -p ${RPMBUILDDIR}/usr/local/bin
	@mkdir -p ${RPMBUILDDIR}/usr/local/share/man/man1
	@cp ${BUILDDIR}/bin/* ${RPMBUILDDIR}/usr/local/bin
	@cp ${BUILDDIR}/man/man1/* ${RPMBUILDDIR}/usr/local/share/man/man1
	@fpm -s dir -t rpm -n go-camo \
		--url "https://github.com/cactus/go-camo" \
		-v "${GOCAMO_VER}" \
		--iteration "${ITERATION}" \
		-C "${RPMBUILDDIR}" \
		${FPM_OPTIONS} \
		usr/local/bin usr/local/share/man/man1
	@mv *.rpm ${BUILDDIR}/

build: build-go-camo build-url-tool build-simple-server
man: man-camo man-url-tool man-simple-server
all: build man
