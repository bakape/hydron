# Path to and target for the MXE cross environment for cross-compiling to
# win_amd64. Default value is the Debian x86-static install path.
MXE_ROOT=/usr/lib/mxe/usr
MXE_TARGET=x86_64-w64-mingw32.static

# Cross-compile from Unix into a Windows_amd64 static binary
# Needs Go >= 1.8.
# Depends on (on Debian-based distros):
# 	mxe-x86-64-w64-mingw32.static-gcc
# 	mxe-x86-64-w64-mingw32.static-libidn
# 	mxe-x86-64-w64-mingw32.static-ffmpeg
#   mxe-x86-64-w64-mingw32.static-graphicsmagick
cross_win64_cli:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
	CC=$(MXE_ROOT)/bin/$(MXE_TARGET)-gcc \
	PKG_CONFIG=$(MXE_ROOT)/bin/$(MXE_TARGET)-pkg-config \
	PKG_CONFIG_LIBDIR=$(MXE_ROOT)/$(MXE_TARGET)/lib/pkgconfig \
	PKG_CONFIG_PATH=$(MXE_ROOT)/$(MXE_TARGET)/lib/pkgconfig \
	go build -v -a -o hydron.exe --ldflags '-extldflags "-static"'

clean:
	$(MAKE) -C hydron-qt/libwrapper clean
	cd hydron-qt; qmake
	$(MAKE) -C hydron-qt clean
	rm -rf dist

qt:
	$(MAKE) -C hydron-qt/libwrapper
	cd hydron-qt; qmake
	$(MAKE) -C hydron-qt
	rm -rf dist
	mkdir -p dist
	cp hydron-qt/libwrapper/libwrapper.so hydron-qt/hydron-qt dist
	cp scripts/unix-launch.sh dist/hydron-qt.sh
