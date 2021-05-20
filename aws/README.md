# create AMI for raw disk

Follow [this][1] and [this][2] to
1) upload the raw image to s3
2) create an snapshot from the s3 using ec2 vmimport function
3) create an ami from the snapshot

[1]: https://docs.01.org/clearlinux/latest/get-started/cloud-install/import-clr-aws.html
[2]: https://docs.aws.amazon.com/vm-import/latest/userguide/vmimport-image-import.html#import-vm-image
[3]: https://docs.aws.amazon.com/vm-import/latest/userguide/vmie_prereqs.html#vmimport-role

