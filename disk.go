package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/binchenx/guestfs"
)

// create disk image with specified disk layout and source content using libguestfs
type Content struct {
	source     string
	sourceType string // support tar only atm
	destDir    string
}

type Disk struct {
	Name string
	// Size in bytes
	Size int64
}

// source point to the content for root parition
func CreateBootableImage(diskImage Disk, diskLayout *DiskLayout, contents *[]Content, debug bool) {
	log.Printf("[Info]create %s\n", diskImage.Name)

	if diskLayout.ParitionType != PartitionTypeGpt {
		log.Fatalf("partition type is not gpt: %s\n ", diskLayout.ParitionType)
	}

	g, errno := guestfs.Create()
	if errno != nil {
		panic(errno)
	}
	defer g.Close()

	// Create a raw-format sparse disk image
	f, ferr := os.Create(diskImage.Name)
	if ferr != nil {
		panic(fmt.Sprintf("could not create file: %s: %s",
			diskImage.Name, ferr))
	}
	defer f.Close()

	if ferr = f.Truncate(diskImage.Size); ferr != nil {
		panic(fmt.Sprintf("could not truncate file: %s", ferr))
	}

	// Set the trace flag so that we can see each libguestfs call.
	if debug == true {
		g.Set_trace(true)
	}

	// Attach the disk image to libguestfs.
	optargs := guestfs.OptargsAdd_drive{
		Format_is_set:   true,
		Format:          "raw",
		Readonly_is_set: true,
		Readonly:        false,
	}
	if err := g.Add_drive(diskImage.Name, &optargs); err != nil {
		panic(err)
	}

	// Run the libguestfs back-end.
	if err := g.Launch(); err != nil {
		panic(err)
	}

	// get list of of device and we should expect only one since we only attched one drive
	devices, err := g.List_devices()
	if err != nil {
		panic(err)
	}
	if len(devices) != 1 {
		panic("expected a single device from list-devices")
	}

	device := devices[0]

	partitionDiskAndCreateFs(g, device, diskLayout)
	setupRootfs(g, device, diskLayout)
	copyRootfsData(g, contents)
	createAdditionalSettings(g, diskLayout)
	installBootloader(g, devices[0], "/boot")

	if err = g.Shutdown(); err != nil {
		panic(fmt.Sprintf("write to disk failed: %s", err))
	}
}

func getPartitionDeviceByName(g *guestfs.Guestfs, partitionName string) (string, error) {
	partitions, err := g.List_partitions()
	if err != nil {
		log.Fatalf("Fail to list partitions %s\n", err)
	}

	for _, p := range partitions {
		n, err := g.Part_to_partnum(p)
		if err != nil {
			log.Fatalf("Failt to get part number for %s : %s \n", p, err)
		}

		device, err := g.Part_to_dev(p)
		if err != nil {
			log.Fatalf("Failt to get device name for %s : %s \n", p, err)
		}

		name, err := g.Part_get_name(device, n)
		if err != nil {
			log.Fatalf("Failt to get partition name for %s : %s \n", p, err)
		}

		if name == partitionName {
			log.Printf("[Info] Find %s partitions at %s\n", partitionName, p)
			return p, nil
		}
	}

	return "", fmt.Errorf("cant find partition with name t")
}

// partition disk and create filesystems
func partitionDiskAndCreateFs(g *guestfs.Guestfs, device string, layout *DiskLayout) {
	g.Part_init(device, layout.ParitionType)

	for _, p := range layout.Partitions {
		g.Part_add(device, "p", p.Start, p.End)
		g.Part_set_name(device, p.ID, p.Name)
		if p.GptType != "" {
			g.Part_set_gpt_type(device, p.ID, p.GptType)
		}

		// create fs on partition (if it has one) with label
		if p.Fstype != "" {
			//FIXME: get device from partition
			partitionDevice := device + strconv.Itoa(p.ID)
			err := g.Mkfs(p.Fstype, partitionDevice, &guestfs.OptargsMkfs{
				Label_is_set: true,
				Label:        p.FsLabel})
			if err != nil {
				log.Fatalf("Fail to create fs %s on device %s\n", p.Fstype, device)
			}
		}
	}
	// check partitions
	partitions, err := g.List_partitions()
	if err != nil {
		panic(err)
	}

	if len(partitions) != len(layout.Partitions) {
		log.Fatalf("expected partitions is %d get %d\n", len(partitions), len(layout.Partitions))
	}
}

// grub-install is not a questfs command
// it is the command installed in the quest os - hence it is a command
// TODO: check if grub-install exsits in the image first
// https://wiki.archlinux.org/title/GRUB#UEFI_systems
func installBootloader(g *guestfs.Guestfs, device string, bootDir string) {
	// TODO:
	// 1. ensure grub package is installed
	// 2. ensure /boot partition is mounted (for efi)
	log.Println("[Info] Install bootloader")
	// Install bios
	// https://wiki.archlinux.org/title/GRUB#Installation
	g.Command([]string{"grub-install", "--target=i386-pc", device})
	// Install grub EFI partition
	// https://wiki.archlinux.org/title/GRUB#UEFI_systems
	g.Command([]string{"grub-install",
		"--target=x86_64-efi",
		"--efi-directory=" + bootDir,
		"--bootloader-id=GRUB",
		"--removable", device})

	log.Println("[Info]   Update grub cfg")
	// update /etc/default/grub and do upgrade-grub to genereate the grub.config and "fix"
	// see https://wiki.archlinux.org/title/GRUB#Generated_grub.cfg
	const grubSettingData = `GRUB_TIMEOUT=5
GRUB_TERMINAL="serial console"
GRUB_GFXPAYLOAD_LINUX=text
GRUB_CMDLINE_LINUX_DEFAULT="console=tty0 console=ttyS0,115200 no_timer_check nofb nomodeset vga=normal"
GRUB_SERIAL_COMMAND="serial --speed=115200 --unit=0 --word=8 --parity=no --stop=1"
`
	err := g.Write(grubSetting, []byte(grubSettingData))
	if err != nil {
		panic(err)
	}
	g.Command([]string{"update-grub"})
	// "fix" generate "$grubCfg" using ROOT and BOOT label instead of hardcoded device
	grubOri := grubCfg + ".ori"
	g.Command([]string{"cp", grubCfg, grubOri})

	g.Command([]string{"sed", "-i", "s%root=/dev/sd[a-z][0-9]%root=LABEL=ROOT%", grubCfg})
	g.Command([]string{"sed", "-i", "s%root='hd[0-9],gpt[0-9]'%root=LABEL=ROOT%", grubCfg})
	g.Command([]string{"sed", "-i", "s%root=UUID=[A-Za-z0-9\\\\-]*%root=LABEL=ROOT%", grubCfg})
	g.Command([]string{"sed", "-i", "s%search --no-floppy --fs-uuid --set=root .*$%search --no-floppy --set=root --label BOOT%", grubCfg})

	log.Println("[Info] Install bootloader DONE")
}

func setupRootfs(g *guestfs.Guestfs, device string, diskLayout *DiskLayout) {
	log.Println("[Info] Rootfs setup start....")
	partitions := diskLayout.Partitions
	// make sure root mount first
	sort.Slice(partitions, func(i, j int) bool { return partitions[i].MountPoint < partitions[j].MountPoint })

	for _, p := range diskLayout.Partitions {
		if p.MountPoint != "" {
			partitionDevice, err := getPartitionDeviceByName(g, p.Name)
			if err != nil {
				log.Fatal(err)
			}

			if p.MountPoint != "/" {
				if err := g.Mkdir_p(p.MountPoint); err != nil {
					log.Fatalln(err.Errmsg)
				}

			}
			if err := g.Mount(partitionDevice, p.MountPoint); err != nil {
				log.Fatalln(err.Errmsg)
			}
			log.Printf("[Info]   Mount %s at %s OK\n", p.MountPoint, partitionDevice)
		}
	}
	log.Println("[Info] Rootfs setup done")

}

// 1. set up fstab - call this after copyRootfsData
// 2. "fix" the side-effect caused by docker create container
func createAdditionalSettings(g *guestfs.Guestfs, diskLayout *DiskLayout) {
	// 1. set up fstab using diskLayout
	var fstabEntries []string
	for _, p := range diskLayout.Partitions {
		// we don't mount boot partition, systemd is taking care of
		if p.MountPoint != "" && p.Fstype != "" && p.FsLabel != "BOOT" {
			entry := fmt.Sprintf("LABEL=%s %s %s %s 0 0", p.FsLabel, p.MountPoint, p.Fstype, p.FsMountOps)
			fstabEntries = append(fstabEntries, entry)
		}
	}

	fstabContent := strings.Join(fstabEntries, "\n")
	log.Printf("/etc/fstab %s\n", fstabContent)
	g.Write_append("/etc/fstab", []byte(fstabContent))

	// 2. "fix"
	g.Write("/etc/hosts", []byte("172.0.0.1 localhost\n"))
	log.Println("[Info] configure nameservers")
	g.Write("/etc/resolv.conf", []byte("nameserver 127.0.0.1\nnameserver 8.8.8.8\n"))
	g.Rm_f("/.dockerenv")
}

func copyRootfsData(g *guestfs.Guestfs, contents *[]Content) error {
	log.Println("[Info] Import rootfs data")
	cs := *contents
	// content are copied as in the same sequences they should be mounted
	// e.g / first, then /boot
	sort.Slice(cs, func(i, j int) bool { return cs[i].destDir < cs[j].destDir })
	for _, c := range cs {
		if c.sourceType == "tar" {
			tarInOpt := guestfs.OptargsTar_in{
				Xattrs_is_set: true,
				Xattrs:        false,
				Acls_is_set:   true,
				Acls:          false,
			}
			if err := g.Tar_in(c.source, c.destDir, &tarInOpt); err != nil {
				return fmt.Errorf("%s", err.Errmsg)
			}
			log.Printf("[Info]   Import %s(%s) %s\n", c.source, c.sourceType, c.destDir)
		}
	}
	log.Println("[Info] Import rootfs data DONE")

	return nil
}
