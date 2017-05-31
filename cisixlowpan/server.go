package cisixlowpan

import (
	"context"
	"log"
	"net"

	coap "github.com/dustin/go-coap"
)

//Server Object defining the server
type Server struct {
	ctx     context.Context
	comlink chan Message
	address net.UDPAddr
}

//NewServer create new server
func NewServer(ctx context.Context, comlink chan Message, host net.UDPAddr) *Server {
	return &Server{
		ctx:     ctx,
		comlink: comlink,
		address: host,
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
	mux.Handle("/hello", coap.FuncHandler(handleHello))
	mux.Handle("/req", coap.FuncHandler(handleReq))

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
func handleHello(l *net.UDPConn, a *net.UDPAddr, m *coap.Message) *coap.Message {
	log.Printf("Got message in handleHello: path=%q: %#v from %v", m.Path(), m, a)
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
func handleReq(l *net.UDPConn, a *net.UDPAddr, m *coap.Message) *coap.Message {
	//TODO: startup hecomm protocol
	log.Printf("Got message in handleReq: path=%q: %#v from %v", m.Path(), m, a)

	//Creating new node
	//node := storage.Node{Addr: a}

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
