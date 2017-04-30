build:
	go generate -v
	qtdeploy build desktop

setup:
	go get -u -v github.com/therecipe/qt/cmd/...
	qtsetup prep
	qtsetup generate
	qtsetup install

clean:
	rm -rf moc* rcc* deploy linux windows
