version=$(shell git describe --tags)
binary=hydron
ifeq ($(OS), Windows_NT)
	export QT_MSYS2=true
	export QT_MSYS2_STATIC=true
	export version:=win_amd64-$(version)
	export GOPATH=$(HOME)/go
	export PATH:=$(PATH):/c/Go/bin:$(GOPATH)/bin
	deploy_dir=windows
	binary=hydron.exe
else ifeq ($(UNAME_S),Darwin)
	export QT_HOMEBREW=true
	export version:=osx_amd64-$(version)
	deploy_dir=osx
else
	export version:=linux_amd64-$(version)
	export deploy_dir=linux
endif

setup:
ifeq ($(OS), Windows_NT)
	pacman -Su --noconfirm --needed mingw-w64-x86_64-qt-creator mingw-w64-x86_64-qt5-static mingw-w64-x86_64-gcc mingw-w64-x86_64-pkg-config mingw-w64-x86_64-ffmpeg mingw-w64-x86_64-graphicsmagick zip
	pacman -Scc --noconfirm
endif
	go get -u -v github.com/bakape/hydron
	$(MAKE) -C hydron-qt setup

all: cli qt
	rm -rf build
	mkdir -p build
	mv $(binary) build
	cp -r hydron-qt/deploy/$(deploy_dir)/* build

cli:
	go build -v

qt:
	$(MAKE) -C hydron-qt build

clean:
	rm -rf hydron hydron.exe dist
	$(MAKE) -C hydron-qt/libwrapper clean
	cd hydron-qt; qmake
	$(MAKE) -C hydron-qt clean

qt:
	$(MAKE) -C hydron-qt/libwrapper
	cd hydron-qt; qmake
	$(MAKE) -C hydron-qt
	rm -rf dist
	mkdir -p dist
	cp hydron-qt/libwrapper/libwrapper.so hydron-qt/hydron-qt dist
	cp scripts/unix-launch.sh dist/hydron-qt.sh
