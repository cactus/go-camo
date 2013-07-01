
BUILDDIR          := ${CURDIR}/build
GOPATH            := ${BUILDDIR}
RPMBUILDDIR       := ${BUILDDIR}/rpm
ARCH              := $(shell uname -m)
FPM_VERSION       := $(shell gem list fpm|grep fpm|sed -E 's/fpm \((.*)\)/\1/g')
FPM_OPTIONS       :=
GOCAMO_VER        := $(shell grep -F 'ServerVersion =' ./camoproxy/vars.go |awk -F\" '{print $$2}')
ITERATION         := 1

.PHONY: help clean build man rpm all

help:
	@echo "Available targets:"
	@echo "  help                this help"
	@echo "  clean               clean up"
	@echo "  all                 build binaries and man pages"
	@echo "  build               build all"
	@echo "  build-go-camo       build go-camo"
	@echo "  build-url-tool      build url tool"
	@echo "  build-simple-server build simple server"
	@echo "  man                 build all man pages"
	@echo "  man-go-camo         build go-camo man pages"
	@echo "  man-url-tool        build url-tool man pages"
	@echo "  man-simple-server   build simple-server man pages"
	@echo "  rpm                 build rpm"

clean:
	-rm -rf "${BUILDDIR}"
	
build-setup:
	@mkdir -p "${GOPATH}/src/github.com/cactus"
	@test -d "${GOPATH}/src/github.com/cactus/go-camo" || ln -s "${CURDIR}" "${GOPATH}/src/github.com/cactus/go-camo"

build-go-camo: build-setup
	@env GOPATH="${GOPATH}" go get -d github.com/cactus/go-camo
	@env GOPATH="${GOPATH}" go install -v github.com/cactus/go-camo

build-url-tool: build-setup
	@env GOPATH="${GOPATH}" go get -d github.com/cactus/go-camo/url-tool
	@env GOPATH="${GOPATH}" go install -v github.com/cactus/go-camo/url-tool

build-simple-server: build-setup
	@env GOPATH="${GOPATH}" go get -d github.com/cactus/go-camo/simple-server
	@env GOPATH="${GOPATH}" go install -v github.com/cactus/go-camo/simple-server

build-devweb: build-setup
	@env GOPATH="${GOPATH}" go get -d github.com/cactus/go-camo/go-camo-devweb
	@env GOPATH="${GOPATH}" go install -v github.com/cactus/go-camo/go-camo-devweb

man-setup:
	@mkdir -p "${BUILDDIR}/man/man1"

man-camo: man-setup
	@pod2man -s 1 -r "go-camo ${GOCAMO_VER}" -n go-camo --center="go-camo manual" man/go-camo.pod |gzip > build/man/man1/go-camo.1.gz

man-url-tool: man-setup
	@pod2man -s 1 -r "url-tool ${GOCAMO_VER}" -n url-tool --center="go-camo manual" man/url-tool.pod |gzip > build/man/man1/url-tool.1.gz

man-simple-server: man-setup
	@pod2man -s 1 -r "simple-server ${GOCAMO_VER}" -n simple-server --center="go-camo manual" man/simple-server.pod |gzip > build/man/man1/simple-server.1.gz

rpm: all
	@mkdir -p ${RPMBUILDDIR}/usr/local/bin
	@mkdir -p ${RPMBUILDDIR}/usr/local/share/man/man1
	@cp ${BUILDDIR}/bin/* ${RPMBUILDDIR}/usr/local/bin
	@cp ${BUILDDIR}/man/man1/* ${RPMBUILDDIR}/usr/local/share/man/man1
	fpm -s dir -t rpm -n go-camo \
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
