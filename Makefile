# base path used to install.
DESTDIR ?= /usr/local

# command name
COMMANDS=benchmark-yamux-app benchmark-yamux-server benchmark-yamux-server-inguest

# binaries
BINARIES=$(addprefix bin/,$(COMMANDS))

# go build command
GO_BUILD_BINARY=go build -tags netgo -ldflags '-w -extldflags "-static"' -o $@ ./$<

.PHONY: build binaries

build: binaries

# force to rebuild
REBUILD:

# build a binary from a cmd.
bin/%: cmd/% REBUILD
	$(GO_BUILD_BINARY)

# build binaries
binaries: $(BINARIES)

# install binaries
install:
	@echo "$@ $(DESTDIR)/$(BINARIES)"
	@mkdir -p $(DESTDIR)/bin
	@install $(BINARIES) $(DESTDIR)/bin

# install_into_guest_rootfs installs the yamux server into guest rootfs.
install_into_guest_rootfs:
	@mkdir -p tmp/rootfs
	@mount tmp/rootfs.ext4 tmp/rootfs
	@install bin/* tmp/rootfs/usr/local/bin/
	@umount tmp/rootfs

# download_guest_rootfs downloads pre-build rootfs from firecracker community.
download_guest_rootfs:
	@mkdir -p tmp
	@echo "download rootfs from firecracker community"
	@wget https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/rootfs/bionic.rootfs.ext4 \
		-O tmp/rootfs.ext4

clean:
	@rm -rf ./bin
