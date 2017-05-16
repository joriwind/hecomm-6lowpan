package cisixlowpan

import (
	"encoding/json"
	"net"
)

/*
 * File describing the messages communicated between 6lowpan server and node
 */

//Message General Message type coming from 6LoWPAN network
type Message struct {
	DevEUI net.HardwareAddr
	DevIP  net.Addr
	Cmd    CmdMessage
}

//CmdMessage Part of message containing command from or to node
type CmdMessage struct {
	CPort int
	Data  []byte
}

//NewMessage Create new message and figure out command
func NewMessage(devIP net.Addr, devEUI net.HardwareAddr, data []byte) (*Message, error) {
	var command CmdMessage
	var message Message

	if err := json.Unmarshal(data, command); err != nil {
		return &message, nil
	}
	message.Cmd = command
	message.DevEUI = devEUI
	message.DevIP = devIP

	return &message, nil
}
