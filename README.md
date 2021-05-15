# docker2boot

`docker2boot` creates a bootable disk from either an [docker image](./images) or
a [config yaml file](./config.yaml)

## Build

```
sudo apt-get install golang-guestfs-dev libguestfs-tools
```

Add `/usr/share/gocode` to your `GOPATH` to use the libguestfs golang binding
installed in that directory.

`make build` to produce `docker2boot`.

## Run and boot

To run guestfish without sudo, use:
```
sudo chmod 0644 /boot/vmlinuz*
```
see [guestfish faq][faq] for more information.

docker2boot support two modes to create a bootable disk.
1. create from a user specified docker image.
2. create from a [config yaml](./config.yaml)

### 1. create bootable image from a docker image

Build the reference docker image `binc/myos:lastest`

```
cd ./images
make
```

Convert it to bootable image `disk.img`

```
./docker2boot -image binc/myos:latest -out disk.img

```
### 2. create a bootable image from [config yaml](./config.yaml)

```
./docker2boot -config config.yaml -out disk.img

```

To Boot the created  `disk.img`:
```
make boot
```

You can login the console with `root:root` and `curl www.google.com`. VM is
ready for use.

[faq]: https://libguestfs.org/guestfs-faq.1.html
