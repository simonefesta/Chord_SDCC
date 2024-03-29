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
	Id          int
	Ip          string
	Predecessor string
	Successor   string
	Objects     map[int]string //mapp <key int, string value>
	Finger      []int
}

type Arg struct {
	Id     int
	Value  string //nel caso dei nodi è è il valore dell'indirizzo IP, per le risorse è il valore della risorsa.
	Type   bool
	PredOp string //Nelle operazioni di inserimento, ricerca o cancellazione è possibile chiedere al predecessore se gestisca lui la risorsa. A seconda dell'operazione richiesta, esplicitata da PredOp, si avrà un comportamento diverso.
}

var RegistryFromInside string = "registry:1234"

var node *Node

var stopChan = make(chan struct{})

type OtherNode string

func newNode(ip string) *Node {
	node = new(Node)
	node.Ip = ip
	return node
}

type Neighbors struct {
	Successor   string
	Predecessor string
	MiddleNode  string
}

/*
******* Gestione dei nodi adiacenti, all'entrata di un nuovo nodo *************************************************************************

Il nodo chiede al Registry quali nodi siano i suoi vicini,
per poi contattarli direttamente in modo da aggiornare le loro conoscenze.
La logica per l'assegnamento delle risorse è separata da questa funzione, poichè, a differenza di quando un nodo esce
e lascia tutte le sue risorse al successore, nel caso dell'entrata di un nodo, non è detto che tutte le risorse del
successore saranno affidate a lui.
*/

func getNeighbors(ip string) *Neighbors {
	var result string
	neighbors := new(Neighbors)
	arg := new(Arg)
	arg.Value = ip
	arg.Id = sha_consistent(ip)
	client, err := rpc.DialHTTP("tcp", RegistryFromInside)
	if err != nil {
		log.Fatal("Errore nodo getNeighbors: non riesco a contattare il registry dall'interno  ", err)
	}
	err = client.Call("Registry.Neighbors", arg, &neighbors)
	if err != nil {
		log.Fatal("Errore nodo getNeighbors: non riesco a chiamare Registry.Neighbors  ", err)
	}
	client.Close()
	neighbors.MiddleNode = ip
	if (ip != neighbors.Successor) && (ip != neighbors.Predecessor) { //se ip = successor = predecessor mi starei contattando da solo!
		if neighbors.Successor != "" {

			client, err = rpc.DialHTTP("tcp", neighbors.Successor)
			if err != nil {
				log.Fatal("Errore nodo getNeighbors: non riesco a contattare il neighbors.Successor  ", err)
			}
			err = client.Call("OtherNode.UpdateSuccessorNode", neighbors.MiddleNode, &result)
			if err != nil {
				log.Fatal("Errore nodo getNeighbors: non riesco a chiamare OtherNode.UpdateSuccessor  ", err)
			}
			client.Close()
		}
		if neighbors.Predecessor != "" {
			client, err = rpc.DialHTTP("tcp", neighbors.Predecessor)
			if err != nil {
				log.Fatal("Errore nodo getNeighbors: non riesco a contattare il neighbors.Predecessor  ", err)
			}
			err = client.Call("OtherNode.UpdatePredecessorNode", neighbors.MiddleNode, &result)
			if err != nil {
				log.Fatal("Errore nodo getNeighbors: non riesco a chiamare OtherNode.UpdatePredecessor  ", err)
			}
			client.Close()
		}

	}

	return neighbors

}

func (t *OtherNode) UpdatePredecessorNode(nodoChiamante string, reply *string) error {
	node.Successor = nodoChiamante
	fmt.Printf("Node %d, il mio nuovo successore e' [%d]:%s \n", node.Id, sha_consistent(node.Successor), node.Successor)

	return nil

}

func (t *OtherNode) UpdateSuccessorNode(nodoChiamante string, reply *string) error {
	node.Predecessor = nodoChiamante
	fmt.Printf("Node %d, il mio nuovo predecessore e' [%d]:%s \n", node.Id, sha_consistent(node.Predecessor), node.Predecessor)

	return nil

}

/*
******* Gestione delle risorse, entrata di un nuovo nodo. ******************************************************************************************

Qui il nodo successore al nodo chiamante, verifica quali risorse dovranno essere affidate al nodo chiamante.
Ciò viene realizzato mediante controllo sugli identificativi, tenendo in considerazione anche i casi "limite",
ovvero quando ci si pone a inizio/fine nodo.
*/

// struct usata per la func 'Keys'.
type ArgId struct {
	Id          int
	Predecessor string
}

func (t *OtherNode) Keys(arg *ArgId, reply *map[int]string) error {
	(*reply) = make(map[int]string)
	idPredecessor := sha_consistent(arg.Predecessor)
	for k := range node.Objects {
		if (arg.Id <= idPredecessor && (k <= arg.Id || k > idPredecessor)) || (k <= arg.Id && k > idPredecessor) {
			(*reply)[k] = node.Objects[k]
			fmt.Printf("Node %d, ho un nuovo elemento: <%d, %s> \n", arg.Id, k, node.Objects[k])
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

	err = client.Call("OtherNode.Keys", arg, &reply)
	if err != nil {
		log.Fatal("Errore nodo getKeys: non riesco a chiamare OtherNode.Keys ", err)
	}
	client.Close()
	return reply

}

/*
******* Gestione dei nodi adiacenti, alla rimozione di un  nodo *******************************************************************************

Il nodo che verrà rimosso viene contattato. Tale nodo contatterà i suoi nodi predecessori e successori,
per metterli in comunicazione. Passerà TUTTE le sue risorse al nodo successore.
*/

func (t *OtherNode) UpdateNeighborsNodeRemoved(idNodo int, result *string) error {

	client, err := rpc.DialHTTP("tcp", node.Predecessor)
	if err != nil {
		log.Fatal("Errore nodo UpdateNeighborsNodeRemoved: non riesco a contattare il predecessore  ", err)
	}
	err = client.Call("OtherNode.UpdatePredecessorNodeRemoved", node, &result)
	if err != nil {
		log.Fatal("Errore nodo UpdateNeighborsNodeRemoved: non riesco a chiamare Registry.RefreshNeighbors  ", err)
	}

	client.Close()

	client, err = rpc.DialHTTP("tcp", node.Successor)
	if err != nil {
		log.Fatal("Errore nodo UpdateNeighborsNodeRemoved: non riesco a contattare il successore  ", err)
	}
	err = client.Call("OtherNode.UpdateSuccessorNodeRemoved", node, &result)
	if err != nil {
		log.Fatal("Errore nodo UpdateNeighborsNodeRemoved: non riesco a chiamare OtherNode.UpdateSuccessorNodeRemoved ", err)
	}

	close(stopChan) //chiudi connessione
	client.Close()

	return nil

}

// funzione usata da un nodo per contattare il suo predecessore, per comunicargli un nuovo successore.
func (t *OtherNode) UpdatePredecessorNodeRemoved(nodoChiamante *Node, reply *string) error {
	node.Successor = nodoChiamante.Successor
	fmt.Printf("Node %d, il mio nuovo successore e' [%d]:%s \n", node.Id, sha_consistent(node.Successor), node.Successor)

	return nil

}

// funzione usata da un nodo per contattare il suo sucessore, per comunicargli un nuovo predecessore. Il nodo successore preleva le risorse lasciate dal nodo uscente.
func (t *OtherNode) UpdateSuccessorNodeRemoved(nodoChiamante *Node, reply *string) error {
	node.Predecessor = nodoChiamante.Predecessor
	fmt.Printf("Node %d, il mio nuovo predecessore e'[%d]:%s \n", node.Id, sha_consistent(node.Predecessor), node.Predecessor)
	for key, value := range nodoChiamante.Objects {
		node.Objects[key] = value
		fmt.Printf("Node %d, ho un nuovo elemento: <%d, %s> \n", node.Id, key, value)
		delete(nodoChiamante.Objects, key)

	}

	return nil

}

/*
******* Funzione per l'inserimento di un nuovo oggetto da memorizzare **************************************************************************************************

Il primo if esegue tre controlli per vedere se la risorsa richiesta deve essere gestita dal nodo nella funzione:
se c'è solo un nodo nella rete, se l'id è di sua competenza, oppure se l'id è di sua competenza ma per via del modulo il check precedente non può essere soddisfatto.
(esempio ultimo caso: ho un anello con nodo '9' e '1', la risorsa 10 è gestita da '1', anche se idRisorsa>1)
Se l'oggetto non è di competenza del nodo in questione, cerca nella sua FT il nodo che dovrebbe gestirla.
Se non lo trova, inoltra la richiesta al nodo più distante.
Viene anche fatta una chiamata al predecessore del nodo in questione, per velocizzare la richiesta.
*/
func (t *OtherNode) AddObject(arg *Arg, reply *string) error {

	idRisorsa := sha_consistent(arg.Value)
	idPredecessor := sha_consistent(node.Predecessor)
	idSuccessor := sha_consistent(node.Successor)

	if (node.Id == idPredecessor && node.Id == idSuccessor) || (idRisorsa <= node.Id && idRisorsa > idPredecessor) || (idPredecessor > node.Id && (idRisorsa > idPredecessor || idRisorsa <= node.Id)) {
		if node.Objects[idRisorsa] != "" {
			*reply = fmt.Sprintf("L'oggetto con id: '%d' è già esistente!\n", idRisorsa)
			return nil
		} else {
			node.Objects[idRisorsa] = arg.Value
			*reply = fmt.Sprintf("Oggetto '%s' aggiunto con id: '%d'\n", arg.Value, idRisorsa)
			fmt.Printf("Nodo: %d, %v\n", node.Id, node.Objects)
			return nil
		}
	}

	//se procedo, vuol dire che il nodo in questione non gestisce la risorsa. Proviamo a chiedere al predecessore.
	client, err := rpc.DialHTTP("tcp", node.Predecessor)
	if err != nil {
		log.Fatal("Errore nodo AddObject: non riesco a contattare il predecessore", err)
	}

	arg.PredOp = "add"      //parametro per AskPredecessor, per specificare che stiamo effettuando una operazione di aggiunta.
	arg.Id = idRisorsa      //risorsa che dovrà cercare il nodo predecessore o il nodo a cui affiderò la richiesta.
	resultPredecessor := "" // inizializzo una stringa vuota, in cui il predecessore immetterà il risultato della ricerca: se è di sua competenza (sia che l'oggetto sia da inserire, o già presente), allora scriverà qualcosa.

	err = client.Call("OtherNode.AskPredecessor", arg, &resultPredecessor)
	if err != nil {
		log.Fatal("Errore nodo AddObject: non riesco a chiamare OtherNode.AskPredecessor ", err)
	}

	client.Close()

	if resultPredecessor != "" { //se la stringa è non vuota, allora l'operazione è stata gestita dal predecessore. Non serve cercare altro.

		*reply = resultPredecessor
		return nil

	} else {
		nodoContact := InterrogateFinger(arg, reply) //altrimenti mi appoggio a questa funzione che interroga la FT per trovare il nodo idoneo alla gestione dell'operazione.

		client, err := rpc.DialHTTP("tcp", nodoContact)
		if err != nil {
			log.Fatal("Errore nodo AddObject: non riesco a contattare il nodo fornito dal registry  ", err)
		}

		err = client.Call("OtherNode.AddObject", arg, &reply)
		if err != nil {
			log.Fatal("Errore nodo AddObject: non riesco a chiamare OtherNode.AddObject ", err)
		}

		client.Close()
	}

	return nil
}

/*
*************** Ricerca di un oggetto ************************************************************************************
Vedo se la risorsa è di mia competenza in base al consistent hashing. Se lo è, allora devo vedere se ce l'ho o meno.
Se non è di mia competenza, contatto il nodo che dovrebbe gestirla secondo la mia FT.
Se l'id è "lontano" dalla mia FT, contatto l'ultimo nodo presente nella mia FT.
*/
func (t *OtherNode) SearchObject(arg *Arg, reply *string) error {

	idRisorsa := arg.Id //id oggetto
	idPredecessor := sha_consistent(node.Predecessor)

	if (idRisorsa <= node.Id && idRisorsa > idPredecessor) || (idPredecessor > node.Id && (idRisorsa > idPredecessor || idRisorsa <= node.Id)) {
		if node.Objects[idRisorsa] == "" { //se non c'è
			*reply = "L'oggetto cercato non è presente.\n"
		} else {
			if arg.Type { //se arg.Type == true, allora la ricerca l'ho fatta per rimuovere l'oggetto dal nodo.
				*reply = fmt.Sprintf("L'oggetto con id '%d' e valore '%s' è stato rimosso.\n", idRisorsa, node.Objects[idRisorsa])
				delete(node.Objects, idRisorsa)
				fmt.Printf("Nodo: %d, Objects: %v\n", node.Id, node.Objects)

			} else {
				*reply = fmt.Sprintf("L'oggetto con id '%d' e valore '%s' è posseduto dal nodo '%d'.\n", idRisorsa, node.Objects[idRisorsa], node.Id)

			}
		}
		return nil
	}
	//se procedo, vuol dire che il nodo in questione non gestisce la risorsa. Proviamo a chiedere al predecessore.
	client, err := rpc.DialHTTP("tcp", node.Predecessor)
	if err != nil {
		log.Fatal("Errore nodo SearchObject: non riesco a contattare il predecessore", err)
	}
	arg.PredOp = "searchOrRemove"
	resultPredecessor := ""

	err = client.Call("OtherNode.AskPredecessor", arg, &resultPredecessor)
	if err != nil {
		log.Fatal("Errore nodo SearchObject: non riesco a chiamare OtherNode.AskPredecessor ", err)
	}

	client.Close()

	if resultPredecessor != "" {
		*reply = resultPredecessor

		return nil

	} else {
		nodoContact := InterrogateFinger(arg, reply)
		client, err := rpc.DialHTTP("tcp", nodoContact)
		if err != nil {
			log.Fatal("Errore nodo SearchObject: non riesco a contattare il nodo trovato sulla FT  ", err)
		}
		err = client.Call("OtherNode.SearchObject", arg, &reply)
		if err != nil {
			log.Fatal("Errore nodo SearchObject: non riesco a chiamare OtherNode.SearchObject ", err)
		}

		client.Close()

	}
	return nil
}

/*
*************** Produzione della Finger Tabble ************************************************************************************
L'idea alla base è la seguente:
quando calcolo la riga i-esima = id + 2^(i-1) mod (2^m), mi chiedo se tale identificativo è gestito dal mio nodo successore.
Se sì, nella mia FT aggiungo l'identificativo del nodo successore.
Altrimenti, passo il valore al mio nodo successore, e lui confronterà col suo nodo successore se quell'identificativo è di sua competenza.
*/

var m int

func Finger(me *Node) error {
	id := me.Id
	var result int
	var err error
	var idSucc = sha_consistent(me.Successor)

	m, err = ReadFromConfig() //leggo "m" dal json
	if err != nil {
		log.Fatal("Errore nella lettura del file config.json, ", err)
	}
	fingerTable := make([]int, m+1)
	fingerTable[0] = me.Id
	fingerTable[1] = idSucc

	for i := 2; i <= m; i++ {
		// Calcola id + 2^(i-1) mod (2^m)
		val := (id + (1 << (i - 1))) % (1 << m)
		if id == idSucc {
			fingerTable[i] = id
			//esempi per le ultime due condizioni   es: id 29, idSucc 2, val 31             //id: 29, idSucc 2, val 1
		} else if ((val > id) && (val <= idSucc)) || ((val > id) && (id > idSucc)) || ((id > idSucc) && (val < id) && (val <= idSucc)) {

			fingerTable[i] = idSucc
		} else {
			client, err := rpc.DialHTTP("tcp", me.Successor)

			if err != nil {
				result = ContactRegistryAliveNode(me.Successor, idSucc)

				if result == 0 {
					return nil //se result == 0, allora il nodo è caduto. quindi interrompo la comunicazione. Il registry cancellerà il nodo dalla lista nodi. Dopo riaggiorno la FT.
				} else {
					log.Fatal("Errore nodo Finger: il nodo da contattare è attivo, ma non riesco a instaurare una connessione.", err)
				}
			}

			err = client.Call("OtherNode.FindSuccessor", val, &result)
			if err != nil {
				log.Fatal("Errore nodo Finger: non riesco a chiamare OtherNode.FindSuccessor ", err)
			}
			fingerTable[i] = result
			client.Close()
		}

	}
	me.Finger = fingerTable
	PrintFingerTable(me)
	return nil
}

func (t *OtherNode) FindSuccessor(val int, reply *int) error {
	var result int

	var idSucc = sha_consistent(node.Successor)
	if ((val > node.Id) && (val <= idSucc)) || ((val > node.Id) && (node.Id > idSucc)) || ((node.Id > idSucc) && (val < node.Id) && (val <= idSucc)) {

		*reply = idSucc

	} else {

		client, err := rpc.DialHTTP("tcp", node.Successor)
		if err != nil {
			result = ContactRegistryAliveNode(node.Successor, idSucc)
			if result == 0 { //se result == 0, allora il nodo è caduto. quindi interrompo la comunicazione. Il registry cancellerà il nodo dalla lista nodi. Dopo riaggiorno la FT.
				return nil
			} else {
				log.Fatal("Errore nodo Finger: il nodo da contattare è attivo, ma non riesco a instaurare una connessione.", err)
			}
		}
		err = client.Call("OtherNode.FindSuccessor", val, &reply)
		if err != nil {
			log.Fatal("Errore nodo FindSuccesor: non riesco a chiamare OtherNode.FindSuccessor ", err)
		}
		client.Close()

	}

	return nil
}

func scanRing(me *Node, stopChan <-chan struct{}) {
	//time.Sleep(5 * time.Second)
	neightbors := getNeighbors(me.Ip)
	me.Successor = neightbors.Successor
	me.Predecessor = neightbors.Predecessor
	me.Objects = getKeys(me)
	if len(me.Objects) != 0 {
		fmt.Println(me.Objects)
	}

	for {
		select {
		case <-stopChan:
			fmt.Printf("Connessione interrotta correttamente.\n")
			os.Exit(0)
		default:
			time.Sleep(10 * time.Second)
			Finger(me)

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
	me.Id = sha_consistent(me.Ip)

	fmt.Printf("Io sono %d, Indirizzo IP:%s\n", me.Id, ipPortString)

	othernode := new(OtherNode)
	rpc.Register(othernode)
	rpc.HandleHTTP()

	go scanRing(me, stopChan)

	listener, err := net.Listen("tcp", ipPortString)
	if err != nil {
		log.Fatal("Listener error in node.go :", err)
	}

	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatal("Errore nel server HTTP: ", err)
	}

}
