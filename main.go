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

//ConfSixlowpanCert ...
var ConfSixlowpanCert = "certs/6lowpan.cert.pem"

//ConfSixlowpanKey ...
var ConfSixlowpanKey = "private/6lowpan.key.pem"

//ConfSixlowpanCaCert ...
var ConfSixlowpanCaCert = "certs/ca-chain.cert.pem"

func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Error in searching localIP: %v\n", err)
		return ""
	}
	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Printf("Error in searching localIP: %v\n", err)
			return ""
		}
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			//If it is not loopback, it should be ok
			if !ip.IsLoopback() {

				return ip.String()
			}

		}
	}
	log.Printf("No non loopback IP addresses found!\n")
	return ""
}

func main() {

	//Flag init
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	localIP := getLocalIP()
	if localIP == "" {
		localIP = "192.168.2.106"
	}

	//Parse the flags
	flCoapAddress := flag.String("coap-address", "[aaaa::1]:5683", "Server address of UDP listener")
	flCert := flag.String("cert", ConfSixlowpanCert, "Certificate used by 6LoWPAN server")
	flKey := flag.String("key", ConfSixlowpanKey, "Private corresponding to 6LoWPAN certificate")
	flCaCert := flag.String("cacert", ConfSixlowpanCaCert, "CA certificate")
	flHecommAddress := flag.String("hecomm-address", localIP+":2001", "Server address of hecomm listener")
	flag.Parse()

	ConfSixlowpanCaCert = *flCaCert
	ConfSixlowpanCert = *flCert
	ConfSixlowpanKey = *flKey

	//Resolve UDP server address
	coapAddr, err := net.ResolveUDPAddr("udp6", *flCoapAddress)
	if err != nil {
		log.Printf("Not valid UDP server address: %v, err = %v\n", *flCoapAddress, err)
		return
	}

	flag.Parse()
	hecommAddr := *flHecommAddress
	if err != nil {
		log.Printf("Not valid UDP server address: %v, err = %v\n", *flHecommAddress, err)
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
	cert, err := tls.LoadX509KeyPair(ConfSixlowpanCert, ConfSixlowpanKey)
	if err != nil {
		log.Fatalf("fogcore: tls error: loadkeys: %s", err)
		return
	}

	caCert, err := ioutil.ReadFile(ConfSixlowpanCaCert)
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
	pl, err := hecommAPI.NewPlatform(ctx, hecommAddr, config, nil, cb.pushKey)
	if err != nil {
		log.Fatalf("Not able to create hecomm platform: %v\n", err)
	}

	//register
	plHecomm := hecomm.DBCPlatform{
		Address: hecommAddr,
		CI:      hecomm.CISixlowpan,
	}
	log.Println("hecomm: registering platform")
	err = pl.RegisterPlatform(plHecomm)
	if err != nil {
		log.Fatalf("Could not register platform: %v\n", err)
	}
	log.Printf("Platform: %v, configured in fog\n", plHecomm)

	//Start hecomm server
	go func() {
		err := pl.Start()
		if err != nil {
			log.Fatalf("Hecomm platform exited with error: %v\n", err)
		} else {
			log.Println("Hecomm platform server stopped!")
		}
	}()

	//Starting 6LoWPAN server
	channel := make(chan cisixlowpan.Message)
	server := cisixlowpan.NewServer(ctx, channel, *coapAddr, store, pl)
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
					err = cisixlowpan.TestReq(coapAddr.String())
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
	log.Printf("Pushing key to node: %v, key: %v", deveui, key)
	node, ok := api.store.GetNode(string(deveui[:]))
	if !ok {
		return fmt.Errorf("Not able to locate node: %v", string(deveui[:]))
	}
	//Send coap post request to coap server of node
	node.Addr.Port = 5683
	err := cisixlowpan.SendCoapRequest(coap.POST, node.Addr.String(), cisixlowpan.APIClientKey, string(key[:]))
	return err
}
