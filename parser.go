package acpinger

import (
	"fmt"
	"net"
)

type parser struct {
	bytes []byte
	i     int
}

type StdPong struct {
	Protocol         int
	Mode             int
	PlayerCount      int
	MinutesRemaining int
	CurrentMap       string
	Description      string
	MaxClients       int
	Flags            int
	Mastermode       int
	Password         bool
}

type ExtPong struct {
	Ack                  int
	Version              int
	ErrorFlag            int
	PlayerStatsRespIDs   int
	PlayerStatsRespStats int
	PlayerCount          int
	Players              []Player
}

type Player struct {
	ClientNumber int
	Ping         int
	Name         string
	Team         string
	Frags        int
	Flagscore    int
	Deaths       int
	Teamkills    int
	Accuracy     int
	Health       int
	Armour       int
	GunSelected  int
	Role         int
	State        int
	IP           string
}

// newParser creates a parser that can parse AC's binary ping/pong protocol into structs.
func newParser(b []byte) parser {
	return parser{
		bytes: b,
		i:     0,
	}
}

func (p *parser) parseStd(pong *StdPong) {
	pong.Protocol = p.getInt()
	pong.Mode = p.getInt()
	pong.PlayerCount = p.getInt()
	pong.MinutesRemaining = p.getInt()
	pong.CurrentMap = p.getString()
	pong.Description = p.getString()
	pong.MaxClients = p.getInt()
	pong.Flags = p.getInt()

	pong.Mastermode = pong.Flags >> 6
	pong.Password = (pong.Flags & 1) == 1
}

func (p *parser) parseExt(pong *ExtPong) {
	pong.Ack = p.getInt()
	pong.Version = p.getInt()
	pong.ErrorFlag = p.getInt()
	playerStatsResp := p.getInt()

	if playerStatsResp == -10 {
		pong.PlayerStatsRespIDs = playerStatsResp
		pong.PlayerCount = len(p.bytes[p.i:])
		p.i += pong.PlayerCount
	} else {
		pong.PlayerStatsRespStats = playerStatsResp
		player := p.parsePlayer()
		pong.Players = append(pong.Players, player)
	}
}

func (p *parser) parsePlayer() Player {
	player := Player{
		ClientNumber: p.getInt(),
		Ping:         p.getInt(),
		Name:         p.getString(),
		Team:         p.getString(),
		Frags:        p.getInt(),
		Flagscore:    p.getInt(),
		Deaths:       p.getInt(),
		Teamkills:    p.getInt(),
		Accuracy:     p.getInt(),
		Health:       p.getInt(),
		Armour:       p.getInt(),
		GunSelected:  p.getInt(),
		Role:         p.getInt(),
		State:        p.getInt(),
	}
	_, cidr, err := net.ParseCIDR(fmt.Sprintf("%d.%d.%d.0/24", p.getByte(), p.getByte(), p.getByte()))
	if err == nil {
		player.IP = cidr.String()
	}
	return player
}

func (p *parser) hasMore() bool {
	return p.i < len(p.bytes)
}

func (p *parser) getInt() int {
	c := int(int8(p.getByte()))
	if c == -128 {
		return int(p.getByte()) | int(int8(p.getByte()))<<8
	} else if c == -127 {
		return int(p.getByte()) | int(p.getByte())<<8 | int(p.getByte())<<16 | int(p.getByte())<<32
	}
	return c
}

func (p *parser) getByte() byte {
	if !p.hasMore() {
		fmt.Errorf("getByte: out of range, attempting to access index %d (slice has length %d)", p.i, len(p.bytes))
	}
	b := p.bytes[p.i]
	p.i++
	return b
}

func (p *parser) getString() string {
	cursor := p.i
	for p.bytes[cursor] != 0 {
		cursor++
	}
	strSlice := p.bytes[p.i:cursor]
	p.i = cursor + 1

	var newSlice []byte

	// Filter out AC's color codes
	// AC represents a color as a byte prefixed by "\f"
	//
	// See: Colouring section at https://assault.cubers.net/docs/server.html
	for i := 0; i < len(strSlice); i++ {
		// 12 == ASCII Form Feed, "\f"
		if strSlice[i] == 12 {
			// Skip the Form Feed and color byte
			i += 1
		} else {
			newSlice = append(newSlice, strSlice[i])
		}
	}
	return string(newSlice)
}
