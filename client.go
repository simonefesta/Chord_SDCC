package main

import (
	"math/big"
	"os"
)

func checkConnectivity(nodeA *Node, nodeB *Node) bool {
	// Effettua una chiamata RPC o un ping tra nodeA e nodeB
	// Verifica se la comunicazione ha successo

	// Restituisci true se la comunicazione è avvenuta con successo, altrimenti false
	return true // o false
}

func main() {

	node1 := &Node{
		IP:          "127.0.0.1",
		Port:        8080,
		ID:          big.NewInt(1234),
		Successor:   nil,
		Predecessor: nil,
		FingerTable: make([]*big.Int, 8), // m è il numero di finger table entries
	}
	node2 := &Node{
		IP:          "127.0.0.1",
		Port:        8081,
		ID:          big.NewInt(1235),
		Successor:   nil,
		Predecessor: nil,
		FingerTable: make([]*big.Int, 8), // m è il numero di finger table entries
	}

	parametro := os.Args[1]
	if parametro == "1" {
		node1.Start()
		//fmt.Println("Nodo Chord in esecuzione all'indirizzo IP ", node1.IP, " sulla porta ", node1.Port)

	} else if parametro == "2" {
		node2.Start()
		node2.Join(node1.IP, node1.Port)

	}

	/*connected := checkConnectivity(node1, node2)
	if connected {
		fmt.Println("I nodi sono collegati.")
	} else {
		fmt.Println("I nodi non sono collegati.")
	}*/
	for true {
	}

}
