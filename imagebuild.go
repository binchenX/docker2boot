package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

// base dockerfile and we'll add package on top of it
var base = `FROM ubuntu:20.04 AS ubuntu

ENV DEBIAN_FRONTEND=nointeractive

# for bootloader/grub and kernel image
ARG KERNEL_VERSION=5.4.0-58
RUN echo "link_in_boot=no" >> /etc/kernel-img.conf \
    && apt-get update \
    && apt-get install --no-install-recommends -y \
        grub-pc \
        grub-efi-amd64-bin \
        grub-efi-amd64-signed \
        linux-image-${KERNEL_VERSION}-generic \
        linux-modules-extra-${KERNEL_VERSION}-generic \
        initramfs-tools \
        intel-microcode

RUN update-initramfs -k ${KERNEL_VERSION}-generic -c

# for systemd, and /sbin/init
RUN apt-get install --no-install-recommends -y \
        systemd \
        systemd-sysv
`

type ResultImageID struct {
	Aux Attr `json:"aux,omitempty"`
}

type Attr struct {
	ID string `json:"ID,omitempty"`
}

// build an image from config and return the image name
func BuildImageFromConfig() (string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// build a complete Dockerfile from the template with the config
	// create build context from the dockerfile with some settings, create a tar from it
	tmpDir, err := ioutil.TempDir(os.TempDir(), "d2b-imagedir")
	if err != nil {
		log.Fatal(err)
	}

	dockerfile, err := os.Create(path.Join(tmpDir, "Dockerfile"))
	if err != nil {
		log.Fatalf("Fail to create a Dockerfile %s\n", err)
	}

	if _, err := dockerfile.Write([]byte(base)); err != nil {
		log.Fatalf("Fail to write files %s\n", err)
	}

	log.Printf("[info] build image from %s\n", tmpDir)

	buildContext, err := archive.TarWithOptions(tmpDir, &archive.TarOptions{})
	if err != nil {
		log.Fatalf("Fail to write files %s\n", err)
	}

	// this will return additional information as:
	//{"aux":{"ID":"sha256:818c2f5454779e15fa173b517a6152ef73dd0b6e3a93262271101c5f4320d465"}}
	out := []types.ImageBuildOutput{
		{
			Type: "string",
			Attrs: map[string]string{
				"ID": "ID",
			},
		},
	}

	resp, err := cli.ImageBuild(ctx, buildContext, types.ImageBuildOptions{Outputs: out})
	if err != nil {
		log.Fatalf("Failed to build image %s\n", err)
	}

	var imageId string
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		message := scanner.Text()
		// TODO: turn it on
		// fmt.Println(message)
		// we didn't enable only ID so expect only one entry with "aux"
		if strings.Contains(message, "aux") {
			id := ResultImageID{}
			if err := json.Unmarshal([]byte(message), &id); err != nil {
				log.Fatalf("Failt to get the image id %s\n", err)
			}
			imageId = id.Aux.ID
			log.Printf("[Info] image id %s\n", imageId)
		}
	}

	return imageId, nil
}
