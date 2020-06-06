# mtufind

MTUFind is a tool to find the maximum transmission unit (MTU) to a destination.

## What is MTU

In computer networking, the maximum transmission unit (MTU) is the size of the largest protocol data unit (PDU) that can be communicated in a single network layer transaction.

## How does it work

1. Send ICMP echo request packet with a DF flag
2. Keep sending packets and increment the data size by 500 bytes each time
3. If a reply is not received then half the increment size and try again
4. Keep trying until we reach an increment of 1 byte

## Install

### Linux
```
wget https://github.com/adamkirchberger/mtufind/releases/download/v1.0.1/mtufind_v1.0.1_linux_amd64.tar.gz && \
tar xvzf mtufind_v1.0.1_linux_amd64.tar.gz && \
sudo mv mtufind /usr/local/sbin
```

### Mac
```
wget https://github.com/adamkirchberger/mtufind/releases/download/v1.0.1/mtufind_v1.0.1_darwin_amd64.tar.gz && \
tar xvzf mtufind_v1.0.1_darwin_amd64.tar.gz && \
xattr -dr com.apple.quarantine mtufind && \
mv mtufind /usr/local/sbin/
```

## Run
```
sudo /usr/local/sbin/mtufind github.com
MTU to 140.82.118.4 is 1500 bytes
```

### Mac OSX

Newer Mac operating systems may complain with
>**"cannot be opened because the developer cannot be verified".**

To fix this use the Mac install command above which resolves the Gatekeeper dialog.

## Limitations

* Currently only works for IPv4 destinations
* Needs root access as the standard ICMP libraries do not support DF flag