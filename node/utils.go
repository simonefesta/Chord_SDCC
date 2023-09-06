package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"time"
)

func sha_consistent(key string) int {

	m, err := ReadFromConfig() //leggo "m" dal json
	if err != nil {
		fmt.Println("Error reading m:", err)
		return -1
	}
	N := (1 << m)
	//l'hash deve essere nell'intorno (0, risorse rete)
	h := sha1.New()         //oggetto hash
	h.Write([]byte(key))    //scrivo key in byte in h
	hashedKey := h.Sum(nil) //calcolo hash
	res := byte(N - 1)      //scrivo in byte N-1
	for i := 0; i < len(hashedKey); i++ {
		res = res ^ (hashedKey[i] % byte(N)) // applico la riduzione modulo N
		//la parte tra parentesi ritorna valore tra 0 e N-1
		//res ^ (...) fa uno XOR  per ogni bit, se uno dei due bit corrispondenti è impostato, il bit nel risultato sarà impostato; altrimenti, il bit nel risultato sarà azzerato.
	}
	return int(res)
}

// test per ip e porta
func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

type Config struct {
	M int `json:"bits"`
}

func ReadFromConfig() (int, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return 0, err
	}
	filePath := filepath.Join(currentDir, "config.json")

	file, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return 0, err
	}

	return config.M, nil
}

func PrintFingerTable(me *Node) {

	fmt.Printf("FT[%d]: ", me.Id)
	for i := 1; i <= len(me.Finger)-1; i++ {
		fmt.Printf("<%d,%d> ", i, me.Finger[i])
	}
	fmt.Printf("\n")
	time.Sleep(15 * time.Second)

}

// se il nodo non risponde, chiedo al registry di verificare se sia attivo
func ContactRegistryAliveNode(nodetocheck string, idnodetocheck int) int {
	var reply int
	register, err := rpc.DialHTTP("tcp", RegistryFromInside)
	if err != nil {
		log.Fatal("Errore nodo Finger: non riesco a contattare il registry dall'interno per vedere se un nodo è vivo ", err)
	}
	arg := new(Arg)
	arg.Value = nodetocheck
	arg.Id = idnodetocheck
	err = register.Call("Registry.IsNodeAlive", arg, &reply)
	if err != nil {
		log.Fatal("Errore nodo Finger: non riesco a chiamare Registry.isNodeAlive ", err)
		return -1

	}
	register.Close()

	return 0
}

func (t *OtherNode) AskPredecessor(arg *Arg, reply *string) error { //un nodo chiede al suo predecessore se è lui a gestire la risorsa.

	idRisorsa := arg.Id
	idPredecessor := sha_consistent(node.Predecessor)
	idSuccessor := sha_consistent(node.Successor)
	fmt.Printf("Sono il %d, contattato da %d, per vedere la risorsa %d\n", node.Id, idSuccessor, idRisorsa)
	switch arg.PredOp {
	case "add":
		if (node.Id == idPredecessor && node.Id == idSuccessor) || (idRisorsa <= node.Id && idRisorsa > idPredecessor) || (idPredecessor > node.Id && (idRisorsa > idPredecessor || idRisorsa <= node.Id)) {
			fmt.Printf("AskPred - add, sono %d e dovrei gestire io la risorsa \n", node.Id)
			if node.Objects[idRisorsa] != "" {
				*reply = fmt.Sprintf("L'oggetto con id: '%d' è già esistente!\n", idRisorsa)
			} else {
				node.Objects[idRisorsa] = arg.Value
				*reply = fmt.Sprintf("Oggetto '%s' aggiunto con id: '%d'\n", arg.Value, idRisorsa)
				fmt.Printf("Nodo: %d, %v\n", node.Id, node.Objects)
			}

		}
	case "searchOrRemove":

		if (idRisorsa <= node.Id && idRisorsa > idPredecessor) || (idPredecessor > node.Id && (idRisorsa > idPredecessor || idRisorsa <= node.Id)) {
			fmt.Printf("AskPred - search, sono %d e dovrei gestire io la risorsa \n", node.Id)

			if node.Objects[idRisorsa] == "" { //se non c'è
				*reply = "L'oggetto cercato non è presente.\n"
			} else {
				if arg.Type { //se arg.Type == true, allora la ricerca l'ho fatta per rimuovere l'oggetto dal nodo.
					*reply = fmt.Sprintf("L'oggetto con id '%d' e valore '%s' è stato rimosso.\n", idRisorsa, node.Objects[idRisorsa])
					delete(node.Objects, idRisorsa)
					fmt.Printf("Nodo: %d, Objects: %v\n", node.Id, node.Objects)

				} else {
					*reply = fmt.Sprintf("L'oggetto con id '%d' e valore '%s' è posseduto dal nodo '%d'.\n", idRisorsa, node.Objects[idRisorsa], node.Id)

				}
			}
		}

	}

	return nil

}

func (t *OtherNode) GiveNodeLookup(idNodo int, ipNodo *string) error {
	if idNodo == sha_consistent(node.Successor) {
		*ipNodo = node.Successor
	} else if idNodo == sha_consistent(node.Predecessor) {
		*ipNodo = node.Predecessor
	} else {
		client, err := rpc.DialHTTP("tcp", node.Successor)
		if err != nil {
			log.Fatal("Errore nodo GiveNodeLookup: non riesco a contattare nodo successore per richiedere IP  ", err)
		}
		err = client.Call("OtherNode.GiveNodeLookup", idNodo, &ipNodo)
		if err != nil {
			log.Fatal("Errore nodo GiveNodeLookup: non riesco a chiamare OtherNode.GiveNodeLookup ", err)
		}

		client.Close()
	}

	return nil

}

func InterrogateFinger(arg *Arg, reply *string) string {

	isFound := false

	var nodoContactId int

	/*if ((node.Id < idRisorsa) && (idRisorsa <= node.Finger[1])) {
		nodoContactId = node.Finger[1]
		isFound = true
	} else {*/
	idRisorsa := arg.Id
	for i := 0; i < len(node.Finger)-1; i++ { //ispeziono FT

		if ((idRisorsa > node.Finger[i]) && (idRisorsa <= node.Finger[i+1])) || (node.Finger[i] > node.Finger[i+1] && idRisorsa <= node.Finger[i+1]) {
			nodoContactId = node.Finger[i+1]
			isFound = true
			break
		}
	}

	//}
	if !isFound { //se non trovo nulla, contatto sempre il nodo più lontano.
		nodoContactId = node.Finger[len(node.Finger)-1]
	}

	var nodoContact string

	if nodoContactId == sha_consistent(node.Successor) { //Se l'ID del nodo da contattare è quello del mio predecessore o successore, già conosco i loro indirizzi.
		nodoContact = node.Successor

	} else if nodoContactId == sha_consistent(node.Predecessor) {
		nodoContact = node.Predecessor

	} else { //altrimenti effettuo lookup per ottenere l'indirizzo.

		client, err := rpc.DialHTTP("tcp", node.Successor)
		if err != nil {
			log.Fatal("Errore nodo AddObject: non riesco a contattare il registry dall'interno  ", err)
		}
		err = client.Call("OtherNode.GiveNodeLookup", nodoContactId, &nodoContact)
		if err != nil {
			log.Fatal("Errore nodo AddObject: non riesco a chiamare OtherNode.GiveNodeLookup ", err)
		}

		client.Close()
	}
	return nodoContact

}
