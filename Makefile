version=$(shell git describe --tags)
binary=hydron
ifeq ($(OS), Windows_NT)
	version:=win_amd64-$(version)
	export GOPATH=$(HOME)/go
	export PATH:=$(PATH):/c/Go/bin:$(GOPATH)/bin
	binary=hydron.exe
else ifeq ($(UNAME_S),Darwin)
	version:=osx_amd64-$(version)
else
	version:=linux_amd64-$(version)
endif

all: client generate

client: css js

css:
	node_modules/.bin/lessc --clean-css client/main.less www/main.css

js:
	node_modules/.bin/uglifyjs client/main.js -o www/main.js

generate:
	go generate ./templates

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

# clean:
# 	rm -rf hydron hydron.exe hydron-*.zip build hydron-qt/moc_* hydron-qt/.o hydron-qt/hydron-qt

# qt:
# 	cd hydron-qt && qmake "CONFIG+=c++17 qtquickcompiler static reduce-relocations ltcg"
# 	$(MAKE) -C hydron-qt
# 	mkdir -p build
# 	cp -f hydron-qt/hydron-qt build
