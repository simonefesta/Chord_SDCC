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
	Objects     map[int]string //mappo le chiavi intere(id) in valori (string) //secondo me è il contrario, perchè le chiavi sono le righe, i valori gli id.
	Finger      []int
}

type Arg struct { //ciò che passo ai metodi
	Id    int
	Value string //ip if is node, object if is resource
	Type  bool
}

var RegistryFromInside string = "registry:1234"

var node *Node

var stopChan = make(chan struct{})

func newNode(ip string) *Node {
	node = new(Node)
	node.Ip = ip
	return node
}

type Neighbors struct {
	Successor   string
	Predecessor string
}

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
		log.Fatal("Errore nodo CreateFingerTable: non riesco a chiamare Registry.RefreshNeighbors  ", err)
	}

	client.Close()
	return result

}

func (t *Successor) UpdatePredecessor(nodoChiamante *Node, reply *string) error {
	node.Successor = nodoChiamante.Successor
	fmt.Printf("Node %d, il mio nuovo successore e' [%d]:%s \n", node.Id, sha_adapted(node.Successor), node.Successor)

	return nil

}

func (t *Successor) UpdateSuccessor(nodoChiamante *Node, reply *string) error {
	node.Predecessor = nodoChiamante.Predecessor
	fmt.Printf("Node %d, il mio nuovo predecessore e'[%d]:%s \n", node.Id, sha_adapted(node.Predecessor), node.Predecessor)
	for key, value := range nodoChiamante.Objects {
		node.Objects[key] = value
		fmt.Printf("Node %d, ho un nuovo elemento: %s \n", node.Id, value)
		delete(nodoChiamante.Objects, key)

	}

	return nil

}

func (t *Successor) UpdateNeighbors(idNodo int, result *string) error { //questo quando ELIMINO UN NODO

	client, err := rpc.DialHTTP("tcp", node.Predecessor)
	if err != nil {
		log.Fatal("Errore nodo UpdateNeighbors: non riesco a contattare il predecessore  ", err)
	}
	err = client.Call("Successor.UpdatePredecessor", node, &result)
	if err != nil {
		log.Fatal("Errore nodo UpdateNeighbors: non riesco a chiamare Registry.RefreshNeighbors  ", err)
	}

	client.Close()

	client, err = rpc.DialHTTP("tcp", node.Successor)
	if err != nil {
		log.Fatal("Errore nodo UpdateNeighbors: non riesco a contattare il successore  ", err)
	}
	err = client.Call("Successor.UpdateSuccessor", node, &result)
	if err != nil {
		log.Fatal("Errore nodo UpdateNeighbors: non riesco a chiamare Successor.UpdateSuccessor ", err)
	}

	close(stopChan) //chiudi connessione
	client.Close()

	return nil

}

type Successor string

type Args struct { //argomenti da passare al metodo remoto Successor
	Ip        string
	CurrentIp string
}

// **OLTRE AL NODO DEVO SPECIFICARE ANCHE LE CHIAVI
type ArgId struct {
	Id          int
	Predecessor string
}

func (t *Successor) Keys(arg *ArgId, reply *map[int]string) error {
	(*reply) = make(map[int]string)
	idPredecessor := sha_adapted(arg.Predecessor)
	for k := range node.Objects {
		if (arg.Id <= idPredecessor && (k <= arg.Id || k > idPredecessor)) || (k <= arg.Id && k > idPredecessor) { //in pratica vedo se gli oggetti sono compresi tra il nodo precedente e quello attuale. Però perchè il metodo è "Successor" se non fa nulla?
			(*reply)[k] = node.Objects[k] //dereferenzio reply (quindi vado sulla mappa) e poi mi sposto di k.
			delete(node.Objects, k)
		}
	}

	return nil
}

func getKeys(me *Node) map[int]string {
	reply := make(map[int]string) //slice vuota

	if me.Successor == me.Ip { //ci sono solo io nella rete?
		return reply
	}

	arg := new(ArgId)
	arg.Id = me.Id
	arg.Predecessor = me.Predecessor

	//io chiedo le chiavi/mi interfaccio sempre col successor
	client, err := rpc.DialHTTP("tcp", me.Successor)
	if err != nil {
		log.Fatal("Errore nodo getKeys: non riesco a contattare il successore (getKeys)  ", err)
	}

	err = client.Call("Successor.Keys", arg, &reply) //ora gestisco questa chiamata
	if err != nil {
		log.Fatal("Errore nodo getKeys: non riesco a chiamare Successor.Keys ", err)
	}
	client.Close()
	return reply

}

func (t *Successor) AddObject(arg *Arg, reply *string) error {

	idRisorsa := sha_adapted(arg.Value)            //id risorsa da aggiugere
	idPredecessor := sha_adapted(node.Predecessor) //id del nodo che ha chiamato il successore (quindi il precedente del successore è il nodo che ha chiamato il successore)
	idSuccessor := sha_adapted(node.Successor)

	if (node.Id == idPredecessor && node.Id == idSuccessor) || (idRisorsa <= node.Id && idRisorsa > idPredecessor) || (idPredecessor > node.Id && (idRisorsa > idPredecessor || idRisorsa <= node.Id)) {
		//primo pezzo è se c'è un solo nodo, il secondo pezzo è il caso 'comune', il terzo pezzo è quando sto a fine anello, quindi dove può verificarsi che il successore di '9' sia '1', per via del modulo. (caso 2 libretto)
		if node.Objects[idRisorsa] != "" {
			*reply = "oggetto con id:  '" + node.Objects[idRisorsa] + "' già esistente!\n"
		} else {
			node.Objects[idRisorsa] = arg.Value
			*reply = fmt.Sprintf("Oggetto '%s' aggiunto con id: '%d'\n", arg.Value, idRisorsa)
			fmt.Println(node.Objects)
		}
	} else {

		isFound := false

		var nodoContactId int                     //qui mi salvo l'id del nodo da contattare
		for i := 1; i < len(node.Finger)-1; i++ { //già sopra ho visto se la ho io, quindi anche se qui lascio questo check, non dovrei avere problemi.
			if (idRisorsa <= node.Finger[1]) && (node.Id < idRisorsa) { // è del successore
				nodoContactId = node.Finger[1]
				isFound = true
				break
			} else if (idRisorsa >= node.Finger[i]) && (idRisorsa < node.Finger[i+1]) { //qui esamino la FT.
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

func (t *Successor) SearchObject(arg *Arg, reply *string) error {

	idRisorsa := arg.Id //id oggetto
	idPredecessor := sha_adapted(node.Predecessor)
	isFound := false
	if (idRisorsa <= node.Id && idRisorsa > idPredecessor) || (idPredecessor > node.Id && (idRisorsa > idPredecessor || idRisorsa <= node.Id)) { //se lo statement è true, vuol dire che se l'oggetto esiste, devo averlo io.
		if node.Objects[idRisorsa] == "" { //se non c'è
			*reply = "L'oggetto cercato non è presente.\n"
		} else {
			if arg.Type { //devo rimuovere la chiave
				delete(node.Objects, idRisorsa)
				*reply = "L'oggetto con id  '" + strconv.Itoa(idRisorsa) + "' è stato rimosso.\n"

			} else {
				*reply = "L'oggetto con id cercato è '" + node.Objects[idRisorsa] + "', posseduto dal nodo '" + strconv.Itoa(node.Id) + "'.\n"
			}
		}
	} else {
		//devo lavorare da qui
		//trovo nella finger table il nodo da contattare
		var nodoContactId int                     //qui mi salvo l'id del nodo da contattare
		for i := 1; i < len(node.Finger)-1; i++ { //già sopra ho visto se la ho io, quindi anche se qui lascio questo check, non dovrei avere problemi.
			if (idRisorsa <= node.Finger[1]) && (node.Id < idRisorsa) { // è del successore
				nodoContactId = node.Finger[1]
				isFound = true
				break
			} else if (idRisorsa >= node.Finger[i]) && (idRisorsa < node.Finger[i+1]) { //qui esamino la FT.
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
	isPrint := true
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
			if isPrint { //aggiorno ogni 5 secondi la fingertable, però non la mostro sempre (troppe info su schermo). Alla fine la stampo ogni 10 secondi.
				fmt.Printf("FT[%d] : ", me.Id)
				for i := 1; i <= len(me.Finger)-1; i++ {
					fmt.Printf("<%d,%d> ", i, me.Finger[i])
				}
				fmt.Printf("\n")
			}
			isPrint = !isPrint
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
