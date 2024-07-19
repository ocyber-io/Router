# Router Projects Setup Guide
This document outlines the steps to set up and run the Router project with GPU support on Ubuntu.
## Prerequisites

Ensure your machine meets the following requirements:

- **Go**: 1.22+ is installed [official website](https://go.dev/doc/install).
- **GOPATH/bin**:  is in your PATH
- **GStreamer**: Follow the instructions on the [GStreamer website](https://gstreamer.freedesktop.org/documentation/installing).

## Setup Instructions

### Clone the Project

First, clone the project repository and navigate to the setup directory:

```sh
git clone https://github.com/ocyber-io/router.git
cd router/
```



### Run the Setup Script

```sh
./setup.sh
```

Ensure the paths are correct and replace with your actual project directory.

### Ports
#### Ensure that the following ports are open on your firewall and accessible on the instance:

##### `443` - primary HTTPS and TURN/TLS
##### `80` - TLS issuance
##### `7881` - WebRTC over TCP
##### `3478/UDP` - TURN/UDP
##### `50000-60000/UDP` - WebRTC over UDP


### DNS
#### Ensure that you have 2 domain names pointing to the instance public ip or with loadbalancing one for TCP/TURN other for HTTPS for Router.
