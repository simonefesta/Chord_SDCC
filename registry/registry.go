package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sort"
	"time"
)

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
		return nil
	}
	if Nodes[id] != "" { //verifica se l'elemento con l'indice id nell'array Nodes non è una stringa vuota.
		reply.Successor = "" //se sto registrando un id già esistente, non posso aggiungerlo
		reply.Predecessor = ""
		return errors.New("esiste già un nodo con questo ID")

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
	//adesso in keys ho tutte le chiavi 'k', tranne quella del nodo di cui sto trovando predecessore e successore.

	sort.Ints(keys)
	Nodes[id] = arg.Value //metto il nodo in Nodes
	fmt.Println(Nodes)

	prevKey := keys[0]
	//cerco il successore
	for _, k := range keys { //sono ordinate
		if id < k { //appena trovo che id è minore di una certa chiave, tale chiave è il successore
			reply.Successor = Nodes[k]
			if len(keys) == 1 {
				reply.Predecessor = Nodes[k]
			}
			if id < keys[0] { //qui gestisco il caso in cui il nuovo nodo in aggiunta è quello con id più piccolo.
				reply.Predecessor = Nodes[keys[len(keys)-1]]
			} else {
				reply.Predecessor = Nodes[prevKey]

			}

			return nil
		}
		prevKey = k
	}
	reply.Successor = Nodes[keys[0]] //se il mio nodo è più grande di tutti, allora il successore è il nodo in posizione 0
	reply.Predecessor = Nodes[keys[len(keys)-1]]

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

/*
	****** Funzione per la rimozione dei nodi controllata *********************************************************************************************.

Scelto il nodo da eliminare (qualora sia presente nel sistema), lo contatto per avviare il processo di aggiornamento dei nodi vicini.
*/
func (t *Registry) RemoveNode(arg *Arg, reply *string) error {
	idNodo := arg.Id

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

		*reply = fmt.Sprintf("Il nodo avente id: '%d' è stato eliminato.\n", idNodo)
		client.Close()

	} else {
		*reply = fmt.Sprintf("Il nodo avente id '%d' non è presente e dunque non è eliminabile.\n", idNodo)
	}

	return nil

}

/*
*************** Gestione nodo caduto **************************************************************************************************************************
Se un nodo cade, il registry trova i nodi adiacenti (considerando anche i casi limite ai bordi) e li contatta per metterli in comunicazione tra loro.
*/
func FixNeighbors(id int) error {
	var succKey int
	var predKey int
	var result string

	//questo pezzo aggiorna ed ordina la lista dei nodi nel registry. Lo vediamo graficamente nel registry.
	keys := make([]int, 0, len(Nodes)) //slice delle chiavi
	for k := range Nodes {
		keys = append(keys, k)

	}
	//adesso in keys ho tutte le chiavi 'k'
	var isFound bool = false
	var lastKey int
	sort.Ints(keys)
	for i, k := range keys {
		if k >= id {
			if i > 0 {
				predKey = keys[i-1]
			} else {
				predKey = keys[len(keys)-1] //se cade il primo nodo, allora il suo precedente era l'ultimo nodo dell'anello.
			}
			succKey = k
			isFound = true
			break
		}
		lastKey = k //se non si verifica il controllo, mantengo la chiave dell'ultimo elemento dell'anello.
	}

	if !isFound {
		if len(Nodes) <= 1 {
			predKey = lastKey
			succKey = lastKey
		}
		if id > lastKey { //se cade ultimo nodo anello, allora il nodo prima dell'ultimo nodo è il precedente del primo nodo dell'anello, e viceversa. Controllo non dovrebbe servire.
			predKey = lastKey
			succKey = keys[0]
		}
	}

	client, err := rpc.DialHTTP("tcp", Nodes[predKey]) //contatto il nodo
	if err != nil {
		log.Fatal("Errore nel registry FixNeighbors: non riesco a contattere il predecessore, ", err)
	}

	err = client.Call("OtherNode.UpdatePredecessorNode", Nodes[succKey], &result) //avvio la pratica per fargli aggiornare precedente e successivo
	if err != nil {
		log.Fatal("Errore nel registry FixNeighbors: non riesco a chiamare la funzione OtherNode.UpdatePredecessorNode: ", err)

	}

	client.Close()

	client, err = rpc.DialHTTP("tcp", Nodes[succKey]) //contatto il nodo
	if err != nil {
		log.Fatal("Errore nel registry FixNeighbors: non riesco a contattere il predecessore, ", err)
	}

	err = client.Call("OtherNode.UpdateSuccessorNode", Nodes[predKey], &result) //avvio la pratica per fargli aggiornare precedente e successivo
	if err != nil {
		log.Fatal("Errore nel registry FixNeighbors: non riesco a chiamare la funzione OtherNode.UpdateSuccessorNode: ", err)

	}

	client.Close()

	return nil

}

func (t *Registry) IsNodeAlive(arg *Arg, reply *int) error {
	node := arg.Value
	id := arg.Id
	if isNodePresent(Nodes, id) {
		client, err := net.DialTimeout("tcp", node, 5*time.Second)
		if err != nil {
			fmt.Printf("Non riesco a contattare [%d:%s], procedo con la sua rimozione.\n", id, node)
			delete(Nodes, id)
			fmt.Println(Nodes)
			FixNeighbors(id)

		} else {
			client.Close()

		}
	}
	return nil
}

func main() {
	// Creazione di un nuovo oggetto Registry
	registry := new(Registry)
	rpc.Register(registry) //l'oggetto registry viene registrato per consentire la comunicazione RPC.
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)

}
