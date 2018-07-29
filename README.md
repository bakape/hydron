<h1>
   <img src="logo/Hydron.png" alt="Hydron" height="25px"> hydron
</h1>
Media tagger and organizer backend and GUI frontend.
Hydron aims to be a much faster alternative to the
[Hydrus Network](https://github.com/hydrusnetwork/hydrus) and is currently in
early development.

Platforms: Linux, OSX, Win32, Win64

## Running

1. `hydron` to start the server. See `hydron -h` for more options.
2. Navigate to "http://localhost:8010" in a web browser

### Runtime dependecies
* ffmpeg >= 3.0 libraries (libswscale, libavcodec, libavutil, libavformat)
* GraphicsMagick

### DBMS settings

By default hydron uses SQLite3 but you might want to switch to a diffrent
DBMS like PostgreSQL for performance reasons. To do this copy the sample config
file `docs/db_conf.json` into either `~/.hydron/` or `%APPDATA%\hydron`,
depending on your OS, and configure appropriately.

## Building

`go get github.com/bakape/hydron`

### Build dependencies
* [Go](https://golang.org/doc/install) >= 1.10
* C11 compiler
* pkg-config
* pthread
* ffmpeg >= 3.0 libraries (libswscale, libavcodec, libavutil, libavformat)
* GraphicsMagick
* Git

On Debian-based systems these can be installed with the following or similar:
`apt-get install -y build-essential pkg-config libpth-dev libavcodec-dev libavutil-dev libavformat-dev libswscale-dev libgraphicsmagick1-dev ghostscript git golang`
