# docker2boot

`docker2boot` creates a bootable disk from either a [Docker image](./images) or
a [config yaml file](./config.yaml)


## Features

|           | status  |
| --------- |---------|
| dns       | Y       |
| cloud-init| Y       |
| network   | Y       |
| ssh       | TODO    |

## Platforms

|           | bios| uefi |console| network|
| --------- |-----|------|-------|--------|
| qemu      | Y   | Y    |  Y    |   Y    |
| openstack | Y   | Y    |  Y    |   Y    |
| aws       | Y   | y    |  Y    |   Y    |
| gce       | todo| todo | todo  | todo    |


## Build

```
sudo apt-get install libguestfs-tools cloud-utils
```

`make build` to produce `docker2boot`.

## Run and boot

### prerequisite

To run guestfish without sudo, use:
```
sudo chmod 0644 /boot/vmlinuz*
```
see [guestfish faq][faq] for more information.

### Run

### 1. Create bootable image from a [docker image](./images)

(optional) Build the reference docker image `binc/myos:lastest`

```
cd ./images
make
```

Convert it to bootable image `disk.img`

```
./docker2boot -image binc/myos:latest -out disk.img

```
### 2. Create a bootable image from [config yaml](./config.yaml)

```
./docker2boot -config config.yaml -out disk.img

```

To Boot the created  `disk.img`:
```
make boot
```

You can login the console with `root:root` and `curl www.google.com`. VM is
ready for use.

## docker image

```
d2bimage="binc/d2b:v$(cat VERSION)"
docker run -v $(pwd):$(pwd) \
    -v /var/run/docker.sock:/var/run/docker.sock \
    ${d2bimage} \
    -config $(pwd)/config.yaml \
    -output $(pwd)/disk.img \
```

[faq]: https://libguestfs.org/guestfs-faq.1.html
