package main

import (
	"fmt"
	"math/big"
)

func main() {

	node := &Node{
		IP:          "127.0.0.1",
		Port:        8080,
		ID:          big.NewInt(1234),
		Successor:   nil,
		Predecessor: nil,
		//FingerTable: make([]*big.Int, 8), // m è il numero di finger table entries
	}
	node.Start()
	fmt.Println("Nodo Chord in esecuzione all'indirizzo IP ", node.IP, " sulla porta ", node.Port)

}
