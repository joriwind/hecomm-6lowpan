package cisixlowpan

import (
	"log"

	coap "github.com/dustin/go-coap"
)

//SendCoapRequest Send a Coap message
func SendCoapRequest(code coap.COAPCode, destination string, path string, payload string) error {
	req := coap.Message{
		Type:      coap.Confirmable,
		Code:      code,
		MessageID: 12345,
		Payload:   []byte(payload),
	}

	req.SetOption(coap.ETag, "weetag")
	req.SetOption(coap.MaxAge, 3)
	req.SetPathString(path)

	c, err := coap.Dial("udp6", destination)
	if err != nil {
		return err
	}

	rv, err := c.Send(req)
	if err != nil {
		return err
	}

	if rv != nil {
		log.Printf("Response payload: %s", rv.Payload)
	}
	return nil
}
