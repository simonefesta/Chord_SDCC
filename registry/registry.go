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
	Ip string
}

type Arg struct {
	Id    int
	Value string
}

type Registry string //fa parte della registrazione anche successor
//Io qui sto registrando un nuovo nodo ("registry") di cui voglio trovare il successore

type NeightborsReply struct {
	Successor   string
	Predecessor string
}

func (t *Registry) Neightbors(arg *Arg, reply *NeightborsReply) error {
	id := arg.Id
	fmt.Println("")
	fmt.Println("*****************")
	if len(Nodes) == 0 { //primo nodo
		fmt.Println("Sotto caso Primo nodo")
		Nodes[id] = arg.Value
		reply.Successor = arg.Value   //successore di sè stesso
		reply.Predecessor = arg.Value //predecessore di me stesso
		fmt.Println(Nodes)
		return nil
	}
	if Nodes[id] != "" { //verifica se l'elemento con l'indice id nell'array Nodes non è una stringa vuota.
		fmt.Println("Sotto caso Nodo già presente")
		reply.Successor = "" //se sto registrando un id già esistente, non posso aggiungerlo
		reply.Predecessor = ""
		return errors.New("esiste già un nodo con questo ID")

	}
	if len(Nodes) >= MaxNodi { //limite massimo nodi raggiunto
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
	for _, k := range keys {
		fmt.Println(k, Nodes[k])
	}

	prevKey := keys[0]
	//cerco il successore
	for _, k := range keys { //sono ordinate
		fmt.Println("fuori dal ciclo", Nodes[k], "valore k", k)
		if id < k { //appena trovo che id è minore di una certa chiave, tale chiave è il successore
			fmt.Println("Ho trovato un successore", Nodes[k])
			reply.Successor = Nodes[k]
			if len(keys) == 1 {
				fmt.Println("Ho trovato un predecessore (caso size 1)", Nodes[k])
				reply.Predecessor = Nodes[k]
			} else {
				fmt.Println("Ho trovato un predecessore (caso size >1)", Nodes[prevKey])
				reply.Predecessor = Nodes[prevKey]
			}

			return nil
		}
		prevKey = k
	}
	fmt.Println("Fuori ciclo successore", Nodes[keys[0]])
	fmt.Println("Fuori ciclo predecessore", Nodes[len(keys)-1])
	reply.Successor = Nodes[keys[0]] //se il mio nodo è più grande di tutti, allora il successore è il nodo in posizione 0
	reply.Predecessor = Nodes[len(keys)-1]
	fmt.Println("")

	return nil
}

func (t *Registry) RefreshNeightbors(arg *Arg, reply *NeightborsReply) error {
	id := arg.Id
	fmt.Println("")
	fmt.Println("*****REFRESH NEIGHBORS ************", id)
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
		fmt.Println("<key,value>", k, Nodes[k])
		latestKey = k
		fmt.Println("latest key", latestKey)

	}
	fmt.Println("LATEST <key,value>", latestKey, Nodes[latestKey])

	prevKey := keys[0]
	//cerco il successore
	for _, k := range keys { //sono ordinate
		if id < k { //appena trovo che id è minore di una certa chiave, tale chiave è il successore
			fmt.Println("CONFRONTO ", id, k, "Ho trovato un successore", Nodes[k])
			reply.Successor = Nodes[k]
			if id == keys[0] { //se sto analizzando il primo nodo, allora il pre
				fmt.Println("check primo nodo", Nodes[latestKey])
				reply.Predecessor = Nodes[latestKey]
			} else {
				fmt.Println("node prevkey", Nodes[prevKey])
				reply.Predecessor = Nodes[prevKey]
			}

			return nil

		} else if id != k {
			prevKey = k
		}

		fmt.Println("latest key", prevKey)

	} //sono il più grande
	fmt.Println("Fuori ciclo successore", Nodes[keys[0]])

	reply.Successor = Nodes[keys[0]]
	reply.Predecessor = Nodes[prevKey]
	//se il mio nodo è più grande di tutti, allora il successore è il nodo in posizione 0
	/*if Nodes[prevKey] != Nodes[id] {
		fmt.Println("Fuori ciclo predecessore", Nodes[prevKey])
		reply.Predecessor = Nodes[prevKey]
	} else {
		fmt.Println("EXTERN CASO prevkey == nodes k")
		reply.Predecessor = Nodes[keys[0]]
	}*/

	fmt.Println("Fuori ciclo pred", Nodes[prevKey], "key = ", prevKey)

	fmt.Println("*****FINE REFRESH NEIGHBORS************", id)
	fmt.Println("")

	return nil
}

func (t *Registry) ReturnChordNode(arg *Arg, reply *string) error {
	if len(Nodes) == 0 {
		return errors.New("non ci sono nodi nell'anello")
	}
	keys := make([]int, 0, len(Nodes))
	for k := range Nodes {
		fmt.Println("->", k)
		keys = append(keys, k)
	}

	rand.NewSource(time.Now().Unix())

	n := rand.Int() % len(keys)
	fmt.Println("**********")
	fmt.Println("SIZE is", len(keys))
	fmt.Println("n from returnchord is ", n)
	fmt.Println("while node is ", Nodes[keys[n]])

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
