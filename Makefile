version_base=$(shell git describe --tags)
version=$(version_base)
binary=hydron
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

.PHONY: client

all: client generate
	go build

client:
	cp client/main.js www/main.js
	node_modules/.bin/lessc --clean-css client/main.less www/main.css
	go-embed --input www --output assets/assets.go

generate:
	go generate ./templates

# Cross-compile from Unix into a Windows x86_64 static binary
# Depends on:
# 	mxe-x86-64-w64-mingw32.static-gcc
# 	mxe-x86-64-w64-mingw32.static-libidn
# 	mxe-x86-64-w64-mingw32.static-ffmpeg
#   mxe-x86-64-w64-mingw32.static-graphicsmagick
cross_compile_windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
	CC=$(MXE_ROOT)/bin/$(MXE_TARGET)-gcc \
	CXX=$(MXE_ROOT)/bin/$(MXE_TARGET)-g++ \
	PKG_CONFIG=$(MXE_ROOT)/bin/$(MXE_TARGET)-pkg-config \
	PKG_CONFIG_LIBDIR=$(MXE_ROOT)/$(MXE_TARGET)/lib/pkgconfig \
	PKG_CONFIG_PATH=$(MXE_ROOT)/$(MXE_TARGET)/lib/pkgconfig \
	CGO_LDFLAGS_ALLOW='-mconsole' \
	go build -v -a -o hydron.exe --ldflags '-extldflags "-static"'

cross_package_windows: client cross_compile_windows
	zip -r -9 hydron-win_x86_64-$(version_base).zip hydron.exe

clean:
	rm -rf hydron hydron.exe hydron-*.zip

# all: cli qt

# package: all
# 	zip -9 hydron-$(version).zip build/*

# setup_mingw:
# 	pacman -Su --noconfirm --needed mingw-w64-x86_64-qt-creator mingw-w64-x86_64-qt5-static mingw-w64-x86_64-gcc mingw-w64-x86_64-pkg-config mingw-w64-x86_64-ffmpeg-static mingw-w64-x86_64-graphicsmagick-static zip
# 	pacman -Scc --noconfirm

# cli:
# 	go build -v
# 	mkdir -p build
# 	cp -f $(binary) build

# qt:
# 	cd hydron-qt && qmake "CONFIG+=c++17 qtquickcompiler static reduce-relocations ltcg"
# 	$(MAKE) -C hydron-qt
# 	mkdir -p build
# 	cp -f hydron-qt/hydron-qt build
