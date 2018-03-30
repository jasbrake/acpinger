package acpinger

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
)

type Server struct {
	IP   string
	Port int
}

var (
	onlineServer = Server{
		IP:   "127.0.0.1",
		Port: 10000,
	}
	offlineServer = Server{
		IP:   "127.0.0.1",
		Port: 11111,
	}
)

func TestMain(m *testing.M) {
	done := make(chan bool)
	go startMockPongServer(done)
	c := m.Run()
	close(done)
	os.Exit(c)
}

func startMockPongServer(done chan bool) {
	address, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", onlineServer.Port+1))
	log.Printf("Starting mock server on port %d\n", address.Port)
	if err != nil {
		panic(err)
	}
	s, err := net.ListenUDP("udp", address)
	if err != nil {
		panic(err)
	}
	defer s.Close()

	// Close the mock server when done
	go func() {
		<-done
		s.Close()
		return
	}()

	buf := make([]byte, 1024)

	for {
		n, conn, err := s.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		if bytes.Equal(buf[:n], stdPingMsg) {
			s.WriteToUDP([]byte{1, 128, 177, 4, 5, 1, 10, 97, 99, 95, 100, 101, 115, 101, 114, 116, 51, 0, 84, 101, 115, 116, 32, 83, 101, 114, 118, 101, 114, 0, 16, 0}, conn)
		} else if bytes.Equal(buf[:n], extPingMsg) {
			s.WriteToUDP([]byte{0, 1, 255, 255, 104, 0, 246, 0}, conn)
			s.WriteToUDP([]byte{0, 1, 255, 255, 104, 0, 245, 0, 128, 186, 0, 80, 108, 97, 121, 101, 114, 49, 0, 67, 76, 65, 0, 0, 0, 1, 0, 0, 100, 0, 8, 0, 0, 1, 2, 3}, conn)
		}
	}
}

func TestStdPingSucceedsWhenReplyValid(t *testing.T) {
	pong, err := PingStd(onlineServer.IP, onlineServer.Port, 0)
	if err != nil {
		t.Error(err)
	}

	if pong.Protocol != 1201 {
		t.Errorf("Expected Protocol to be 1201, found %d\n", pong.Protocol)
	}
	if pong.Mode != 5 {
		t.Errorf("Expected Mode to be 5, found %d\n", pong.Mode)
	}
	if pong.PlayerCount != 1 {
		t.Errorf("Expected PlayerCount to be 1, found %d\n", pong.PlayerCount)
	}
	if pong.MinutesRemaining != 10 {
		t.Errorf("Expected MinutesRemaining to be 10, found %d\n", pong.MinutesRemaining)
	}
	if pong.CurrentMap != "ac_desert3" {
		t.Errorf("Expected CurrentMap to be 'ac_desert3', found '%s'\n", pong.CurrentMap)
	}
	if pong.Description != "Test Server" {
		t.Errorf("Expected Description to be 'Test Server', found '%s'\n", pong.Description)
	}
	if pong.MaxClients != 16 {
		t.Errorf("Expected MaxClients to be 16, found %d\n", pong.MaxClients)
	}
	if pong.Flags != 0 {
		t.Errorf("Expected Flags to be 0, found %d\n", pong.Flags)
	}
	if pong.Mastermode != 0 {
		t.Errorf("Expected MasterMode to be 0, found %d\n", pong.Mastermode)
	}
	if pong.Password {
		t.Errorf("Expected Password to be false, found '%s'\n", pong.Password)
	}
}

func TestStdPingFailsWhenNoReply(t *testing.T) {
	_, err := PingStd(offlineServer.IP, offlineServer.Port, 0)
	if err == nil {
		t.Errorf("Expected no reply from %s:%d\n", offlineServer.IP, offlineServer.Port)
	}
}

func TestExtPingSucceedsWhenReplyValid(t *testing.T) {
	pong, err := PingExt(onlineServer.IP, onlineServer.Port, 0)
	if err != nil {
		t.Error(err)
	}

	if pong.Ack != 255 {
		t.Errorf("Expected Ack to be 255, found %d", pong.Ack)
	}
	if pong.Version != 104 {
		t.Errorf("Expected Version to be 104, found %d", pong.Version)
	}
	if pong.ErrorFlag != 0 {
		t.Errorf("Expected ErrorFlag to be 0, found %d", pong.ErrorFlag)
	}
	if pong.PlayerStatsRespIDs != 246 {
		t.Errorf("Expected PlayerStatsRespIDs to be 246, found %d", pong.PlayerStatsRespIDs)
	}
	if pong.PlayerStatsRespStats != 245 {
		t.Errorf("Expected PlayerStatsRespStats to be 245, found %d", pong.PlayerStatsRespStats)
	}
	if pong.PlayerCount != 1 || len(pong.Players) != 1 {
		t.Errorf("Expected PlayerCount to be 1, found PlayerCount: %d, length of Players array: %d", pong.PlayerCount, len(pong.Players))
	}

	p := pong.Players[0]
	if p.ClientNumber != 0 {
		t.Errorf("Expected ClientNumber to be 0, found %d", p.ClientNumber)
	}
	if p.Ping != 186 {
		t.Errorf("Expected Ping to be 186, found %d", p.Ping)
	}
	if p.Name != "Player1" {
		t.Errorf("Expected Name to be 'Player1', found '%s'", p.Name)
	}
	if p.Team != "CLA" {
		t.Errorf("Expected Team to be 'CLA', found '%s'", p.Team)
	}
	if p.Frags != 0 {
		t.Errorf("Expected Frags to be 0, found %d", p.Frags)
	}
	if p.Flagscore != 0 {
		t.Errorf("Expected Flagscore to be 0, found %d", p.Flagscore)
	}
	if p.Deaths != 1 {
		t.Errorf("Expected Deaths to be 1, found %d", p.Deaths)
	}
	if p.Teamkills != 0 {
		t.Errorf("Expected Teamkills to be 0, found %d", p.Teamkills)
	}
	if p.Accuracy != 0 {
		t.Errorf("Expected Accuracy to be 0, found %d", p.Accuracy)
	}
	if p.Health != 100 {
		t.Errorf("Expected Health to be 100, found %d", p.Health)
	}
	if p.Armour != 0 {
		t.Errorf("Expected Armour to be 0, found %d", p.Armour)
	}
	if p.GunSelected != 8 {
		t.Errorf("Expected GunSelected to be 8, found %d", p.GunSelected)
	}
	if p.Role != 0 {
		t.Errorf("Expected Role to be 0, found %d", p.Role)
	}
	if p.State != 0 {
		t.Errorf("Expected State to be 0, found %d", p.State)
	}
	if p.IP != "1.2.3.0/24" {
		t.Errorf("Expected IP to be 1.2.3.0/24, found %s", p.IP)
	}
}

func TestExtPingFailsWhenNoReply(t *testing.T) {
	_, err := PingExt(offlineServer.IP, offlineServer.Port, 0)
	if err == nil {
		t.Errorf("Expected no reply from %s:%d\n", offlineServer.IP, offlineServer.Port)
	}
}
