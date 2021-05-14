build:
	GO111MODULE=off go build .
run:
	# enable for debug guestfish
	# export LIBGUESTFS_DEBUG=1 LIBGUESTFS_TRACE=1
	sudo time ./docker2boot -image binc/myos:latest -debug
test:
	# check partitions
	sudo guestfish add disk.img : run : list_partitions
	# check contents on boot /dev/sda2
	mkdir -p ./mnt/boot
	mkdir -p ./mnt/root
	sudo guestmount -a disk.img -m /dev/sda2 ./mnt/boot
	sudo ls ./mnt/boot
	# check contents on root /dev/sda3
	sudo guestmount -a disk.img -m /dev/sda3 ./mnt/root
	sudo ls ./mnt/root
	# boot the disk
boot:
	# create a qcow on top of the raw
	qemu-img create -f qcow2 -b disk.img disk.qcow2 5G
	qemu-system-x86_64 -nographic -serial mon:stdio --enable-kvm -m 2G \
		-netdev user,id=mynet0 \
		-device e1000,netdev=mynet0 \
		-drive file=disk.qcow2,if=virtio,format=qcow2
clean:
	rm docker2boot
	rm *.tar -f
	rm disk.img -f
	rm disk.qcow2 -f
	sudo guestunmount ./mnt/boot
	sudo guestunmount ./mnt/root
	rm ./mnt -f
