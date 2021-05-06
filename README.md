# docker2boot

`docker2boot` converts a docker image to a bootable disk so that you can create your OS
from a Dockerfile with ease.

## Build

```
sudo apt-get install golang-guestfs-dev libguestfs-tools
```

Add `/usr/share/gocode` to your `GOPATH` to use the libguestfs golang binding
installed in that directory.

`make build` to produce `docker2boot`.

## Run and boot

Build the reference docker image `binc/myos:lastest`

```
cd ./images
make
```

Convert it to bootable image `disk.img`

```
sudo ./docker2boot -image binc/myos:latest -out disk.img

```

Boot the `disk.img`
```
make boot
```

You can login the console with `root:root`.
