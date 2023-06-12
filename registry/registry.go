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
	MaxNodi int = 10
)

var Nodes = make(map[int]string)

// argomenti per richiesta rpc client
type Argclient struct {
	ip string
}

type Arg struct {
	id    int
	value string
}

type Registry string //fa parte della registrazione anche successor
//Io qui sto registrando un nuovo nodo ("registry") di cui voglio trovare il successore

func (t *Registry) Successor(arg *Arg, reply *string) error {
	id := arg.id
	if len(Nodes) == 0 { //primo nodo
		Nodes[id] = arg.value
		*reply = arg.value //successore di sè stesso
		fmt.Println(Nodes)
		return nil
	}
	if Nodes[id] != "" { //verifica se l'elemento con l'indice id nell'array Nodes non è una stringa vuota.
		*reply = "" //se sto registrando un id già esistente, non posso aggiungerlo
		return errors.New("Esiste già un nodo con questo ID")

	}
	if len(Nodes) >= MaxNodi { //limite massimo nodi raggiunto
		*reply = ""
		return errors.New("limite nodi raggiunto")
	}

	//negli altri casi, ovvero il nodo che voglio aggiungere si può aggiungere, devo trovare il successore.
	keys := make([]int, 0, len(Nodes)) //slice delle chiavi
	for k := range Nodes {
		keys = append(keys, k)
	}
	//adesso in keys ho tutte le chiavi 'k'

	sort.Sort(sort.IntSlice(keys)) //le ordino
	Nodes[id] = arg.value          //metto il nodo in Nodes
	fmt.Println(Nodes)

	//cerco il successore
	for _, k := range keys { //sono ordinate
		if id < k { //appena trovo che id è minore di una certa chiave, tale chiave è il successore
			*reply = Nodes[k]
			return nil
		}
	}
	*reply = Nodes[keys[0]] //se il mio nodo è più grande di tutti, allora il successore è 0

	return nil
}

/*
	TODO: ReturnChordNode A CHE SERVE?

func (t *Registry) ReturnChordNode(arg *Arg, reply *string) error {

}
*/
func (t *Registry) ReturnChordNode(arg *Arg, reply *string) error {
	if len(Nodes) == 0 {
		return errors.New("Non ci sono nodi nell'anello")
	}
	keys := make([]int, 0, len(Nodes))
	for k := range Nodes {
		keys = append(keys, k)
	}

	rand.Seed(time.Now().Unix())
	n := rand.Int() % len(keys)
	*reply = Nodes[keys[n]]

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
