package main

import (
	"log"
	"net/rpc"
)

type Node struct {
	id          int
	ip          string
	predecessor string
	successor   string
	fingerTable map[int]string //mappo gli chiavi intere(id) in valori (string) //secondo me è il contrario, perchè le chiavi sono le righe, i valori gli id.
}

type Arg struct { //ciò che passo ai metodi
	id    int
	value string
}

var node *Node

func newNode(ip string) *Node {
	node = new(Node)
	node.ip = ip
	return node
}

func getSuccessorIp(ip string) string {

	var result string
	arg := new(Arg)
	arg.value = ip
	arg.id = sha_adapted(ip)

	client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}
	err = client.Call("Registry.Successor", arg, &result)
	if err != nil {
		log.Fatal("Client invocation error: ", err)
	}

	return result

}
