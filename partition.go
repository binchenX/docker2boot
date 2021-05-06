package main

import "fmt"

// disk parititions definition, generation and validation
type DiskLayout struct {
	// support gpt Only
	ParitionType string
	Partitions   []Partition
}
type Partition struct {
	// TODO: rename to Num
	ID int
	// Start is the start sector of the partition
	Start int64
	// End is the end sector of the partition
	End        int64
	Name       string
	GptType    string
	Fstype     string
	FsLabel    string
	MountPoint string
	FsMountOps string
}

const (
	PartitionTypeGpt = "gpt"
)

// https://en.wikipedia.org/wiki/GUID_Partition_Table
const (
	GptTypeBiosBoot = "21686148-6449-6E6F-744E-656564454649"
	GptTypeEFI      = "C12A7328-F81F-11D2-BA4B-00A0C93EC93B"
)

// grub settings on ubuntu distro
const (
	grubSetting = "/etc/default/grub"
	grubCfg     = "/boot/grub/grub.cfg"
)

// following partitions MUST exsits in disk layout
const (
	PartitionNameRoot = "root"
	// boot partition
	PartitionNameEFI = "efi"
	// biosboot partition
	PartitionNameBiosboot = "biosboot"
)

func (d *DiskLayout) validate() error {
	if d.ParitionType != PartitionTypeGpt {
		return fmt.Errorf("partition type must be gpt but found %s", PartitionTypeGpt)
	}

	var has_boot bool
	var has_root bool
	for _, p := range d.Partitions {
		if p.Name == PartitionNameEFI {
			has_boot = true
		}
		if p.Name == PartitionNameRoot {
			has_root = true
		}
	}

	if has_boot != true {
		return fmt.Errorf("parition table missiong efi boot partition")
	}

	if has_root != true {
		return fmt.Errorf("parition table missiong root partition")
	}

	// TODO: check partition overlap and make sure not occupy mbr and biosboot partitions
	return nil
}

// the default layout is BIOS/GPT/EFI setup, with following partition table
// mbr
// bios boot parition: 1M, for grub to intall core.img
// efi(boot)
// root
// var
//
// Info: https://wiki.archlinux.org/title/GRUB#GUID_Partition_Table_(GPT)_specific_instructions

func NewDefaultLayout() *DiskLayout {
	return &DiskLayout{
		ParitionType: "gpt",
		Partitions: []Partition{
			{
				ID:      1,
				Start:   2048,
				End:     4095,
				Name:    PartitionNameBiosboot,
				GptType: GptTypeBiosBoot,
			},
			{
				ID:         2,
				Start:      8192,
				End:        212991,
				Name:       PartitionNameEFI,
				GptType:    GptTypeEFI,
				Fstype:     "vfat",
				FsLabel:    "BOOT",
				MountPoint: "/boot",
			},
			{
				ID:         3,
				Start:      212992,
				End:        3751007,
				Name:       PartitionNameRoot,
				Fstype:     "ext4",
				FsLabel:    "ROOT",
				MountPoint: "/",
				FsMountOps: "defaults,noatime,rw",
			},
			{
				ID:         4,
				Start:      3751936,
				End:        4161535,
				Name:       "var",
				Fstype:     "ext4",
				FsLabel:    "VAR",
				MountPoint: "/var",
				FsMountOps: "defaults,noatime,rw",
			},
		},
	}
}

func parseDisklayout(file string) *DiskLayout {
	// TODO:
	return NewDefaultLayout()
}
