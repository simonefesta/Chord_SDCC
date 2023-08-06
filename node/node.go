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
	//IN REALTA OBJECT SONO LE RISORSE NON E' FINGER TABLE
	Finger []int
}

type Arg struct { //ciò che passo ai metodi
	Id    int
	Value string
}

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

	client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}
	err = client.Call("Registry.Neighbors", arg, &result)
	if err != nil {
		log.Fatal("Client invocation error nel registry.neighbors: ", err)
	}

	return result

}

func CreateFingerTable(node *Node) error {
	arg := new(Arg)
	arg.Value = node.Ip
	arg.Id = sha_adapted(node.Ip)

	client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}
	err = client.Call("Registry.Finger", arg, &node.Finger)
	if err != nil {
		log.Fatal("Client invocation error nel registry.Finger: ", err)
	}

	return nil

}

func refreshNeighbors(node *Node) *Neighbors {

	result := new(Neighbors)
	arg := new(Arg)
	arg.Value = node.Ip
	arg.Id = sha_adapted(node.Ip)

	client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}
	err = client.Call("Registry.RefreshNeighbors", arg, &result)
	if err != nil {
		log.Fatal("Client invocation error nel registry.neighbors: ", err)
	}

	return result

}

func (t *Successor) UpdatePredecessor(nodoChiamante *Node, reply *string) error {
	node.Successor = nodoChiamante.Successor
	fmt.Printf("Il mio nuovo successore e' %s", node.Successor)
	CreateFingerTable(node) //devo rigenerarla

	return nil

}

func (t *Successor) UpdateSuccessor(nodoChiamante *Node, reply *string) error {
	node.Predecessor = nodoChiamante.Predecessor
	fmt.Printf("Il mio nuovo predecessore e' %s", node.Predecessor)
	for key, value := range nodoChiamante.Objects {
		node.Objects[key] = value
		fmt.Printf("Ho un nuovo elemento: %s \n", value)
		delete(nodoChiamante.Objects, key)

	}
	CreateFingerTable(node) //devo rigenerarla

	return nil

}

func (t *Successor) UpdateNeighbors(idNodo int, result *string) error {

	close(stopChan)
	fmt.Printf("chiudo la sessione")

	client, err := rpc.DialHTTP("tcp", node.Predecessor)
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}
	err = client.Call("Successor.UpdatePredecessor", node, &result)
	if err != nil {
		log.Fatal("Client invocation error nel registry.neighbors: ", err)
	}

	client, err = rpc.DialHTTP("tcp", node.Successor)
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}
	err = client.Call("Successor.UpdateSuccessor", node, &result)
	if err != nil {
		log.Fatal("Client invocation error nel registry.neighbors: ", err)
	}

	fmt.Printf("Aggiornamento completato\n")

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

/*
Qui immagino questa situazione:
un "nuovo" nodo nel sistema, che conosce precedente e successivo, deve prendersi le risorse che gli spettano.
Quindi quello che fa è vedere quale delle risorse che ha il successivo sono destinate a lui (se nodo nuovo è '1', precedente è '9' e successivo '5', il nuovo nodo prende le risorse di '5' che sono maggiori di 9 e minori uguali ad 1.
*/
func (t *Successor) Keys(arg *ArgId, reply *map[int]string) error {
	(*reply) = make(map[int]string)
	idPredecessor := sha_adapted(arg.Predecessor)
	for k := range node.Objects {
		if (arg.Id <= idPredecessor && (k <= arg.Id || k > idPredecessor)) || (k <= arg.Id && k > idPredecessor) { //in pratica vedo se gli oggetti sono compresi tra il nodo precedente e quello attuale. Però perchè il metodo è "Successor" se non fa nulla?
			(*reply)[k] = node.Objects[k] //dereferenzio reply (quindi vado sulla mappa) e poi mi sposto di k.
			delete(node.Objects, k)       //perchè toglierlo dalla lista??
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
		log.Fatal("client error connection during getKeys", err)
	}

	err = client.Call("Successor.Keys", arg, &reply) //ora gestisco questa chiamata
	if err != nil {
		log.Fatal("client error during Call 'Successor.Keys procedure", err)
	}
	return reply

}

func (t *Successor) AddObject(arg *Arg, reply *string) error {
	idRisorsa := sha_adapted(arg.Value)            //id risorsa da aggiugere
	idPredecessor := sha_adapted(node.Predecessor) //id del nodo che ha chiamato il successore (quindi il precedente del successore è il nodo che ha chiamato il successore)
	idSuccessor := sha_adapted(node.Successor)
	//l'idea è questa: risorsa = 3, il nodo chiamante è 1, il successivo è 5. chiamo '5' e gli dico di aggiungere risorsa '3' e che il suo nodo precedente è '1'.

	if (node.Id == idPredecessor && node.Id == idSuccessor) || (idRisorsa <= node.Id && idRisorsa > idPredecessor) || (idPredecessor > node.Id && (idRisorsa > idPredecessor || idRisorsa <= node.Id)) {
		//primo pezzo è se c'è un solo nodo, il secondo pezzo è il caso 'comune', il terzo pezzo è quando sto a fine anello, quindi dove può verificarsi che il successore di '9' sia '1', per via del modulo. (caso 2 libretto)
		if node.Objects[idRisorsa] != "" {
			*reply = "oggetto con id:  '" + node.Objects[idRisorsa] + "' già esistente!"
		} else {
			node.Objects[idRisorsa] = arg.Value
			*reply = fmt.Sprintf("Oggetto '%s' aggiunto con id: '%d'", arg.Value, idRisorsa)
			fmt.Println("pred - succ", node.Predecessor, node.Successor) //precedessore e successore
			fmt.Println(node.Objects)
		}
	} else { //devo provare col successivo! NB: QUI ANCORA NO FINGER TABLE, QUINDI ME LI GIRO TUTTI
		client, err := rpc.DialHTTP("tcp", node.Successor)
		if err != nil {
			log.Fatal("client connection AddObject successor error", err)
		}

		err = client.Call("Successor.AddObject", arg, &reply)
		if err != nil {
			log.Fatal("client call AddObject successor error", err)
		}

	}
	return nil
}

func (t *Successor) SearchObject(arg *Arg, reply *string) error {

	idRisorsa := arg.Id //id oggetto
	idPredecessor := sha_adapted(node.Predecessor)
	isFound := false
	if (idRisorsa <= node.Id && idRisorsa > idPredecessor) || (idPredecessor > node.Id && (idRisorsa > idPredecessor || idRisorsa <= node.Id)) {
		if node.Objects[idRisorsa] == "" {
			*reply = "L'oggetto cercato non è presente"
		} else {
			*reply = "L'oggetto con id cercato è '" + node.Objects[idRisorsa] + "', posseduto dal nodo '" + strconv.Itoa(node.Id) + "'. "
		}
	} else {
		// l'oggetto cercato non è nel nodo successore, quindi devo 'iterare', nb: questo poi dovrò farlo con la finger table}
		//devo lavorare da qui
		//trovo nella finger table il nodo da contattare
		var nodoContactId int                     //qui mi salvo l'id del nodo da contattare
		for i := 1; i < len(node.Finger)-1; i++ { //già sopra ho visto se la ho io, quindi anche se qui lascio questo check, non dovrei avere problemi.
			fmt.Printf("Sto cercando nella finger, esamino %d \n", node.Finger[i])
			if (node.Id < idRisorsa) && (idRisorsa <= node.Finger[1]) {
				nodoContactId = node.Finger[1]
				fmt.Printf("Ho trovato %d ", nodoContactId)
				isFound = true
				break
			} else if (idRisorsa >= node.Finger[i]) && (idRisorsa < node.Finger[i+1]) { //nel primo if inoltro al successore, nel secondo caso esamino la ft.
				nodoContactId = node.Finger[i]
				fmt.Printf("Ho trovato %d ", nodoContactId)
				isFound = true
				break
			}
		}
		if !isFound {
			nodoContactId = node.Finger[len(node.Finger)-1] //se l'id risorsa eccede tutta la mia ft, allora inoltro all'ultimo nodo conosciuto.
		}
		fmt.Printf("Io sono %d, non possiedo la risorsa con id %d, la vado a chiedere al nodo di ID: %d con stato %t \n", node.Id, idRisorsa, nodoContactId, isFound)
		//adesso devo chiedere al registry chi è questo nodo con indice
		var nodoContact string
		client, err := rpc.DialHTTP("tcp", "localhost:1234")
		if err != nil {
			log.Fatal("Client connection error ask node 2 contact: ", err)
		}
		err = client.Call("Registry.GiveNodeLookup", nodoContactId, &nodoContact)
		if err != nil {
			log.Fatal("Client invocation error nel registry.neighbors: ", err)
		}

		fmt.Printf("Il registry mi ha detto che l'id del nodo %d è associato al nodo %s", nodoContactId, nodoContact)

		client, err = rpc.DialHTTP("tcp", nodoContact)
		if err != nil {
			log.Fatal("Errore dialHttp nodocontact fingertable", err)
		}
		err = client.Call("Successor.SearchObject", arg, &reply)
		if err != nil {
			log.Fatal("Client call error", err)
		}
	}
	return nil
}

func scanRing(me *Node, stopChan <-chan struct{}) {
	for {
		select {
		case <-stopChan:
			fmt.Printf("FINE DEI GIOCHI")
			return
		default:
			time.Sleep(10 * time.Second)
			neightbors := refreshNeighbors(me) //successore nodo creato
			me.Successor = neightbors.Successor
			if neightbors.Predecessor != "" {
				me.Predecessor = neightbors.Predecessor
			}
			CreateFingerTable(me)
			fmt.Println(me.Finger)

		}

	}

}

func (t *Successor) CloseConn(idNodo int, result *string) error {
	fmt.Print("Mi hanno contattato per uccidermi!!!!")
	close(stopChan)
	return nil
}

func main() {
	arg := os.Args

	if len(arg) < 2 {
		log.Fatal("Invocazione con argomento ip:port")
	} //secondo argomento indirizzoIp 127.0.0.4:8084

	me := newNode(arg[1])
	neightbors := getNeighbors(me.Ip) //successore nodo creato
	me.Successor = neightbors.Successor
	me.Id = sha_adapted(me.Ip)
	me.Predecessor = neightbors.Predecessor

	//pred := getPredecessor(me) //cerco il precedente
	//me.predecessor = pred

	fmt.Println("Io sono ", me.Id)
	//fmt.Println("il mio successore è ", me.successor)
	//fmt.Println("il mio predecessore è", me.predecessor)
	me.Objects = getKeys(me)
	//CreateFingerTable(me)

	go scanRing(me, stopChan)

	successor := new(Successor) //ci si mette in ascolto per ricevere un messaggio in caso di join di nodi dopo il predecessore
	rpc.Register(successor)
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", me.Ip)
	if err != nil {
		log.Fatal("Listener error in node.go :", err)
	}
	http.Serve(listener, nil)

}
