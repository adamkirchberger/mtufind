# mtufind

MTUFind is a tool to find the maximum transmission unit (MTU) to a destination.

## What is MTU

In computer networking, the maximum transmission unit (MTU) is the size of the largest protocol data unit (PDU) that can be communicated in a single network layer transaction.

## How does it work

1. Send ICMP echo request packet with a DF flag
2. Keep sending packets and increment the data size by 500 bytes each time
3. If a reply is not received then half the increment size and try again
4. Keep trying until we reach an increment of 1 byte


## Limitations

* Currently only works for IPv4 destinations
* Needs root access as the standard ICMP libraries do not support DF flag