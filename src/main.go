package main

import (
	"bufio"
	"fmt"
	raftkv "kvraft"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"raft"
	"rpcraft"
	"strconv"
	"strings"
)

func MakeAndServe(clients []*rpcraft.ClientEnd, me int) {
	persister := raft.MakePersister()
	raftkv.StartKVServer(clients, me, persister, -1)
	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", clients[me].Port))
	if err != nil {
		log.Fatalf("could not listen: %v", err)
	}
	log.Printf("serving kvraft server on port %d", clients[me].Port)

	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatalf("could not listen: %v", err)
	}
}

// usage: go run main.go [server/client] [ports] [me]
func main() {
	if len(os.Args) == 1 {
		fmt.Printf("usage: go run main.go [ports] [me]")
		return
	}

	var ports []int
	for _, port := range strings.Split(os.Args[2], ",") {
		nport, err := strconv.Atoi(port)
		if err != nil {
			log.Fatalf("bad port: %s\n", port)
		}

		ports = append(ports, nport)
	}
	ip := "localhost"
	clients := rpcraft.MakeClients(ip, ports)

	switch os.Args[1] {
	case "server":
		me, err := strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatalf("bad index: %s\n", os.Args[2])
		}

		MakeAndServe(clients, me)
	case "client":
		clerk := raftkv.MakeClerk(clients)
		scanner := bufio.NewScanner(os.Stdin)
		for {
			scanner.Scan()
			input := scanner.Text()
			tokens := strings.Split(input, " ")
			switch tokens[0] {
			case "get":
				key := tokens[1]
				fmt.Println(clerk.Get(key))
			case "put":
				key := tokens[1]
				value := tokens[2]
				clerk.Put(key, value)
			case "append":
				key := tokens[1]
				value := tokens[2]
				clerk.Append(key, value)
			case "exit":
				return
			}
		}
	}
}
