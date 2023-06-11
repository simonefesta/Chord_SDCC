package main

import (
	"fmt"
	"log"
	"net/rpc"
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
	id    int
	value string
}

var node *Node

func newNode(ip string) *Node {
	node = new(Node)
	node.ip = ip
	return node
}

func getSuccessorIp(ip string) string {

	var result string
	arg := new(Arg)
	arg.value = ip
	arg.id = sha_adapted(ip)

	client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}
	err = client.Call("Registry.Successor", arg, &result)
	if err != nil {
		log.Fatal("Client invocation error: ", err)
	}

	return result

}

type Successor string

type Args struct { //argomenti da passare al metodo remoto Successor
	ip        string
	currentIp string
}

//****QUI STO FACENDO I COLLEGAMENTI TRA NODI, SPECIFICANDO NODO IP CORRENTE & NODO A CUI COLLEGARMI.
//method available for remote access must be like func (t *T) MethodName(argType T1, replyType *T2) error

func (t *Successor) Predecessor(args *Args, reply *string) error { //DA RIVEDEERE
	*reply = node.predecessor
	if node.predecessor == node.successor { //qui il campo predecessor deve essere marcato
		node.successor = args.currentIp //se nodo precedente e successore sono uguali, vuol dire che ho una rete di un nodo? Quindi sono io il successore di me stesso
	} //non dovrebbe essere args.ip?
	if reply == nil { //vedo se puntatore è nullo
		reply = &args.ip
	} else {
		node.predecessor = args.currentIp
	}
	return nil

}

func getPredecessor(me *Node) string { //qui il nodo stesso, localmente, cerca il suo predecessore

	var reply string
	if me.successor == me.ip { //successore di me stesso
		return me.ip
	}
	args := new(Args)
	args.ip = me.successor //quindi ip è del successore?
	args.currentIp = me.ip //questo sono io

	client, err := rpc.DialHTTP("tcp", me.successor)
	if err != nil {
		log.Fatal("Client connection error: ", err)
	}

	err = client.Call("Successor.Predecessor", args, &reply)
	if err != nil {
		log.Fatal("errore invocazione chiamata remota", err)

	}
	return reply

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
	id := sha_adapted(arg.value)                   //id risorsa da aggiugere
	idPredecessor := sha_adapted(node.predecessor) //id del nodo che ha chiamato il successore (quindi il precedente del successore è il nodo che ha chiamato il successore)
	//l'idea è questa: risorsa =3, il nodo chiamante è 1, il successivo è 5. chiamo '5' e gli dico di aggiungere risorsa '3' e che il suo nodo precedente è '1'.
	if (id <= node.id && id > idPredecessor) || (idPredecessor > node.id && (id > idPredecessor || id <= node.id)) {
		//il primo pezzo è il caso 'comune', il secondo pezzo è quando sto a fine anello, quindi dove può verificarsi che il successore di '9' sia '1', per via del modulo. (caso 2 libretto)
		if node.objects[id] != "" {
			*reply = "oggetto con id:  " + node.objects[id] + " già esistente "
		} else {
			node.objects[id] = arg.value
			*reply = fmt.Sprintf("Oggetto '%s' aggiunto con id: '%d'", arg.value, id)
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

//todo search object
//todo main
