package main

import (
	"flag"
	"log"
)

func main() {
	pImage := flag.String("image", "", "the user specified os base image")
	pConfig := flag.String("config", "", "the yaml config")
	pOut := flag.String("output", "disk.img", "the output bootable disk image")
	pDebug := flag.Bool("debug", false, "enable debug message")
	layoutFile := flag.String("diskLayout", "", "disk partitions layout, if not provided use the default")

	flag.Parse()

	if *pImage == "" && *pConfig == "" {
		log.Printf("Error: should specify either -image or -config")
		flag.Usage()
	}

	if *pImage == "" {
		config, _ := getConfigFromFile(*pConfig)
		log.Printf("config %#v\n", config)
		imageId, err := BuildImageFromConfig(config)
		if err != nil {
			log.Fatalf("Fail to create image %s with built-in setup\n", err)
		}
		*pImage = imageId
	}

	log.Printf("[Info] Create boot image from docker image %s\n", *pImage)

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

	// build the imgage from the config

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
