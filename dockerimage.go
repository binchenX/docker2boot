package main

import (
	"context"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// UnpackDockerimage unpack docker image to a tar file
func UnpackDockerImage(image string) (string, error) {
	// create temp tar file to unpack
	outFile := path.Join(os.TempDir(), "d2b"+strconv.Itoa(int(time.Now().Unix()))+".tar")
	outf, err := os.Create(outFile)
	if err != nil {
		log.Fatalf("Failed to create file %s %s", outFile, err)
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// create container and export
	tmpContainer, err := cli.ContainerCreate(ctx, &container.Config{Image: image}, nil, nil, "")
	if err != nil {
		log.Fatalf("fail to create container from image %s %s", image, err)
	}

	fe, err := cli.ContainerExport(ctx, tmpContainer.ID)
	if err != nil {
		return "", err
	}

	io.Copy(outf, fe)
	return outFile, nil
}
