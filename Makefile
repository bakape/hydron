# Path to and target for the MXE cross environment for cross-compiling to
# win_amd64. Default value is the debian x86-static install path.
MXE_ROOT=/usr/lib/mxe/usr
MXE_TARGET=x86_64-w64-mingw32.static

# Cross-compile from Unix into a Windows_amd64 static binary
# Needs Go checkout dfbbe06a205e7048a8541c4c97b250c24c40db96 or later. At the
# moment of writing this change is not released yet. Should probably make it
# into Go 1.7.1.
# Depends on:
# 	mxe-x86-64-w64-mingw32.static-gcc
# 	mxe-x86-64-w64-mingw32.static-libidn
# 	mxe-x86-64-w64-mingw32.static-ffmpeg
#   mxe-x86-64-w64-mingw32.static-graphicsmagick
cross_compile_win_amd64:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
	CC=$(MXE_ROOT)/bin/$(MXE_TARGET)-gcc \
	PKG_CONFIG=$(MXE_ROOT)/bin/$(MXE_TARGET)-pkg-config \
	PKG_CONFIG_LIBDIR=$(MXE_ROOT)/$(MXE_TARGET)/lib/pkgconfig \
	PKG_CONFIG_PATH=$(MXE_ROOT)/$(MXE_TARGET)/lib/pkgconfig \
	go build -v -a -o hydron.exe --ldflags '-extldflags "-static"'

