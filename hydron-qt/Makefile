.PHONY: linux

linux:
	go generate
	qtdeploy build linux

# TODO
# cross_win64:
# 	QT_MXE_ARCH=amd64 QT_MXE_STATIC=true QT_PKG_CONFIG=true \
# 	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
# 	go generate
# 	QT_MXE_ARCH=amd64 QT_MXE_STATIC=true QT_PKG_CONFIG=true \
# 	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
# 	PKG_CONFIG=$(MXE_ROOT)/bin/$(MXE_TARGET)-pkg-config \
# 	PKG_CONFIG_LIBDIR=$(MXE_ROOT)/$(MXE_TARGET)/lib/pkgconfig \
# 	PKG_CONFIG_PATH=$(MXE_ROOT)/$(MXE_TARGET)/lib/pkgconfig \
# 	qtdeploy --ldflags '-extldflags "-static"' build windows

clean:
	rm -rf moc* rcc* deploy linux windows
