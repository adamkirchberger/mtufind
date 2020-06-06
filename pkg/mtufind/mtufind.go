package mtufind

import (
	"errors"
	"math/rand"
	"time"

	"net"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// MTUFind is a tool to find the maximum transmission unit (MTU) to a destination
type MTUFind struct {
	Destination net.IP
	startSize   int
	ID          int
	conn        net.PacketConn
}

// New is a constructor to create MTUFind object
func New(destination string) (*MTUFind, error) {
	conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}

	// Seed for ID
	rand.Seed(time.Now().Unix())
	ID := rand.Intn(65535)
	log.Debug().Msgf("ICMP ID will be: %d", ID)

	resolve46, err := net.LookupIP(destination)
	if err != nil {
		return nil, errors.New("Invalid IP/FQDN")
	}

	// Remove any IPv6 addresses
	var resolve4 []net.IP
	for _, addr := range resolve46 {
		if addr.To4() != nil {
			resolve4 = append(resolve4, addr)
		}
	}

	ip := resolve4[rand.Intn(len(resolve4))]
	log.Debug().Msgf("DNS lookup for `%s` returned `%v`", destination, resolve4)
	log.Debug().Msgf("Destination set to `%s`", ip)

	if ip.String() == "127.0.0.1" {
		return nil, errors.New("127.0.0.1 is not a valid destination")
	}

	return &MTUFind{
		Destination: ip,
		startSize:   50,
		ID:          ID,
		conn:        conn,
	}, nil

}

// Run is used to run an MTUFind test
func (mf *MTUFind) Run() (int, error) {
	maxSize := 0
	inc := 500
	count := 1

	// Ensure destination is reachable
	log.Debug().Msgf("Check reachability send 64 bytes")
	if err := mf.send(56); err == nil {
		if err := mf.receive(56); err != nil {
			return -1, errors.New("Unreachable")
		}
	}

	for size := mf.startSize; size <= 20000; {
		size += inc
		log.Debug().Msgf("count=%d size=%d inc=%d max=%d", count, size, inc, maxSize)
		count++

		// Throttle requests
		time.Sleep(50 * time.Millisecond)

		// Send ICMP echo request
		if err := mf.send(size); err != nil {
			// Error sending packet
			log.Debug().Msgf("Send error: %s", err)

			// Error at smallest increment means the end
			if inc == 1 {
				break
			}

			// Resize and try again
			mf.resizer(&size, &inc)
			continue
		}

		log.Debug().Msgf("Sent %d bytes to %s", size, mf.Destination)

		// Look for ICMP echo reply
		err := mf.receive(size)
		if err != nil {
			log.Debug().Msgf("Reply error: %s", err)

			// Error at smallest increment means the end
			if inc == 1 {
				break
			}

			// Resize and try again
			mf.resizer(&size, &inc)
			continue
		}
		log.Debug().Msgf("Received reply from %s", mf.Destination)

		// New MAX
		if size > maxSize {
			maxSize = size
		}
	}

	log.Debug().Msgf("Result is 20(IPv4) + 8(ICMP) + %d(Max Data)", maxSize)
	return 20 + 8 + maxSize, nil
}

// resizer is used to adjust the size of packet everytime it is too large
func (mf *MTUFind) resizer(size, inc *int) {
	if *inc < 10 {
		*size -= *inc - 1
		*inc = 1
	} else {
		*size -= *inc - 1
		*inc /= 2
	}
}

// send an ICMP packet with specified data size
func (mf *MTUFind) send(size int) error {
	p, err := ipv4.NewRawConn(mf.conn)
	if err != nil {
		return err
	}

	// Get ICMP packet as bytes
	ib, err := mf.icmp(size).Marshal(nil)
	if err != nil {
		return err
	}

	// Send IPv4 header + ICMP bytes to destination
	if err := p.WriteTo(mf.header(size), ib, nil); err != nil {
		return errors.New("Packet too large for local MTU")
	}

	// Successful
	return nil
}

// receive will check the next 10 received ICMP packets for matching response
func (mf *MTUFind) receive(size int) error {
	// Timeout if no ICMP packet is received within this value then give up
	timeout := 1 * time.Second
	// How many ICMP packets to check for our reply until we give up
	packetsCheck := 10

	for current := 1; current <= packetsCheck; current++ {
		// Create byte object to hold packet
		reply := make([]byte, 10000)
		// Wait for an ICMP packet to arrive
		if err := mf.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return err
		}

		// Read in the packet
		len, peer, err := mf.conn.ReadFrom(reply)
		if err != nil {
			return errors.New("timeout")
		}

		// Parse the ICMP segment
		parse, err := icmp.ParseMessage(ipv4.ICMPTypeEcho.Protocol(), reply[:len])
		if err != nil {
			return err
		}

		// Type switch to filter received echo replies only
		switch body := parse.Body.(type) {
		case *icmp.Echo:
			// Ensure that the received packet is from our destination
			if peer.String() != mf.Destination.String() {
				// Skip this packet and try again
				continue
			}

			// Ensure ID matches
			if body.ID != mf.ID {
				// Skip this packet and try again
				continue
			}

			// All checks passed so exit successfully
			return nil
		}

	}
	return errors.New("no reply")
}

// header is used to generate necessary IPv4 headers
func (mf *MTUFind) header(size int) *ipv4.Header {
	h := &ipv4.Header{
		Version:  ipv4.Version,
		Len:      ipv4.HeaderLen,
		TotalLen: ipv4.HeaderLen + 8 + size,    // IPv4 + ICMP + Payload
		Protocol: ipv4.ICMPTypeEcho.Protocol(), // ICMP
		TTL:      64,
		Dst:      mf.Destination,
		Flags:    2, // Set DF
	}
	return h
}

// icmp is used to build our ICMP packet
func (mf *MTUFind) icmp(size int) *icmp.Message {
	i := &icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: mf.ID, Seq: 0,
			Data: make([]byte, size), // 8 bytes ICMP header + payload
		},
	}
	return i
}
