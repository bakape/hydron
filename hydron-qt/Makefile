flags=""
ifeq ($(OS), Windows_NT)
	flags=-ldflags '-L/mingw64/lib'
endif

build:
	go generate -v
	qtdeploy $(flags) build desktop

setup:
	go get -u -v github.com/therecipe/qt/cmd/...
	qtsetup prep
	qtsetup generate
	qtsetup install

clean:
	rm -rf moc* rcc* deploy linux windows
