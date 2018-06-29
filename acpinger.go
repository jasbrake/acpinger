package acpinger

import (
	"fmt"
	"net"
	"time"
)

var (
	DefaultReadTimeout = time.Duration(3) * time.Second

	// anything other than 0 is the standard pong flag
	stdPingMsg = []byte{0x01}

	// 0 = extended pong flag
	// 1 = get player stats flag
	// -1 = get all players (instead of specifying client number)
	extPingMsg = []byte{0x00, 0x01, 0xFF}
)

func newConnection(ip string, port int) (*net.UDPConn, error) {
	// The server ping protocol port is always the game port + 1
	address := fmt.Sprintf("%s:%d", net.ParseIP(ip).String(), port+1)
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return &net.UDPConn{}, err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return &net.UDPConn{}, err
	}
	return conn, nil
}

// PingStd sends the standard ping to the server and parses the response.
// If readTimeout is 0, the default timeout will be used.
func PingStd(ip string, port int, readTimeout time.Duration) (StdPong, error) {
	if readTimeout == 0 {
		readTimeout = DefaultReadTimeout
	}

	c, err := newConnection(ip, port)
	if err != nil {
		return StdPong{}, err
	}
	defer c.Close()

	c.Write(stdPingMsg)

	res := make([]byte, 1024)
	c.SetReadDeadline(time.Now().Add(readTimeout))
	n, err := c.Read(res)
	if err != nil {
		return StdPong{}, err
	}

	parser := newParser(res[len(stdPingMsg):n])
	pong := StdPong{}
	parser.parseStd(&pong)
	return pong, nil
}

// PingExt sends the extended ping to the server and parses the response.
// If readTimeout is 0, the default timeout will be used.
func PingExt(ip string, port int, readTimeout time.Duration) (ExtPong, error) {
	if readTimeout == 0 {
		readTimeout = DefaultReadTimeout
	}

	c, err := newConnection(ip, port)
	if err != nil {
		return ExtPong{}, err
	}
	defer c.Close()

	c.Write(extPingMsg)

	pong := ExtPong{}
	// This loop logic is necessary because occasionally a server will send a
	// player before the extended pong server information.
	for pong.PlayerStatsRespIDs != -10 || len(pong.Players) < pong.PlayerCount {
		res := make([]byte, 1024)
		c.SetReadDeadline(time.Now().Add(readTimeout))
		n, err := c.Read(res)
		if err != nil {
			return ExtPong{}, err
		}
		// Trim the len of the ping we sent as it is echoed back
		parser := newParser(res[len(extPingMsg):n])
		parser.parseExt(&pong)
	}

	return pong, nil
}
