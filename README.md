<p align="center"><img src="logo/Hydron.png" alt="Hydron" height="150px"></p>

# hydron
Media tagger and organizer backend and GUI frontend.
Hydron aims to be a much faster alternative to the
[Hydrus Network](https://github.com/hydrusnetwork/hydrus) and is currently in
early development.

Platforms: Linux, OSX, Win64

## Running

1. `hydron` to start the server. See `hydron -h` for more options.
2. Navigate to "http://localhost:8010" in a web browser

### Runtime dependecies
* ffmpeg >= 3.0 libraries (libswscale, libavcodec, libavutil, libavformat)
* GraphicsMagick++

## Building

`go get github.com/bakape/hydron`

### Build dependencies
* [Go](https://golang.org/doc/install) >= 1.10
* C11 and C++17 compilers
* pkg-config
* pthread
* ffmpeg >= 3.0 libraries (libswscale, libavcodec, libavutil, libavformat)
* GraphicsMagick++
