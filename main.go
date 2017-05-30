package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"net"

	"log"

	coap "github.com/dustin/go-coap"
	"github.com/joriwind/hecomm-6lowpan/cisixlowpan"
)

func main() {

	//Flag init
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("-address")
	}

	//Resolve UDP server address
	address := flag.String("address", "[::1]:5683", "Server address of UDP listener")
	flag.Parse()
	if "" == *address {
		*address = "[::1]:5683"
	}
	srvAddress, err := net.ResolveUDPAddr("udp6", *address)
	if err != nil {
		log.Printf("Not valid UDP server address: %v, err = %v\n", *address, err)
		return
	}

	//Create server context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Starting 6LoWPAN server
	channel := make(chan cisixlowpan.Message)
	server := cisixlowpan.NewServer(ctx, channel, *srvAddress)
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
				sendCoapRequest(code, subcommand[1], subcommand[2], subcommand[3])

			case "help":

			case "":
			default:
				fmt.Printf("Did not understand command: %v\n", command[0])
			}
		}
	}
}

func sendCoapRequest(code coap.COAPCode, destination string, path string, payload string) {
	req := coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.GET,
		MessageID: 12345,
		Payload:   []byte(payload),
	}

	req.SetOption(coap.ETag, "weetag")
	req.SetOption(coap.MaxAge, 3)
	req.SetPathString(path)

	c, err := coap.Dial("udp", destination)
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}

	rv, err := c.Send(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	if rv != nil {
		log.Printf("Response payload: %s", rv.Payload)
	}
}
