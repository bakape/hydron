# # Differentiate between Unix and MSYS2 builds
# ifeq ($(OS), Windows_NT)
# 	export PKG_CONFIG_PATH:=$(PKG_CONFIG_PATH):/mingw64/lib/pkgconfig/
# 	export PKG_CONFIG_LIBDIR=/mingw64/lib/pkgconfig/
# 	export PATH:=$(PATH):/mingw64/bin/
# endif

version=$(shell git describe --tags)
ifeq ($(OS), Windows_NT)
	export QT_MSYS2=true
	export version:=win_amd64-$(version)
	export deploy_dir=windows
else ifeq ($(UNAME_S),Darwin)
	export QT_HOMEBREW=true
	export version:=osx_amd64-$(version)
	export deploy_dir=osx
else
	export version:=linux_amd64-$(version)
	export deploy_dir=linux
endif

setup:
	go get -u -v github.com/bakape/hydron
	$(MAKE) -C hydron-qt setup

all: cli qt
	rm -rf build
	mkdir -p build
	mv hydron build
	cp -r hydron-qt/deploy/$(deploy_dir)/* build

cli:
	go build -v

qt:
	$(MAKE) -C hydron-qt build

clean:
	rm -rf build hydron hydron.exe
	$(MAKE) -C hydron-qt clean

