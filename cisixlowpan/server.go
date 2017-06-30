package cisixlowpan

import (
	"bytes"
	"context"
	"encoding/binary"
	"log"
	"net"

	coap "github.com/dustin/go-coap"
	"github.com/joriwind/hecomm-6lowpan/storage"
	"github.com/joriwind/hecomm-api/hecomm"
	"github.com/joriwind/hecomm-api/hecommAPI"
)

//Server Object defining the server
type Server struct {
	ctx     context.Context
	comlink chan Message
	address net.UDPAddr
	hecomm  *hecommAPI.Platform
	store   *storage.Storage
}

//Some exports easier to use in global program
const (
	APIReq   string = "/req"
	APIHello string = "/hello"

	APIClientKey string = "/key"
)

//NewServer create new server
func NewServer(ctx context.Context, comlink chan Message, host net.UDPAddr, store *storage.Storage, pl *hecommAPI.Platform) *Server {
	return &Server{
		ctx:     ctx,
		comlink: comlink,
		address: host,
		store:   store,
		hecomm:  pl,
	}
}

//Start Start listening on configured UDP address
func (s *Server) Start() error {
	ln, err := net.ListenUDP("udp6", &s.address)
	if err != nil {
		return err
	}
	defer ln.Close()

	mux := coap.NewServeMux()
	mux.Handle(APIHello, coap.FuncHandler(s.handleHello))
	mux.Handle(APIReq, coap.FuncHandler(s.handleReq))

	log.Printf("Startin UDP server on %v\n", &s.address)

	//Start listening for coap packets --> send error back if occurs
	ch := make(chan error)
	go func() {
		err := coap.Serve(ln, mux)
		ch <- err
	}()

	//Block until stop from main or error from coap server
	select {
	case err := <-ch:
		return err
	case <-s.ctx.Done():
		return nil

	}

}

//handleHello Handle the hello path request
func (s *Server) handleHello(l *net.UDPConn, a *net.UDPAddr, m *coap.Message) *coap.Message {
	log.Printf("Got message in handleHello: path=%q: %#v from %v", m.Path(), m.Payload, a)
	defer l.Close()
	if m.IsConfirmable() {
		res := &coap.Message{
			Type:      coap.Acknowledgement,
			Code:      coap.Content,
			MessageID: m.MessageID,
			Token:     m.Token,
			Payload:   []byte("hello to you to!"),
		}
		res.SetOption(coap.ContentFormat, coap.TextPlain)

		log.Printf("Transmitting from A %#v", res)
		return res
	}
	return nil
}

//handleReq Handle the request path
func (s *Server) handleReq(l *net.UDPConn, a *net.UDPAddr, m *coap.Message) *coap.Message {
	//TODO: startup hecomm protocol
	log.Printf("Got message in handleReq: path=%q: %#v from %v", m.Path(), m.Payload, a.String())
	defer l.Close()
	//Creating new node
	//node := storage.Node{Addr: a}
	node := s.store.FindNode(a)
	if node == nil {
		//Not known node
		log.Printf("could not find node, adding new one: %v\n", *a)
		//Add node to storage
		node = &storage.Node{Addr: a, DevEUI: []byte(a.IP.String())}
		s.store.AddNode(*node)
		//Register node to fog hecomm
		nodes := []hecomm.DBCNode{
			hecomm.DBCNode{
				DevEUI:     node.DevEUI,
				InfType:    1,
				IsProvider: false,
				PlAddress:  s.hecomm.Address,
				PlType:     hecomm.CISixlowpan,
			},
		}
		err := s.hecomm.RegisterNodes(nodes)
		if err != nil {
			log.Printf("Could not register node in hecomm fog: %v\n", err)
			return nil
		}
		log.Printf("Registered new node in storage and hecommAPI: %v\n", node.Addr)
	}

	//Decode payload
	//infType := binary.BigEndian.Uint32(m.Payload)
	buf := bytes.NewBuffer(m.Payload) // b is []byte
	infType, err := binary.ReadUvarint(buf)
	if err != nil {
		log.Printf("Could read int from payload: %v\n", err)
		return nil
	}
	if infType < 1 {
		log.Printf("handleReq failed, not able to decode payload: %v\n", infType)
		return nil
	}

	//Start hecomm protocol
	if s.hecomm == nil {
		log.Printf("No hecomm configured\n")
		return nil
	}
	log.Printf("Starting requetlink with fog: node: %v, infType: %v using pl: %v", node, infType, s.hecomm)
	err = s.hecomm.RequestLink(node.DevEUI, int(infType))
	if err != nil {
		log.Printf("Error in requesting link: %v\n", err)
		return nil
	}
	log.Printf("Key established")

	if m.IsConfirmable() {
		res := &coap.Message{
			Type:      coap.Acknowledgement,
			Code:      coap.Content,
			MessageID: m.MessageID,
			Token:     m.Token,
			Payload:   []byte("good bye!"),
		}
		res.SetOption(coap.ContentFormat, coap.TextPlain)

		log.Printf("Transmitting from B %#v", res)
		return res
	}
	return nil
}
