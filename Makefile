version_base=$(shell git describe --tags)
version=$(version_base)
binary=hydron
WIN_ARCH=amd64
ifeq ($(OS), Windows_NT)
	version:=win_x86_64-$(version)
	export GOPATH=$(HOME)/go
	export PATH:=$(PATH):/c/Go/bin:$(GOPATH)/bin
	binary=hydron.exe
else ifeq ($(UNAME_S),Darwin)
	version:=osx_x86_64-$(version)
else
	version:=linux_x86_64-$(version)
endif

# Path to and target for the MXE cross environment for cross-compiling to
# win_amd64. Default value is the debian x86-static install path.
MXE_ROOT=$(HOME)/src/mxe/usr
MXE_TARGET=x86_64-w64-mingw32.static

.PHONY: client all generate

all: generate client
	go build -v

client:
	cp client/main.js www/main.js
	cp client/import.js www/import.js
	node_modules/.bin/lessc --clean-css client/main.less www/main.css
	go get github.com/pyros2097/go-embed
	go-embed --input www --output assets/assets.go

generate:
	go get github.com/valyala/quicktemplate/qtc
	go generate ./templates

# Cross-compile from Unix into a Windows x86_64 static binary
# Depends on:
# 	mxe-x86-64-w64-mingw32.static-gcc
# 	mxe-x86-64-w64-mingw32.static-libidn
# 	mxe-x86-64-w64-mingw32.static-ffmpeg
#   mxe-x86-64-w64-mingw32.static-graphicsmagick
#
# To cross-compile for windows-x86 use:
# make cross_compile_windows WIN_ARCH=386 MXE_TARGET=i686-w64-mingw32.static
cross_compile_windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=$(WIN_ARCH) \
	CC=$(MXE_ROOT)/bin/$(MXE_TARGET)-gcc \
	CXX=$(MXE_ROOT)/bin/$(MXE_TARGET)-g++ \
	PKG_CONFIG=$(MXE_ROOT)/bin/$(MXE_TARGET)-pkg-config \
	PKG_CONFIG_LIBDIR=$(MXE_ROOT)/$(MXE_TARGET)/lib/pkgconfig \
	PKG_CONFIG_PATH=$(MXE_ROOT)/$(MXE_TARGET)/lib/pkgconfig \
	CGO_LDFLAGS_ALLOW='-mconsole' \
	go build -v -a -o hydron.exe --ldflags '-extldflags "-static" -H=windowsgui'

cross_package_windows: cross_compile_windows
	zip -r -9 hydron-win_$(WIN_ARCH)-$(version_base).zip hydron.exe

clean:
	rm -rf hydron hydron.exe hydron-*.zip
