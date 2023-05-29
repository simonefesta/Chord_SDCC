package main

import (
	"crypto/sha1"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
)

type Node struct {
	ID          *big.Int   //identificatore nodo
	IP          string     //indirizzo ip nodo
	Port        int        //numero di porta
	Predecessor *Node      //nodo precedente
	Successor   *Node      //nodo successivo
	FingerTable []*big.Int //la finger table è un array di puntatori ad altri nodi
}

const (
	m    = 8
	port = 8080
)

// calcolo l'hash di "key" tramite sha1
func hash(key string) *big.Int {
	h := sha1.New()
	h.Write([]byte(key))
	hash := fmt.Sprintf("%x", h.Sum(nil))
	intHash := new(big.Int)
	intHash.SetString(hash, 16)
	return intHash
}

// avvio di un nodo
func (n *Node) Start() {
	rpc.Register(n)  //registro il nodo corrente come servente rpc, gli altri nodi possono chiamarlo.
	rpc.HandleHTTP() //gestisco richieste http in arrivo

	l, err := net.Listen("tcp", n.IP+":"+strconv.Itoa(n.Port)) //apre connessione in ascolto su n.Port dell' n.Ip specificato. Cosi altri nodi vi si possono connettere.
	if err != nil {
		fmt.Println("Errore durante l'ascolto sulla porta ", n.Port, ": ", err)
		os.Exit(1)
	}

	fmt.Println("Nodo Chord in esecuzione all'indirizzo IP ", n.IP, " sulla porta ", n.Port)

	go http.Serve(l, nil) //avvio server http in bg, per accettare richieste http in arrivo tramite i metodi di prima.

}

func (n *Node) Join(ip string, port int) {
	client, err := rpc.DialHTTP("tcp", ip+":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("Errore durante la connessione al nodo ", ip, ":", port, ": ", err)
		return
	}

	var successor *Node
	err = client.Call("Node.FindSuccessor", n.ID, &successor)
	if err != nil {
		fmt.Println("Errore durante la chiamata RPC a FindSuccessor sul nodo ", ip, ":", port, ": ", err)
		return
	}

	n.Successor = successor

}

func (n *Node) FindSuccessor(key *big.Int, successor *Node) error {
	if betweenRightInclusive(key, n.ID, n.Successor.ID) {
		*successor = *n.Successor
		return nil
	}

	// Chiamata RPC a closestPrecedingNode sul nodo corrente
	closestNode := n.closestPrecedingNode(key)
	if closestNode == nil {
		return fmt.Errorf("No closest preceding node found for key: %s", key.String())
	}

	// Chiamata RPC a FindSuccessor sul nodo closestNode
	client, err := rpc.DialHTTP("tcp", closestNode.IP+":"+strconv.Itoa(closestNode.Port))
	if err != nil {
		return err
	}

	err = client.Call("Node.FindSuccessor", key, successor)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) getSuccessor(nodeID *big.Int, key *big.Int, successor *big.Int) error {
	client, err := rpc.DialHTTP("tcp", n.IP+":"+strconv.Itoa(n.Port))
	if err != nil {
		return err
	}

	// Chiamata RPC al metodo FindSuccessor sul nodo successore attuale
	err = client.Call("Node.FindSuccessor", key, successor)
	if err != nil {
		return err
	}

	return nil
}

func betweenRightInclusive(key, start, end *big.Int) bool {
	if start.Cmp(end) == 0 {
		return true
	}
	if start.Cmp(end) < 0 {
		return key.Cmp(start) > 0 && key.Cmp(end) <= 0
	}
	return key.Cmp(start) > 0 || key.Cmp(end) <= 0
}

func (n *Node) closestPrecedingNode(key *big.Int) *Node {
	for i := m - 1; i >= 0; i-- {
		if n.FingerTable[i] != nil && betweenRightInclusive(n.FingerTable[i], n.ID, key) {
			return n
		}
	}
	return n.Successor // In questa implementazione, assumiamo che n.Successor sia di tipo *Node
}

/*
func remove(key string) error {

	// TODO: remove key
}

func get(key string) error {
	//TODO: retrieve key

}

func put(key string, value int) error {
	//TODO: put value V associated with key K

}

//Introduzione di un nuovo nodo

func registerP2P() {
	//TODO: devo registrarmi presso qualche parte per appartenere alla rete.
}

func init_node() {

	//TODO: assegnazione di valori etc
} */
