# hydron
Media tagger and organizer backend and GUI frontend.
Hydron aims to be a much faster alternative to the
[Hydrus Network](https://github.com/hydrusnetwork/hydrus) and is currently in
early development.

Platforms: Linux, OSX, Win64

## Installation

<details><summary>Windows</summary>

While it is possible to compile binaries on Windows with MinGW/MSYS2 similar
to how you would on Unix-like systems, it is a huge pain in the ass.
Just download statically precompiled binaries from the
[release page](https://github.com/bakape/hydron/releases).

</details>

<details><summary>Linux / OS X</summary>

- Install dependencies listed below.
On a Debian-based system those would the following packages or similar:

`golang build-essential pkg-config libpth-dev libavcodec-dev libavutil-dev libavformat-dev libgraphicsmagick1-dev qtdeclarative5-dev qt5-default qt5-qmake`

- Set up a Go workspace (not needed with Go >= 1.8)

`mkdir ~/go; echo 'export GOPATH=~/go' >> ~/.bashrc; . ~/.bashrc`

- Run `make qt`. The binaries will be located in `./build`. The GUI can be
launched with `hydron-qt.sh`.

- To build only the CLI client run `go get && go build`

</details>

### Build dependencies
* [Go](https://golang.org/doc/install) >= 1.7
* GCC or Clang
* pkg-config
* pthread
* ffmpeg >= 3.0 libraries (libswscale, libavcodec, libavutil, libavformat)
* GraphicsMagick
* Qt5 >= 5.7
* qmake
