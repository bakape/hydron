# hydron
Media tagger and organizer backend.
Hydron aims to be a much faster alternative to the
[Hydrus Network](https://github.com/hydrusnetwork/hydrus) and is currently in
early development.

Platforms: Linux, OSX, Win64

## Building

<!--<details>
	<summary>Windows</summary>
	While it is possible to compile binaries on Windows with MinGW/MSYS2 similar
	to how you would on Unix-like systems, it is a huge pain in the ass.
	Just download statically precompiled binaries from the
	<a href=https://github.com/bakape/hydron/releases>release page</a>.
</details>-->

<details><summary>OS X</summary>

- Install XCode
- Follow the Linux guide

</details>

<details><summary>Linux</summary>

- Install [QT SDK](https://download.qt.io/official_releases/qt/5.8/5.8.0/)
- Install package manager dependencies.
On a Debian-based system those would the following packages or similar:
`golang build-essential pkg-config libpth-dev libavcodec-dev libavutil-dev
libavformat-dev libgraphicsmagick1-dev`
- Set up a Go workspace (not needed with Go >= 1.8)
```
mkdir ~/go
echo 'export GOPATH=~/go' >> ~/.bashrc
. ~/.bashrc
```
- Add Go bin directory to your path
```
echo 'export PATH=$PATH:~/go/bin' >> ~/.bashrc
. ~/.bashrc
```
- Fetch and build dependencies with `make setup`
- Build hydron with `make all`.
The binaries will be located in the build directory.

</details>

### Build dependencies
* [Go](https://golang.org/doc/install) >= 1.7
* GCC or Clang
* pkg-config
* pthread
* ffmpeg >= 3.0 libraries (libavcodec, libavutil, libavformat)
* GraphicsMagick
* Qt5
