package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

// base dockerfile and we'll add package on top of it
var base = `FROM ubuntu:{{.UbuntuVersion}} AS ubuntu

ENV DEBIAN_FRONTEND=nointeractive

# for bootloader/grub and kernel image
ARG KERNEL_VERSION={{.Kernel}}
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

# todo: handle {{.Systemd}}
RUN systemctl enable systemd-networkd.service

# handle login
{{ if .Login }}
RUN echo '{{.Login}}' | chpasswd
{{end}}

# handle packages
{{ if .Packages }}
RUN apt-get update \
    && apt-get install --no-install-recommends -y \
	{{ join .Packages " "}}
{{end}}

# handle files
# files are created first in buildcontext/tree
{{ if .Files }}
COPY tree/ /
{{end}}
`

type ResultImageID struct {
	Aux Attr `json:"aux,omitempty"`
}

type Attr struct {
	ID string `json:"ID,omitempty"`
}

// build an image from config and return the image name
func BuildImageFromConfig(c *Config) (string, error) {
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

	// create "${tmpDir}/tree" for files in c.Files
	// and in dockerfile they will be copied over using COPY tree/ /
	generateFilesIfAny(c, path.Join(tmpDir, "tree"))

	dockerfileContent := generateDockerfileContent(c)
	log.Printf("[info] dockerfile content is %s\n", dockerfileContent)

	dockerfile, err := os.Create(path.Join(tmpDir, "Dockerfile"))
	if err != nil {
		log.Fatalf("Fail to create a Dockerfile %s\n", err)
	}

	if _, err := dockerfile.Write([]byte(dockerfileContent)); err != nil {
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
		log.Println(message)
		if strings.Contains(message, "errorDetail") {
			return "", fmt.Errorf("Faild to create image %s", message)
		}

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

	if imageId == "" {
		return "", fmt.Errorf("Faild to create image - no image id genereated, enable debug to see docker build output")
	}

	return imageId, nil
}

// generate dockerfile using config from template
func generateDockerfileContent(c *Config) string {
	var funcs = template.FuncMap{"join": strings.Join}
	w := bytes.NewBufferString("")
	log.Printf("dockerfile %#v \n", *c)
	tmpl, err := template.New("dockerfile").Funcs(funcs).Parse(base)
	if err != nil {
		log.Fatalf("Fail to parse the dockerfile template %s\n", err)
	}
	if err := tmpl.Execute(w, *c); err != nil {
		log.Fatalf("Failt to execute the template to generate dockerfile from config %s\n", err)
	}

	return w.String()
}

// generate files in dir using content from Config.Files
func generateFilesIfAny(c *Config, dir string) {
	if len(c.Files) == 0 {
		return
	}

	for _, f := range c.Files {
		// default set as 0644
		var perm uint64 = 0644
		var err error
		if f.Mode != "" {
			perm, err = strconv.ParseUint(f.Mode, 8, 0)
			if err != nil {
				log.Fatalf("Fail to parse file mode %s %s\n", f.Path, f.Mode)
			}
		}
		targetFile := path.Join(dir, f.Path)
		// dir need the x bits for owner (a.k.a need the 7) so that owner can read the conents
		// so 775 is the correct permission for directoryies
		if err := os.MkdirAll(path.Dir(targetFile), 0775); err != nil {
			log.Fatalf("Fail to create dir for files %s\n", err)
		}
		if err := os.WriteFile(targetFile, []byte(f.Content), os.FileMode(perm)); err != nil {
			log.Fatalf("Fail to create file %s %s\n", f.Path, err)
		}

		log.Printf("[Info] create file %s mode %s\n", f.Path, f.Mode)
	}
}
