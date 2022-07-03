## benchmark yamux

It is used to emulate kata-container v1 which uses virtio-serial-pci and proxy
to manage kata-agent in the guest.

### How to test

1. Download Guest RootFS from firecracker community.

```bash
make download_guest_rootfs
```

2. Build the yamux server and application.

```bash
make binaries
```

3. Install the yamux server and application into Guest RootFS.

```bash
# use sudo for mount/umount 
sudo make install_into_guest_rootfs
```

4. Run the Guest OS

```bash
qemu-system-x86_64 \
  --nodefaults --no-reboot --display none -serial mon:stdio \
  -cpu kvm64 -enable-kvm -smp "${NPROC}" -m 4G \
  -drive file=tmp/rootfs.ext4,format=raw,index=1,media=disk,if=virtio,cache=none \
  -device virtio-serial-pci,disable-modern=false,id=serial0,romfile= \
  -device virtserialport,chardev=charch0,id=channel0,name=agent.channel.0 \
  -chardev socket,id=charch0,path=/tmp/benchmark-yamux-server.sock,server,nowait \
  -kernel kernel-v5.18.img -append "root=/dev/vda rw console=ttyS0,115200"
```

5. Test it

```bash
# Open one terminal
cd cmd/benchmark-yamux-client
cargo run --example client

# Back to the qemu console
/usr/local/bin/benchmark-yamux-server-inguest /dev/virtio-ports/agent.channel.0
```
