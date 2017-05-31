package main

import (
	"fmt"

	coap "github.com/dustin/go-coap"
	"github.com/joriwind/hecomm-6lowpan/cisixlowpan"
	"github.com/joriwind/hecomm-6lowpan/storage"
)

type hecommSixlowpanAPI struct {
	store *storage.Storage
}

func (api hecommSixlowpanAPI) pushKey(deveui []byte, key []byte) error {
	//Add item, pushing key down to node
	node, ok := api.store.GetNode(string(deveui[:]))
	if !ok {
		return fmt.Errorf("Not able to locate node: %v", string(deveui[:]))
	}
	err := cisixlowpan.SendCoapRequest(coap.POST, node.Addr.String(), "/key", string(key[:]))
	return err
}
