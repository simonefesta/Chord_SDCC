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
	"strconv"
	"time"
)

var Nodes = make(map[int]string)

// argomenti per richiesta rpc client
type Argclient struct {
	Ip string
}

type Arg struct {
	Id     int
	Value  string
	Choice int
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
		fmt.Println("Error reading m:", err)
		return err
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

func (t *Registry) Finger(arg *Arg, reply *[]int) error {
	id := arg.Id
	//fmt.Printf("in finger ho m = %d\n", m)
	/*m, err := ReadFromConfig() //leggo "m" dal json
	if err != nil {
		fmt.Println("Error reading m:", err)
		return err
	}*/
	//fmt.Printf("in finger ho DOPO m = %d\n", m)

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
		//fmt.Printf("FT per id: %d \n", id)
		//fmt.Println(Nodes)
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

		*reply = fingerTable
	}
	return nil

}

func (t *Registry) RefreshNeighbors(arg *Arg, reply *NeighborsReply) error {
	id := arg.Id
	if len(Nodes) <= 1 { //primo nodo
		//fmt.Println("Sotto caso Primo nodo")
		Nodes[id] = arg.Value
		reply.Successor = arg.Value   //successore di sè stesso
		reply.Predecessor = arg.Value //predecessore di me stesso
		//fmt.Println(Nodes)
		return nil
	}

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
	}
	return nil
}

func (t *Registry) EnterRing(arg *Arg, reply *string) error {
	choice := arg.Choice

	if len(Nodes) == 0 {
		return errors.New("non ci sono nodi nell'anello")
	}
	keys := make([]int, 0, len(Nodes))
	//fmt.Print(keys)
	for k := range Nodes {
		keys = append(keys, k)
	}

	rand.NewSource(time.Now().Unix())
	//*reply
	n := rand.Int() % len(keys)
	var result string
	nodeContact := Nodes[keys[n]]

	switch choice {
	case 1:
		client, err := rpc.DialHTTP("tcp", nodeContact) //contatto il nodo che ho trovato prima.
		if err != nil {
			var opErr *net.OpError
			if errors.As(err, &opErr) {
				// Errore specifico dell'operazione di rete
				log.Fatalf("Errore connessione client nodo da contattare: %s, Op: %s, Net: %s", err, opErr.Op, opErr.Net)
			} else {
				// Altro tipo di errore
				log.Fatal("Errore connessione client nodo da contattare: ", err)
			}
		}

		err = client.Call("Successor.AddObject", arg, &result) //chiamo metodo, passando come argomento "keyboardArgoment" ed ottengo "result"
		if err != nil {
			// Gestisci l'errore se si verifica
			log.Fatal("Errore nella chiamata di metodo RPC: ", err)
		}

		*reply = result

	case 2:
		client, err := rpc.DialHTTP("tcp", nodeContact) //contatto il nodo che ho trovato prima.
		if err != nil {
			var opErr *net.OpError
			if errors.As(err, &opErr) {
				// Errore specifico dell'operazione di rete
				log.Fatalf("Errore connessione client nodo da contattare: %s, Op: %s, Net: %s", err, opErr.Op, opErr.Net)
			} else {
				// Altro tipo di errore
				log.Fatal("Errore connessione client nodo da contattare: ", err)
			}
		}

		err = client.Call("Successor.SearchObject", arg, &result) //iterativamente parte una ricerca tra i nodi usando le FT per trovare la risorsa.
		if err != nil {
			log.Fatal("Client invocation error: ", err)
		}

		*reply = result

	}

	return nil
}

func (t *Registry) GiveNodeLookup(idNodo int, ipNodo *string) error {
	*ipNodo = Nodes[idNodo]
	return nil

}

func (t *Registry) RemoveNode(arg *Arg, reply *string) error {
	idNodo := arg.Id
	//fmt.Println(Nodes)
	if len(Nodes) <= 2 {
		*reply = "Raggiunto il lower bound di 2 nodi nella rete. Impossibile cancellarne altri."
		return nil

	}
	if isNodePresent(Nodes, idNodo) {

		removedNode := Nodes[idNodo]
		var result string

		//fmt.Printf("\nDevo eliminare %d, ovvero %s \n", idNodo, removedNode) //ok, removedNode è quello da togliere.

		client, err := rpc.DialHTTP("tcp", removedNode) //contatto il nodo
		if err != nil {
			log.Fatal("Client connection error ask node 2 contact: ", err)
		}

		err = client.Call("Successor.UpdateNeighbors", idNodo, &result) //avvio la pratica per fargli aggiornare precedente e successivo
		if err != nil {
			log.Fatal("Client invocation error nel registry.removeNode: ", err)

		}

		delete(Nodes, idNodo)
		fmt.Printf("Nodi dopo la rimozione : %v\n", Nodes)

		*reply = "Il nodo avente id: '" + strconv.Itoa(idNodo) + "' è stato eliminato.\n"
	} else {
		*reply = "Il nodo avente id '" + strconv.Itoa(idNodo) + "' non è presente e dunque non è eliminabile.\n"
	}

	return nil

}

func main() {
	// Creazione di un nuovo oggetto Registry
	registry := new(Registry)
	rpc.Register(registry) //l'oggetto registry viene registrato per consentire la comunicazione RPC.
	rpc.HandleHTTP()       //La funzione HandleHTTP configura il pacchetto rpc per l'uso con il protocollo HTTP. Ciò consente al server RPC di gestire le richieste e le risposte RPC utilizzando il protocollo HTTP.
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil) //avvia un server HTTP che ascolta sul listener l e gestisce le richieste in arrivo utilizzando il gestore predefinito di http.DefaultServeMux.

}

func isNodePresent(nodes map[int]string, idNodo int) bool {
	for _, node := range nodes {
		if node == Nodes[idNodo] {
			return true
		}
	}
	return false
}
