FROM ubuntu:20.04 AS ubuntu

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

# user spaces service & packages
#
# Enable essential services
RUN systemctl enable systemd-networkd.service

# for debug
RUN echo 'root:root' | chpasswd

# Install essential and utility packages
RUN apt-get update \
    && apt-get install --no-install-recommends -y \
        # Essential packages
        # we use netplan to configure network
        netplan.io \
        # utility packages
        curl \
        dnsutils \
        iproute2 \
        cloud-init \
        # required by cloud-init
        lsb-release

# configs for essential services
COPY tree/ /
