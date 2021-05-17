build:
	GO111MODULE=off go build .
run:
	# enable for debug guestfish
	# export LIBGUESTFS_DEBUG=1 LIBGUESTFS_TRACE=1
	./docker2boot -image binc/myos:latest -debug
	#./docker2boot -config config.yaml -debug
test:
	# check partitions
	guestfish add disk.img : run : list_partitions
	# check contents on boot /dev/sda2
	mkdir -p ./mnt/boot
	mkdir -p ./mnt/root
	guestmount -a disk.img -m /dev/sda2 ./mnt/boot
	ls ./mnt/boot
	# check contents on root /dev/sda3
	guestmount -a disk.img -m /dev/sda3 ./mnt/root
	ls ./mnt/root
	# boot the disk
boot:
	# create a qcow on top of the raw
	qemu-img create -f qcow2 -b disk.img disk.qcow2 5G
	# create local datesource of cloud-init
	echo "#cloud-config\nhostname: devicu\n" > user_data
	cloud-localds config.img user_data
	qemu-system-x86_64 -nographic -serial mon:stdio --enable-kvm -m 2G \
		-netdev user,id=mynet0 \
		-device e1000,netdev=mynet0 \
		-drive file=disk.qcow2,if=virtio,format=qcow2 \
		-drive file=config.img,if=virtio
clean:
	rm docker2boot
	rm *.tar -f
	rm disk.img -f
	rm disk.qcow2 -f
	rm config.img -f
	guestunmount ./mnt/boot
	guestunmount ./mnt/root
	rm ./mnt -f
