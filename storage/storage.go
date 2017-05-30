package storage

import (
	"net"

	"log"

	"github.com/joriwind/hecomm-fog/hecomm"
)

//Storage Contains the devices of the 6lowpan network
type Storage struct {
	nodes map[string]Node //!!map entries defined by DevEUI
}

//Node Connected node information
type Node struct {
	Addr   *net.UDPAddr
	DevEUI string
	Link   hecomm.LinkContract
	OsSKey [32]byte
	//AppSKey & NwkSKey managed by border router

}

//NewStorage Creating the global place to store and retrieve nodes
func NewStorage(nodes ...Node) *Storage {
	st := Storage{}
	if len(nodes) > 0 {
		for _, node := range nodes {
			st.nodes[node.DevEUI] = node
		}
	}
	return &st
}

//GetNode Retrieve the details of this node
func (st *Storage) GetNode(deveui string) (Node, bool) {
	n, ok := st.nodes[deveui]
	return n, ok
}

//AddNode Add a node to the storage
func (st *Storage) AddNode(node Node) {
	if _, ok := st.nodes[node.DevEUI]; ok {
		log.Printf("Overwriting node: %v\n", node.DevEUI)
	}
	st.nodes[node.DevEUI] = node
}
