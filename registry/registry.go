package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"sort"
	"time"
)

const (
	m int = 5 //, 2^m ci fornisce il numero massimo di nodi.
)

var Nodes = make(map[int]string)

// argomenti per richiesta rpc client
type Argclient struct {
	Ip string
}

type Arg struct {
	Id    int
	Value string
}

type Registry string //fa parte della registrazione anche successor
//Io qui sto registrando un nuovo nodo ("registry") di cui voglio trovare il successore

type NeighborsReply struct {
	Successor   string
	Predecessor string
}

func (t *Registry) Neighbors(arg *Arg, reply *NeighborsReply) error {
	id := arg.Id

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
	//adesso in keys ho tutte le chiavi 'k'

	sort.Ints(keys)
	fmt.Println(keys)
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

func (t *Registry) Finger(arg *Arg, reply *[]int) error {
	id := arg.Id
	//questo pezzo aggiorna ed ordina la lista dei nodi nel registry. Lo vediamo graficamente nel registry.
	keys := make([]int, 0, len(Nodes)) //slice delle chiavi
	for k := range Nodes {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	fingerTable := make([]int, m+1)
	//fmt.Printf("\n ANALISI DEL NODO %d", id)

	for i := 1; i <= m; i++ {
		// Calcola id + 2^(i-1) mod (2^m)
		val := (id + (1 << (i - 1))) % (1 << m)
		//fmt.Printf("\n val = %d = %d + %d mod %d", val, id, (1 << (i - 1)), (1 << m))
		foundSuccessor := false

		for _, k := range keys { //sono ordinate
			//fmt.Printf("\n comparazione tra k %d e val %d \n", k, val)

			if val <= k { // il PRIMO NODO >k è il successore
				//fmt.Printf("\n VA BENE tra k %d e val %d \n", k, val)

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
	fmt.Printf("Finger Table per %d ", id)
	for i := 1; i <= m; i++ {
		fmt.Printf("<%d,%d > ", i, fingerTable[i])
	}
	fmt.Printf("\n\n")

	*reply = fingerTable
	return nil

}

func (t *Registry) RefreshNeighbors(arg *Arg, reply *NeighborsReply) error {
	id := arg.Id
	if len(Nodes) <= 1 { //primo nodo
		//fmt.Println("Sotto caso Primo nodo")
		Nodes[id] = arg.Value
		reply.Successor = arg.Value   //successore di sè stesso
		reply.Predecessor = arg.Value //predecessore di me stesso
		fmt.Println(Nodes)
		return nil
	}

	//questo pezzo aggiorna ed ordina la lista dei nodi nel registry. Lo vediamo graficamente nel registry.
	keys := make([]int, 0, len(Nodes)) //slice delle chiavi
	for k := range Nodes {
		keys = append(keys, k)
	}
	//adesso in keys ho tutte le chiavi 'k'

	sort.Ints(keys)
	Nodes[id] = arg.Value //metto il nodo in Nodes
	var latestKey int
	for _, k := range keys {
		//fmt.Println("<key,value>", k, Nodes[k])
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

	/*fmt.Println("Fuori ciclo pred", Nodes[prevKey], "key = ", prevKey)
	fmt.Println("*****FINE REFRESH NEIGHBORS************", id)
	fmt.Println("")*/

	return nil
}

func (t *Registry) ReturnChordNode(arg *Arg, reply *string) error {
	//fmt.Printf("\n nodi %d \n", len(Nodes))
	if len(Nodes) == 0 {
		return errors.New("non ci sono nodi nell'anello")
	}
	keys := make([]int, 0, len(Nodes))
	fmt.Print(keys)
	for k := range Nodes {
		keys = append(keys, k)
	}

	rand.NewSource(time.Now().Unix())

	n := rand.Int() % len(keys)
	//fmt.Printf("\n ho scelto %d \n", Nodes[keys[n]])
	*reply = Nodes[keys[n]]

	return nil
}

func (t *Registry) GiveNodeLookup(idNodo int, ipNodo *string) error {
	*ipNodo = Nodes[idNodo]
	return nil

}

func main() {
	// Creazione di un nuovo oggetto Registry
	registry := new(Registry)
	rpc.Register(registry) //l'oggetto registry viene registrato per consentire la comunicazione RPC.
	rpc.HandleHTTP()       //La funzione HandleHTTP configura il pacchetto rpc per l'uso con il protocollo HTTP. Ciò consente al server RPC di gestire le richieste e le risposte RPC utilizzando il protocollo HTTP.
	l, e := net.Listen("tcp", "localhost:1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil) //avvia un server HTTP che ascolta sul listener l e gestisce le richieste in arrivo utilizzando il gestore predefinito di http.DefaultServeMux.

}
