package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

type Node struct {
	Id          int
	Ip          string
	Predecessor string
	Successor   string
	Objects     map[int]string //mapp <key int, string value>
	Finger      []int
}

type Arg struct {
	Id    int
	Value string //nel caso dei nodi è è il valore dell'indirizzo IP, per le risorse è il valore della risorsa.
	Type  bool
}

var RegistryFromInside string = "registry:1234"

var node *Node

var stopChan = make(chan struct{})

type Successor string

func newNode(ip string) *Node {
	node = new(Node)
	node.Ip = ip
	return node
}

type Neighbors struct {
	Successor   string
	Predecessor string
}

// funzione per ottenere i nodi vicini.
func getNeighbors(ip string) *Neighbors {

	result := new(Neighbors)
	arg := new(Arg)
	arg.Value = ip
	arg.Id = sha_adapted(ip)
	client, err := rpc.DialHTTP("tcp", RegistryFromInside)
	if err != nil {
		log.Fatal("Errore nodo getNeighbors: non riesco a contattare il registry dall'interno  ", err)
	}
	err = client.Call("Registry.Neighbors", arg, &result)
	if err != nil {
		log.Fatal("Errore nodo getNeighbors: non riesco a chiamare Registry.Neighbors  ", err)
	}

	client.Close()
	return result

}

// funzione per creare una finger table da parte del nodo. Il registry fornisce gli id da inserire nella FT.
func CreateFingerTable(node *Node) error {
	arg := new(Arg)
	arg.Value = node.Ip
	arg.Id = sha_adapted(node.Ip)
	client, err := rpc.DialHTTP("tcp", RegistryFromInside)
	if err != nil {
		log.Fatal("Errore nodo CreateFingerTable: non riesco a contattare il registry dall'interno  ", err)
	}
	err = client.Call("Registry.Finger", arg, &node.Finger)
	if err != nil {
		log.Fatal("Errore nodo CreateFingerTable: non riesco a chiamare Registry.Finger  ", err)
	}

	client.Close()
	return nil

}

// funzione che permette ad un nodo di refreshare la propria FT.
func refreshNeighbors(node *Node) *Neighbors {

	result := new(Neighbors)
	arg := new(Arg)
	arg.Value = node.Ip
	arg.Id = sha_adapted(node.Ip)

	client, err := rpc.DialHTTP("tcp", RegistryFromInside)
	if err != nil {
		log.Fatal("Errore nodo refreshNeighbors: non riesco a contattare il registry dall'interno  ", err)
	}
	err = client.Call("Registry.RefreshNeighbors", arg, &result)
	if err != nil {
		log.Fatal("Errore nodo refreshNeighbors: non riesco a chiamare Registry.RefreshNeighbors  ", err)
	}

	client.Close()
	return result

}

// aggiornamento dei nodi adiacenti ad un nodo prossimo alla rimozione.
func (t *Successor) UpdateNeighborsNodeRemoved(idNodo int, result *string) error {

	client, err := rpc.DialHTTP("tcp", node.Predecessor)
	if err != nil {
		log.Fatal("Errore nodo UpdateNeighborsNodeRemoved: non riesco a contattare il predecessore  ", err)
	}
	err = client.Call("Successor.UpdatePredecessorNodeRemoved", node, &result)
	if err != nil {
		log.Fatal("Errore nodo UpdateNeighborsNodeRemoved: non riesco a chiamare Registry.RefreshNeighbors  ", err)
	}

	client.Close()

	client, err = rpc.DialHTTP("tcp", node.Successor)
	if err != nil {
		log.Fatal("Errore nodo UpdateNeighborsNodeRemoved: non riesco a contattare il successore  ", err)
	}
	err = client.Call("Successor.UpdateSuccessorNodeRemoved", node, &result)
	if err != nil {
		log.Fatal("Errore nodo UpdateNeighborsNodeRemoved: non riesco a chiamare Successor.UpdateSuccessor ", err)
	}

	close(stopChan) //chiudi connessione
	client.Close()

	return nil

}

// funzione usata da un nodo per contattare il suo predecessore, per comunicargli un nuovo successore.
func (t *Successor) UpdatePredecessorNodeRemoved(nodoChiamante *Node, reply *string) error {
	node.Successor = nodoChiamante.Successor
	fmt.Printf("Node %d, il mio nuovo successore e' [%d]:%s \n", node.Id, sha_adapted(node.Successor), node.Successor)

	return nil

}

// funzione usata da un nodo per contattare il suo sucessore, per comunicargli un nuovo predecessore. Il nodo successore preleva le risorse lasciate dal nodo uscente.
func (t *Successor) UpdateSuccessorNodeRemoved(nodoChiamante *Node, reply *string) error {
	node.Predecessor = nodoChiamante.Predecessor
	fmt.Printf("Node %d, il mio nuovo predecessore e'[%d]:%s \n", node.Id, sha_adapted(node.Predecessor), node.Predecessor)
	for key, value := range nodoChiamante.Objects {
		node.Objects[key] = value
		fmt.Printf("Node %d, ho un nuovo elemento: %s \n", node.Id, value)
		delete(nodoChiamante.Objects, key)

	}

	return nil

}

type Args struct { //argomenti da passare al metodo remoto Successor
	Ip        string
	CurrentIp string
}

// struct usata per la func 'Keys'.
type ArgId struct {
	Id          int
	Predecessor string
}

// processo logico che permette ad un nodo entrante l'adozione di risorse precedentemente gestite da un altro nodo.
func (t *Successor) Keys(arg *ArgId, reply *map[int]string) error {
	(*reply) = make(map[int]string)
	idPredecessor := sha_adapted(arg.Predecessor)
	for k := range node.Objects {
		if (arg.Id <= idPredecessor && (k <= arg.Id || k > idPredecessor)) || (k <= arg.Id && k > idPredecessor) {
			(*reply)[k] = node.Objects[k]
			delete(node.Objects, k)
		}
	}

	return nil
}

// funzione che permette ad un nodo di chiamare il successore per verificare la presenza di risorse a lui destinate.
func getKeys(me *Node) map[int]string {
	reply := make(map[int]string)

	if me.Successor == me.Ip {
		return reply
	}

	arg := new(ArgId)
	arg.Id = me.Id
	arg.Predecessor = me.Predecessor

	client, err := rpc.DialHTTP("tcp", me.Successor)
	if err != nil {
		log.Fatal("Errore nodo getKeys: non riesco a contattare il successore (getKeys)  ", err)
	}

	err = client.Call("Successor.Keys", arg, &reply)
	if err != nil {
		log.Fatal("Errore nodo getKeys: non riesco a chiamare Successor.Keys ", err)
	}
	client.Close()
	return reply

}

/*
Funzione per l'inserimento di un nuovo oggetto da memorizzare.

Il primo if esegue tre controlli per vedere se la risorsa richiesta dovrebbe essere gestita dal nodo nella funzione:
se c'è solo un nodo nella rete, se l'id è di sua competenza, oppure se l'id è di sua competenza ma per via del modulo il check precedente non può essere soddisfatto.
(ultimo caso: ho un anello con nodo '9' e '1', la risorsa 10 è gestita da '1', anche se idRisorsa>1)

Se l'oggetto non è di competenza del nodo in questione, cerca nella sua FT il nodo che dovrebbe gestirla.
*/
func (t *Successor) AddObject(arg *Arg, reply *string) error {

	idRisorsa := sha_adapted(arg.Value)
	idPredecessor := sha_adapted(node.Predecessor)
	idSuccessor := sha_adapted(node.Successor)

	if (node.Id == idPredecessor && node.Id == idSuccessor) || (idRisorsa <= node.Id && idRisorsa > idPredecessor) || (idPredecessor > node.Id && (idRisorsa > idPredecessor || idRisorsa <= node.Id)) {
		if node.Objects[idRisorsa] != "" {
			*reply = "oggetto con id:  '" + node.Objects[idRisorsa] + "' già esistente!\n"
		} else {
			node.Objects[idRisorsa] = arg.Value
			*reply = fmt.Sprintf("Oggetto '%s' aggiunto con id: '%d'\n", arg.Value, idRisorsa)
			fmt.Println(node.Objects)
		}
	} else {

		isFound := false

		var nodoContactId int
		for i := 1; i < len(node.Finger)-1; i++ {
			if (idRisorsa <= node.Finger[1]) && (node.Id < idRisorsa) {
				nodoContactId = node.Finger[1]
				isFound = true
				break
			} else if (idRisorsa >= node.Finger[i]) && (idRisorsa < node.Finger[i+1]) {
				nodoContactId = node.Finger[i]
				isFound = true
				break
			}
		}
		if !isFound {
			nodoContactId = node.Finger[len(node.Finger)-1]
		}

		var nodoContact string
		client, err := rpc.DialHTTP("tcp", RegistryFromInside)
		if err != nil {
			log.Fatal("Errore nodo AddObject: non riesco a contattare il registry dall'interno  ", err)
		}
		err = client.Call("Registry.GiveNodeLookup", nodoContactId, &nodoContact)
		if err != nil {
			log.Fatal("Errore nodo AddObject: non riesco a chiamare Registry.GiveNodeLookup ", err)
		}

		client.Close()

		client, err = rpc.DialHTTP("tcp", nodoContact)
		if err != nil {
			log.Fatal("Errore nodo AddObject: non riesco a contattare il nodo trovato sulla FT  ", err)
		}

		err = client.Call("Successor.AddObject", arg, &reply)
		if err != nil {
			log.Fatal("Errore nodo AddObject: non riesco a chiamare Successor.AddObject ", err)
		}

		client.Close()

	}

	return nil
}

/*
Ricerca di un oggetto.
Vedo se la risorsa è di mia competenza in base al consistent hashing. Se lo è, allora devo vedere se ce l'ho o meno.
Se non è di mia competenza, contatto il nodo che dovrebbe gestirla secondo la mia FT.
Se l'id è "lontano" dalla mia FT, contatto l'ultimo nodo presente nella mia FT.
*/
func (t *Successor) SearchObject(arg *Arg, reply *string) error {

	idRisorsa := arg.Id //id oggetto
	idPredecessor := sha_adapted(node.Predecessor)
	isFound := false
	if (idRisorsa <= node.Id && idRisorsa > idPredecessor) || (idPredecessor > node.Id && (idRisorsa > idPredecessor || idRisorsa <= node.Id)) {
		if node.Objects[idRisorsa] == "" { //se non c'è
			*reply = "L'oggetto cercato non è presente.\n"
		} else {
			if arg.Type { //se arg.Type == true, allora la ricerca l'ho fatta per rimuovere l'oggetto dal nodo.
				delete(node.Objects, idRisorsa)
				*reply = "L'oggetto con id  '" + strconv.Itoa(idRisorsa) + "' è stato rimosso.\n"

			} else {
				*reply = "L'oggetto con id cercato è '" + node.Objects[idRisorsa] + "', posseduto dal nodo '" + strconv.Itoa(node.Id) + "'.\n"
			}
		}
	} else {

		var nodoContactId int //qui mi salvo l'id del nodo da contattare
		for i := 1; i < len(node.Finger)-1; i++ {
			if (idRisorsa <= node.Finger[1]) && (node.Id < idRisorsa) {
				nodoContactId = node.Finger[1]
				isFound = true
				break
			} else if (idRisorsa >= node.Finger[i]) && (idRisorsa < node.Finger[i+1]) {
				nodoContactId = node.Finger[i]
				isFound = true
				break
			}
		}
		if !isFound {
			nodoContactId = node.Finger[len(node.Finger)-1] //se l'id risorsa eccede tutta la mia ft, allora inoltro all'ultimo nodo conosciuto.
		}
		//adesso devo chiedere al registry chi è questo nodo con indice
		var nodoContact string
		client, err := rpc.DialHTTP("tcp", RegistryFromInside)
		if err != nil {
			log.Fatal("Errore nodo SearchObject: non riesco a contattare il registry dall'interno  ", err)
		}
		err = client.Call("Registry.GiveNodeLookup", nodoContactId, &nodoContact)
		if err != nil {
			log.Fatal("Errore nodo SearchObject: non riesco a chiamare Registry.GiveNodeLookup ", err)
		}

		client.Close()

		client, err = rpc.DialHTTP("tcp", nodoContact)
		if err != nil {
			log.Fatal("Errore nodo SearchObject: non riesco a contattare il nodo trovato sulla FT  ", err)
		}
		err = client.Call("Successor.SearchObject", arg, &reply)
		if err != nil {
			log.Fatal("Errore nodo SearchObject: non riesco a chiamare Successor.SearchObject ", err)
		}

		client.Close()

	}
	return nil
}

func scanRing(me *Node, stopChan <-chan struct{}) {
	isPrint := 1 //variabile che permette di stampare ogni isPrint * 5 secondi. Ogni 5 secondi controllo i vicini, ogni 15 vorrei stampare le FT (5*3, dove 3 è il valore per entrare in stampa.)
	for {
		select {
		case <-stopChan:
			fmt.Printf("Connessione interrotta correttamente.")
			return
		default:
			time.Sleep(5 * time.Second)
			neightbors := refreshNeighbors(me) //successore nodo creato
			me.Successor = neightbors.Successor
			if neightbors.Predecessor != "" {
				me.Predecessor = neightbors.Predecessor
			}
			CreateFingerTable(me)
			if isPrint == 3 { //aggiorno ogni 5 secondi la fingertable, però non la mostro sempre (troppe info su schermo). Alla fine la stampo ogni 10 secondi.
				fmt.Printf("FT[%d]: ", me.Id)
				for i := 1; i <= len(me.Finger)-1; i++ {
					fmt.Printf("<%d,%d> ", i, me.Finger[i])
				}
				fmt.Printf("\n")
				isPrint = 1

			}
			isPrint++
		}

	}

}

func main() {

	ipAddress, err := getLocalIP()
	if err != nil {
		fmt.Println("Errore nell'ottenere l'indirizzo IP:", err)
		return
	}

	ipPort := os.Getenv("NODE_PORT")
	if ipPort == "" {
		fmt.Println("La porta non è specificata.")
		return
	}

	ipPortString := fmt.Sprintf("%s:%s", ipAddress, ipPort)

	me := newNode(ipPortString)
	neightbors := getNeighbors(me.Ip) //successore nodo creato
	me.Successor = neightbors.Successor
	me.Id = sha_adapted(me.Ip)
	me.Predecessor = neightbors.Predecessor

	fmt.Printf("Io sono %d, Indirizzo IP:%s\n", me.Id, ipPortString)

	me.Objects = getKeys(me)

	go scanRing(me, stopChan)

	successor := new(Successor) //ci si mette in ascolto per ricevere un messaggio in caso di join di nodi dopo il predecessore
	rpc.Register(successor)
	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", ipPortString)
	if err != nil {
		log.Fatal("Listener error in node.go :", err)
	}

	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatal("Errore nel server HTTP: ", err)
	}

}
