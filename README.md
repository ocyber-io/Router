# Router Projects Setup Guide
This document outlines the steps to set up and run the Router project with GPU support on Ubuntu.
## Prerequisites

Ensure your machine meets the following requirements:

- **Ubuntu**: ubuntu 23.10 is required
- **Go**: 1.22+ is installed [official website](https://go.dev/doc/install).
- **GOPATH/bin**:  is in your PATH
- **GStreamer**: Follow the instructions on the [GStreamer website](https://gstreamer.freedesktop.org/documentation/installing).

## Setup Instructions

### Clone the Project

First, clone the project repository and navigate to the setup directory:

```sh
git clone git@github.com:JoinReup/reup_cohost_router.git
cd reup_cohost_router/
```



### Run the Setup Script

```sh
mage build
sudo cp ./bin/router-server /usr/local/bin/

# For Production 
sudo cp ./config-production.yaml /usr/local/etc/router.yaml


# For Staging 
sudo cp ./config-staging.yaml /usr/local/etc/router.yaml

# create linux service
sudo cp router.service/etc/systemd/system/router.service
sudo systemctl-daemon-reload

# enable service to start it automatically
sudo systemctl enable router.service 
sudo systemctl start router.service

# to restart encoder
sudo systemctl restart router.service
```

### For Check System Logs
```sh
# run this command to check the service logs
journalctl -u router.service -f

```
Ensure the paths are correct and replace with your actual project directory.

### Ports
#### Ensure that the following ports are open on your firewall and accessible on the instance:

##### `443` - primary HTTPS and TURN/TLS
##### `80` - TLS issuance
##### `7881` - WebRTC over TCP
##### `5349` - TURN/TLS TCP
##### `3478/UDP` - TURN/UDP
##### `50000-60000/UDP` - WebRTC over UDP


### DNS
#### Ensure that you have 2 domain names pointing to the instance public ip or with loadbalancing one for TCP/TURN other for HTTPS for Router.
