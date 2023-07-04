package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

type Node struct {
	id          int
	ip          string
	predecessor string
	successor   string
	objects     map[int]string //mappo gli chiavi intere(id) in valori (string) //secondo me è il contrario, perchè le chiavi sono le righe, i valori gli id.
	//IN REALTA OBJECT SONO LE RISORSE NON E' FINGER TABLE

}

type Arg struct { //ciò che passo ai metodi
	Id    int
	Value string
}

var node *Node

func newNode(ip string) *Node {
	node = new(Node)
	node.ip = ip
	return node
}

type Neightbors struct {
	Successor   string
	Predecessor string
}

func getNeightbors(ip string) *Neightbors {

	result := new(Neightbors)
	arg := new(Arg)
	arg.Value = ip
	arg.Id = sha_adapted(ip)

	client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}
	err = client.Call("Registry.Neightbors", arg, &result)
	if err != nil {
		log.Fatal("Client invocation error nel registry.neightbors: ", err)
	}

	return result

}

func refreshNeightbors(node *Node) *Neightbors {

	result := new(Neightbors)
	arg := new(Arg)
	arg.Value = node.ip
	arg.Id = sha_adapted(node.ip)

	client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}
	err = client.Call("Registry.RefreshNeightbors", arg, &result)
	if err != nil {
		log.Fatal("Client invocation error nel registry.neightbors: ", err)
	}

	return result

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
	for k := range node.objects {
		if (arg.Id <= idPredecessor && (k <= arg.Id || k > idPredecessor)) || (k <= arg.Id && k > idPredecessor) { //in pratica vedo se gli oggetti sono compresi tra il nodo precedente e quello attuale. Però perchè il metodo è "Successor" se non fa nulla?
			(*reply)[k] = node.objects[k] //dereferenzio reply (quindi vado sulla mappa) e poi mi sposto di k.
			delete(node.objects, k)       //perchè toglierlo dalla lista??
		}
	}

	return nil
}

func getKeys(me *Node) map[int]string {
	reply := make(map[int]string) //slice vuota

	if me.successor == me.ip { //ci sono solo io nella rete?
		return reply
	}

	arg := new(ArgId)
	arg.Id = me.id
	arg.Predecessor = me.predecessor

	//io chiedo le chiavi/mi interfaccio sempre col successor
	client, err := rpc.DialHTTP("tcp", me.successor)
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
	id := sha_adapted(arg.Value)                   //id risorsa da aggiugere
	idPredecessor := sha_adapted(node.predecessor) //id del nodo che ha chiamato il successore (quindi il precedente del successore è il nodo che ha chiamato il successore)
	//l'idea è questa: risorsa =3, il nodo chiamante è 1, il successivo è 5. chiamo '5' e gli dico di aggiungere risorsa '3' e che il suo nodo precedente è '1'.
	if (id <= node.id && id > idPredecessor) || (idPredecessor > node.id && (id > idPredecessor || id <= node.id)) {
		//il primo pezzo è il caso 'comune', il secondo pezzo è quando sto a fine anello, quindi dove può verificarsi che il successore di '9' sia '1', per via del modulo. (caso 2 libretto)
		if node.objects[id] != "" {
			*reply = "oggetto con id:  " + node.objects[id] + " già esistente "
		} else {
			node.objects[id] = arg.Value
			*reply = fmt.Sprintf("Oggetto '%s' aggiunto con id: '%d'", arg.Value, id)
			fmt.Println("pred - succ", node.predecessor, node.successor) //precedessore e successore
			fmt.Println(node.objects)

		}
	} else { //devo provare col successivo! NB: QUI ANCORA NO FINGER TABLE, QUINDI ME LI GIRO TUTTI
		client, err := rpc.DialHTTP("tcp", node.successor)
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
	id := arg.Id //id oggetto
	idPredecessor := sha_adapted(node.predecessor)

	if (id <= node.id && id > idPredecessor) || (idPredecessor > node.id && (id > idPredecessor || id <= node.id)) {
		if node.objects[arg.Id] == "" {
			*reply = "L'oggetto cercato non è presente"
		} else {
			*reply = "L'oggetto con id cercato è " + node.objects[arg.Id] + " "
		}
	} else { // l'oggetto cercato non è nel nodo successore, quindi devo 'iterare', nb: questo poi dovrò farlo con la finger table}
		client, err := rpc.DialHTTP("tcp", node.successor)
		if err != nil {
			log.Fatal("Errore dialHttp node successor", err)
		}
		err = client.Call("Successor.SearchObject", arg, &reply)
		if err != nil {
			log.Fatal("Client call error", err)
		}
	}
	return nil
}

func main() {
	arg := os.Args

	if len(arg) < 2 {
		log.Fatal("Invocazione con argomento ip:port")
	} //secondo argomento indirizzoIp 127.0.0.4:8084

	me := newNode(arg[1])
	neightbors := getNeightbors(me.ip) //successore nodo creato
	me.successor = neightbors.Successor
	me.id = sha_adapted(me.ip)
	me.predecessor = neightbors.Predecessor

	//pred := getPredecessor(me) //cerco il precedente
	//me.predecessor = pred

	fmt.Println("Io sono ", me.id)
	fmt.Println("il mio successore è ", me.successor)
	fmt.Println("il mio predecessore è", me.predecessor)
	me.objects = getKeys(me)

	go scanRing(me)

	successor := new(Successor) //ci si mette in ascolto per ricevere un messaggio in caso di join di nodi dopo il predecessore
	rpc.Register(successor)
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", me.ip)
	if err != nil {
		log.Fatal("Listener error in node.go :", err)
	}
	http.Serve(listener, nil)

}

func scanRing(me *Node) {
	for {
		time.Sleep(10 * time.Second)
		neightbors := refreshNeightbors(me) //successore nodo creato
		me.successor = neightbors.Successor
		//me.id = sha_adapted(me.ip)
		if neightbors.Predecessor != "" {
			me.predecessor = neightbors.Predecessor
		}

		//pred := getPredecessor(me) //cerco il precedente
		//me.predecessor = pred

		fmt.Println("Io sono ", me.id)
		fmt.Println("il mio nuovo successore è ", me.successor)
		fmt.Println("il mio nuovo predecessore è", me.predecessor)
		//me.objects = getKeys(me)

	}

}
