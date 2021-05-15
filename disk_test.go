package main

import (
	"fmt"
	"sort"
	"testing"
)

func TestBuildImage(t *testing.T) {
	BuildImageFromConfig()
}
func TestSortMount(t *testing.T) {
	layout := DiskLayout{
		Partitions: []Partition{
			{
				ID:      1,
				Start:   2048,
				End:     4095,
				Name:    "biosboot",
				GptType: GptTypeBiosBoot,
			},
			{
				ID:         2,
				Start:      8192,
				End:        212991,
				Name:       "efi",
				GptType:    GptTypeEFI,
				Fstype:     "vfat",
				FsLabel:    "BOOT",
				MountPoint: "/boot",
			},
			{
				ID:         3,
				Start:      212992,
				End:        3751007,
				Name:       "root",
				Fstype:     "ext4",
				FsLabel:    "ROOT",
				MountPoint: "/",
			},
			{
				ID:         4,
				Start:      3751936,
				End:        4161535,
				Name:       "var",
				Fstype:     "ext4",
				FsLabel:    "VAR",
				MountPoint: "/var",
			},
			{
				ID:         5,
				Start:      212992,
				End:        3751007,
				Name:       "root",
				Fstype:     "ext4",
				FsLabel:    "ROOT",
				MountPoint: "/boot/efi",
			},
		},
	}

	p := layout.Partitions
	sort.Slice(p, func(i, j int) bool { return p[i].MountPoint < p[j].MountPoint })

	for _, pp := range p {
		if pp.MountPoint != "" {
			fmt.Println(pp.MountPoint)
		}
	}
}
