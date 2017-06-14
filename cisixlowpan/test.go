package cisixlowpan

import (
	"io"
	"log"
	"net"

	coap "github.com/dustin/go-coap"
)

//TestReq Test the request functionality
func TestReq(serverAddress string) error {
	hostIP, err := net.ResolveUDPAddr("udp6", "[::1]:5683")
	if err != nil {
		return err
	}
	ln, err := net.ListenUDP("udp6", hostIP)
	if err != nil {
		return err
	}
	defer ln.Close()

	rcvKey := func(l *net.UDPConn, a *net.UDPAddr, m *coap.Message) *coap.Message {
		log.Printf("Received the key!: %v", m.Payload)
		ln.Close()
		return nil
	}

	mux := coap.NewServeMux()
	mux.Handle(APIClientKey, coap.FuncHandler(rcvKey))

	log.Printf("Startin UDP server on %v\n", "[::1]:5683")

	err = SendCoapRequest(coap.POST, serverAddress, APIReq, string(1))
	if err != nil {
		return err
	}

	//Wait for response
	err = coap.Serve(ln, mux)
	if err != nil {
		if err != io.EOF {
			log.Printf("Error in coap.Serve: %v\n", err)
		}
	}
	log.Printf("Test finished!\n")

	return nil
}
