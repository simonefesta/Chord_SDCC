package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sort"
	"strconv"
	"sync"
	"time"
)

var nodesMutex sync.Mutex

var Nodes = make(map[int]string) //il registry mantiene un vettore di indice intero e valore 'string'. L'indice è l'id del nodo, il valore è l'ip+porta.

type Arg struct {
	Id     int
	Value  string
	Choice int
	Type   bool //1 se devo rimuovere la chiave, altrimenti la cerco e basta.
}

type Registry string //fa parte della registrazione anche successor
//Io qui sto registrando un nuovo nodo ("registry") di cui voglio trovare il successore

type NeighborsReply struct {
	Successor   string
	Predecessor string
}

var m int //numero di bits

func (t *Registry) Finger(arg *Arg, reply *[]int) error {
	id := arg.Id

	//questo pezzo aggiorna ed ordina la lista dei nodi nel registry. Lo vediamo graficamente nel registry.
	keys := make([]int, 0, len(Nodes)) //slice delle chiavi
	idInNodes := false                 //mi chiedo se l'id per cui calcolo la FT sia nella lista dei nodi. Questo perchè, se elimino un nodo dalla lista di nodi, potrei comunque calcolare la sua FT.

	for k := range Nodes {
		keys = append(keys, k)
		if id == k {
			idInNodes = true
		}
	}
	if idInNodes { //se il nodo è nella lista, allora calcolo effettivamente la FT, sennò non ha senso.
		sort.Ints(keys)

		fingerTable := make([]int, m+1)

		//fmt.Printf("\n ANALISI DEL NODO %d", id)
		for i := 1; i <= m; i++ {
			// Calcola id + 2^(i-1) mod (2^m)
			val := (id + (1 << (i - 1))) % (1 << m)
			foundSuccessor := false

			for _, k := range keys { //sono ordinate

				if val <= k {

					fingerTable[i] = k
					foundSuccessor = true

					break

				}
			}
			if !foundSuccessor { //se non ho trovato nessun successore, allora è una risorsa del primo nodo.
				fingerTable[i] = keys[0]
			}
		}
		fingerTable[0] = arg.Id

		*reply = fingerTable
	}
	return nil

}

func (t *Registry) Neighbors(arg *Arg, reply *NeighborsReply) error {
	id := arg.Id
	var err error

	m, err = ReadFromConfig() //leggo "m" dal json
	if err != nil {
		log.Fatal("Errore nella lettura del file config.json, ", err)
	}

	if len(Nodes) == 0 { //primo nodo
		Nodes[id] = arg.Value
		reply.Successor = arg.Value   //successore di sè stesso
		reply.Predecessor = arg.Value //predecessore di me stesso
		fmt.Println(Nodes)
		return nil
	}
	if Nodes[id] != "" { //verifica se l'elemento con l'indice id nell'array Nodes non è una stringa vuota.
		reply.Successor = "" //se sto registrando un id già esistente, non posso aggiungerlo
		reply.Predecessor = ""
		return errors.New("esiste già un nodo con questo ID (o si è verificata una collisione!)")

	}
	if len(Nodes) >= (1 << m) { //limite massimo nodi raggiunto
		reply.Successor = "" //se sto registrando un id già esistente, non posso aggiungerlo
		reply.Predecessor = ""
		return errors.New("limite nodi raggiunto")
	}
	//questo pezzo aggiorna ed ordina la lista dei nodi nel registry. Lo vediamo graficamente nel registry.
	keys := make([]int, 0, len(Nodes)) //slice delle chiavi
	for k := range Nodes {
		keys = append(keys, k)
	}
	//adesso in keys ho tutte le chiavi 'k'

	sort.Ints(keys)
	//fmt.Println(keys)
	Nodes[id] = arg.Value //metto il nodo in Nodes
	fmt.Println(Nodes)

	prevKey := keys[0]
	//cerco il successore
	for _, k := range keys { //sono ordinate
		if id < k { //appena trovo che id è minore di una certa chiave, tale chiave è il successore
			reply.Successor = Nodes[k]
			if len(keys) == 1 {
				reply.Predecessor = Nodes[k]
			} else {
				reply.Predecessor = Nodes[prevKey]
			}

			return nil
		}
		prevKey = k
	}

	reply.Successor = Nodes[keys[0]] //se il mio nodo è più grande di tutti, allora il successore è il nodo in posizione 0
	reply.Predecessor = Nodes[len(keys)-1]

	return nil
}

func (t *Registry) RefreshNeighbors(arg *Arg, reply *NeighborsReply) error {
	id := arg.Id
	if len(Nodes) <= 1 { //primo nodo
		Nodes[id] = arg.Value
		reply.Successor = arg.Value   //successore di sè stesso
		reply.Predecessor = arg.Value //predecessore di me stesso
		return nil
	}

	nodesMutex.Lock()
	defer nodesMutex.Unlock()
	//questo pezzo aggiorna ed ordina la lista dei nodi nel registry. Lo vediamo graficamente nel registry.
	keys := make([]int, 0, len(Nodes)) //slice delle chiavi
	isInNodes := false
	for k := range Nodes {
		keys = append(keys, k)
		if id == k {
			isInNodes = true
		}
	}
	//adesso in keys ho tutte le chiavi 'k'
	if isInNodes {
		sort.Ints(keys)
		Nodes[id] = arg.Value //metto il nodo in Nodes
		var latestKey int
		for _, k := range keys {
			latestKey = k

		}

		prevKey := keys[0]
		//cerco il successore
		for _, k := range keys { //sono ordinate
			if id < k { //appena trovo che id è minore di una certa chiave, tale chiave è il successore
				reply.Successor = Nodes[k]
				if id == keys[0] { //se sto analizzando il primo nodo, allora il pre
					reply.Predecessor = Nodes[latestKey]
				} else {
					reply.Predecessor = Nodes[prevKey]
				}

				return nil

			} else if id != k {
				prevKey = k
			}

		} //sono il più grande

		reply.Successor = Nodes[keys[0]]
		reply.Predecessor = Nodes[prevKey]

	}
	return nil
}

var lastSelectedNode = -1 // Variabile globale per tenere traccia dell'ultimo nodo selezionato. Politica Round Robin per load balancing.

func (t *Registry) EnterRing(arg *Arg, reply *string) error {

	if len(Nodes) == 0 {
		return errors.New("non ci sono nodi nell'anello")
	}
	keys := make([]int, 0, len(Nodes))
	for k := range Nodes {
		keys = append(keys, k)
	}
	sort.Ints(keys) // Ordinamento

	lastSelectedNode = (lastSelectedNode + 1) % len(keys) // Calcola l'indice del prossimo nodo

	nodeContact := Nodes[keys[lastSelectedNode]]
	*reply = ObtainAddress(nodeContact)

	return nil
}

func (t *Registry) GiveNodeLookup(idNodo int, ipNodo *string) error {
	*ipNodo = Nodes[idNodo]
	return nil

}

/*
 Funzione per la rimozione dei nodi. Scelto il nodo da eliminare (qualora sia presente nel sistema, lo contatto per avviare il processo di aggiornamento dei nodi vicini.)
*/

func (t *Registry) RemoveNode(arg *Arg, reply *string) error {
	idNodo := arg.Id
	if len(Nodes) <= 2 {
		*reply = "Raggiunto il lower bound di 2 nodi nella rete. Impossibile cancellarne altri."
		return nil

	}
	if isNodePresent(Nodes, idNodo) {

		removedNode := Nodes[idNodo]
		var result string

		client, err := rpc.DialHTTP("tcp", removedNode) //contatto il nodo
		if err != nil {
			log.Fatal("Errore nel registry RemoveNode: non riesco a contattere il nodo da rimuovere, ", err)
		}

		err = client.Call("OtherNode.UpdateNeighborsNodeRemoved", idNodo, &result) //avvio la pratica per fargli aggiornare precedente e successivo
		if err != nil {
			log.Fatal("Errore nel registry RemoveNode: non riesco a chiamare la funzione OtherNode.UpdateNeighborsNodeRemoved: ", err)

		}

		delete(Nodes, idNodo)
		fmt.Printf("Nodi dopo la rimozione : %v\n", Nodes)

		*reply = "Il nodo avente id: '" + strconv.Itoa(idNodo) + "' è stato eliminato.\n"
		client.Close()

	} else {
		*reply = "Il nodo avente id '" + strconv.Itoa(idNodo) + "' non è presente e dunque non è eliminabile.\n"
	}

	return nil

}

func isNodeAlive() {
	time.Sleep(10 * time.Second)
	for {
		for index, node := range Nodes {
			client, err := net.DialTimeout("tcp", node, 10*time.Second)
			if err != nil {
				fmt.Printf("Non riesco a contattare [%d:%s], procedo con la sua rimozione.\n", index, node)
				delete(Nodes, index)

			} else {
				client.Close()
			}

		}
		time.Sleep(5 * time.Second)

	}
}

func main() {
	// Creazione di un nuovo oggetto Registry
	registry := new(Registry)
	go isNodeAlive()
	rpc.Register(registry) //l'oggetto registry viene registrato per consentire la comunicazione RPC.
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)

}
