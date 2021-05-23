FROM ubuntu:20.04 AS build
# install guestfish (include dev) and golang
ENV GO_VERSION=1.16.4
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
		g++ \
		libc6-dev \
        # we don't need golang-guestfs-dev since we use mirrored package
        golang-guestfs-dev libguestfs-dev libguestfs-tools cloud-utils \
	&& rm -rf /var/lib/apt/lists/* \
	&& curl -sSL "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz" | tar -xz -C /usr/local/

ENV PATH=$PATH:/usr/local/go/bin
# Copy and build it
WORKDIR /go/src/github.com/binchenx/docker2boot
COPY . .
RUN go build -o /usr/bin/docker2boot .

# guestfish need access /boot/vmlinuz* so we download it
# and change the access permissio
# see guestfish faq
ARG KERNEL_VERSION=5.4.0-58
RUN apt-get update && apt-get install -y --no-install-recommends \
        linux-image-${KERNEL_VERSION}-generic \
    && chmod 0644 /boot/vmlinuz*

ENTRYPOINT ["/usr/bin/docker2boot"]

