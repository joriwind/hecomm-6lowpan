package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"net"

	"log"

	coap "github.com/dustin/go-coap"
	"github.com/joriwind/hecomm-6lowpan/cisixlowpan"
	"github.com/joriwind/hecomm-6lowpan/storage"
	"github.com/joriwind/hecomm-api/hecomm"
	"github.com/joriwind/hecomm-api/hecommAPI"
)

const (
	//hecommAddress string = "192.168.2.123:2001"
	sixlowpanCert string = "certs/6lowpan.pem"
	sixlowpanKey  string = "certs/6lowpan-key.unencrypted.pem"
)

func main() {

	//Flag init
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("-address")
	}

	//Resolve UDP server address
	address := flag.String("address", "[aaaa::1]:5683", "Server address of UDP listener")
	flag.Parse()
	srvAddress, err := net.ResolveUDPAddr("udp6", *address)
	if err != nil {
		log.Printf("Not valid UDP server address: %v, err = %v\n", *address, err)
		return
	}

	address = flag.String("hostHecomm", "192.168.2.123:2001", "Server address of hecomm listener")
	flag.Parse()
	hecommAddress := *address
	if err != nil {
		log.Printf("Not valid UDP server address: %v, err = %v\n", *address, err)
		return
	}

	//Storage of nodes
	store := storage.NewStorage()
	//Create server context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Setup hecomm system
	//Callback api initialisation
	var cb hecommSixlowpanAPI
	cb = hecommSixlowpanAPI{store: store}
	//Certificates
	cert, err := tls.LoadX509KeyPair(sixlowpanCert, sixlowpanKey)
	if err != nil {
		log.Fatalf("fogcore: tls error: loadkeys: %s", err)
		return
	}

	caCert, err := ioutil.ReadFile(sixlowpanCert)
	if err != nil {
		log.Fatalf("cacert error: %v\n", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		ClientCAs:          caCertPool,
		InsecureSkipVerify: true,
	}

	//Start hecomm platform
	pl, err := hecommAPI.NewPlatform(ctx, hecommAddress, config, nil, cb.pushKey)
	if err != nil {
		log.Fatalf("Not able to create hecomm platform: %v\n", err)
	}

	//register
	plHecomm := hecomm.DBCPlatform{
		Address: hecommAddress,
		CI:      hecomm.CISixlowpan,
	}
	log.Println("hecomm: registering platform")
	err = hecommAPI.RegisterPlatform(plHecomm, config)
	if err != nil {
		log.Fatalf("Could not register platform: %v\n", err)
	}
	log.Printf("Platform: %v, configured in fog\n", plHecomm)

	//Start hecomm server
	go pl.Start()

	//Starting 6LoWPAN server
	channel := make(chan cisixlowpan.Message)
	server := cisixlowpan.NewServer(ctx, channel, *srvAddress, store, pl)
	go func() {
		err := server.Start()
		if err != nil {
			log.Fatalf("Server exited with error: %v", err)
		} else {
			log.Printf("Server exited\n")
		}
	}()

	//Start io input --> commands
	//command line interface of hecomm-fog
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if scanner.Scan() {
			line := scanner.Text()
			//Split line into 2 parts, the command and OPTIONALY data
			command := strings.SplitN(line, " ", 2)
			switch command[0] {

			case "exit":
				cancel()
				return

			case "send":
				subcommand := strings.SplitN(command[1], " ", 4)
				i, err := strconv.Atoi(subcommand[0])
				if err != nil {
					fmt.Printf("Error in conversion: %v", err)
					break
				}
				code := coap.COAPCode(uint8(i))
				err = cisixlowpan.SendCoapRequest(code, subcommand[1], subcommand[2], subcommand[3])
				if err != nil {
					fmt.Printf("Error in sending frame!: %v\n", err)
				}

			case "help":

			case "test":
				//subcommand := strings.SplitN(command[1], " ", 2)
				switch command[1] {
				case "req":
					log.Printf("Testing request\n")
					err = cisixlowpan.TestReq(srvAddress.String())
					if err != nil {
						log.Printf("Error occurred in test: %v\n", err)
					}
				default:
					log.Printf("Not implemented test!\n")
				}

			case "":
			default:
				fmt.Printf("Did not understand command: %v\n", command[0])
			}
		}
	}
}

type hecommSixlowpanAPI struct {
	store *storage.Storage
}

func (api hecommSixlowpanAPI) pushKey(deveui []byte, key []byte) error {
	//Add item, pushing key down to node
	node, ok := api.store.GetNode(string(deveui[:]))
	if !ok {
		return fmt.Errorf("Not able to locate node: %v", string(deveui[:]))
	}
	err := cisixlowpan.SendCoapRequest(coap.POST, node.Addr.String()+":5683", cisixlowpan.APIClientKey, string(key[:]))
	return err
}
