package main

import (
	"flag"
	"log"
)

func main() {
	pImage := flag.String("image", "", "the os base image")
	pOut := flag.String("output", "disk.img", "the output bootable disk image")
	pDebug := flag.Bool("debug", false, "enable debug message")
	layoutFile := flag.String("diskLayout", "", "disk partitions layout, if not provided use the default")

	flag.Parse()

	if pImage == nil {
		log.Fatalf("must specify a docker image")
		flag.Usage()
	}

	log.Printf("Create boot image from base %s\n", *pImage)

	var layout *DiskLayout
	if *layoutFile != "" {
		layout = parseDisklayout(*layoutFile)

	} else {
		layout = NewDefaultLayout()
	}

	if err := layout.validate(); err != nil {
		log.Fatalf("invalid paritions setting %s", err)
	}

	const GB = 1024 * 1024 * 1024

	// output disk
	disk := Disk{
		Name: *pOut,
		Size: 2 * GB,
	}

	outTar, err := UnpackDockerImage(*pImage)
	if err != nil {
		log.Fatalf("Fail to unpack docker image %s\n", err)
	}

	// root filesystem content
	content := &[]Content{
		{
			source:     outTar,
			sourceType: "tar",
			destDir:    "/",
		},
	}

	CreateBootableImage(disk, layout, content, *pDebug)
}
