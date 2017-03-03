# hydron
Media tagger and organizer backend.
Hydron aims to be a much faster alternative to the
[Hydrus Network](https://github.com/hydrusnetwork/hydrus) and is currently in
early development.

Platforms: Linux, OSX, Win64

##Installation

<details>
	<summary>Windows</summary>
	While it is possible to compile binaries on Windows with MinGW/MSYS2 similar
	to how you would on Unix-like systems, it is a huge pain in the ass.
	Just download statically precompiled binaries from the
	<a href=https://github.com/bakape/hydron/releases>release page</a>.
</details>
<details>
	<summary>Linux / OS X</summary>
	1) Install dependencies listed below. On a Debian-based system those would
	the following packages or similar:
	`golang build-essential pkg-config libpth-dev libavcodec-dev libavutil-dev
	libavformat-dev libgraphicsmagick1-dev`
	2) Set up a Go workspace (not needed with Go >= 1.8)
	```bash
	mkdir ~/go
	echo "export GOPATH=~/go" >> ~/.bashrc
	. ~/.bashrc
	```
	3) Add Go bin directory to your path
	```bash
	echo "export PATH=$PATH:~/go/bin" >> ~/.bashrc
	. ~/.bashrc
	```
	4) Download and install Hydron with `go get github.com/bakape/hydron`
</details>

###Build dependencies
* [Go](https://golang.org/doc/install) >=1.8
* GCC or Clang
* pkg-config
* pthread
* ffmpeg >= 3.0 libraries (libavcodec, libavutil, libavformat)
* GraphicsMagick
