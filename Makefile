version=$(shell git describe --tags)
binary=hydron
ifeq ($(OS), Windows_NT)
	export version:=win_amd64-$(version)
	export GOPATH=$(HOME)/go
	export PATH:=$(PATH):/c/Go/bin:$(GOPATH)/bin
	binary=hydron.exe
else ifeq ($(UNAME_S),Darwin)
	export version:=osx_amd64-$(version)
else
	export version:=linux_amd64-$(version)
	export deploy_dir=linux
endif

all: cli qt

setup_mingw:
	pacman -Su --noconfirm --needed mingw-w64-x86_64-qt-creator mingw-w64-x86_64-qt5-static mingw-w64-x86_64-gcc mingw-w64-x86_64-pkg-config mingw-w64-x86_64-ffmpeg mingw-w64-x86_64-graphicsmagick zip
	pacman -Scc --noconfirm

cli:
	go build -v
	mkdir -p build
	mv $(binary) build/

clean:
	rm -rf hydron hydron.exe build
	cd hydron-qt; qmake
	$(MAKE) -C hydron-qt clean

qt:
	cd hydron-qt && qmake "CONFIG+=c++17 qtquickcompiler reduce-relocations ltcg"
	$(MAKE) -C hydron-qt
	mkdir -p build
	cp -f hydron-qt/hydron-qt build
