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
	"github.com/joriwind/hecomm-6lowpan/storage"
)

type key int

const (
	keyStorageID key = iota
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

	//Storage of nodes
	store := storage.NewStorage()

	//Create server context
	ctx, cancel := context.WithCancel(context.Background())
	ctxv := context.WithValue(ctx, keyStorageID, store)
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
				err := cisixlowpan.SendCoapRequest(code, subcommand[1], subcommand[2], subcommand[3])
				if err != nil {
					fmt.Printf("Error in sending frame!: %v\n", err)
				}

			case "help":

			case "":
			default:
				fmt.Printf("Did not understand command: %v\n", command[0])
			}
		}
	}
}
