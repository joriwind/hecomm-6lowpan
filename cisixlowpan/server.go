package cisixlowpan

import "context"
import "net"
import "github.com/joriwind/hecomm-fog/hecomm"
import "log"

//Server Object defining the server
type Server struct {
	ctx     context.Context
	comlink chan Message
	address net.UDPAddr
	Nodes   []Node
}

//NewServer create new server
func NewServer(ctx context.Context, comlink chan Message, host net.UDPAddr) *Server {
	return &Server{
		ctx:     ctx,
		comlink: comlink,
		address: host,
	}
}

type packet struct {
	Addr net.UDPAddr
	Data []byte
}

//Node Connected node information
type Node struct {
	Addr   net.UDPAddr
	Link   hecomm.LinkContract
	OsSKey [16]byte
	//AppSKey & NwkSKey managed by slip

}

//Start Start listening on configured UDP address
func (s *Server) Start() error {
	ln, err := net.ListenUDP("udp6", &s.address)
	if err != nil {
		return err
	}
	defer ln.Close()

	log.Printf("Startin UDP server on %v\n", &s.address)

	buf := make([]byte, 1024)

	for {
		n, addr, err := ln.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		p := packet{Addr: *addr, Data: buf[:n]}
		go handleUDPPacket(p)

		select {
		case <-s.ctx.Done():
			return nil
		default:

		}
	}
}

func handleUDPPacket(p packet) error {
	return nil
}
