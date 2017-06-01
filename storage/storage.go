package storage

import (
	"net"

	"log"

	"github.com/joriwind/hecomm-api/hecomm"
	"github.com/joriwind/hecomm-api/hecommAPI"
)

//Storage Contains the devices of the 6lowpan network
type Storage struct {
	nodes map[string]Node //!!map entries defined by DevEUI
}

//Node Connected node information
type Node struct {
	Addr   *net.UDPAddr
	DevEUI []byte
	Link   hecomm.LinkContract
	OsSKey [hecommAPI.KeySize]byte
	//AppSKey & NwkSKey managed by border router

}

//NewStorage Creating the global place to store and retrieve nodes
func NewStorage(nodes ...Node) *Storage {
	st := Storage{}
	if len(nodes) > 0 {
		for _, node := range nodes {
			st.nodes[string(node.DevEUI[:])] = node
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
	if _, ok := st.nodes[string(node.DevEUI[:])]; ok {
		log.Printf("Overwriting node: %v\n", node.DevEUI)
	}
	st.nodes[string(node.DevEUI[:])] = node
}

//FindNode Try locating node in storage
func (st *Storage) FindNode(address *net.UDPAddr) *Node {
	for _, value := range st.nodes {
		if value.Addr.String() == address.String() {
			return &value
		}
	}
	return nil
}
