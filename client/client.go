package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
)

type KeyboardArgoment struct { //ciò che viene preso da tastiera se vogliamo inserire un oggetto
	Id    int
	Value string
}

func main() {

	keyboardArgoment := new(KeyboardArgoment) //creo nuova istanza di keyboardArgoment
	var result string
	var resp int

	fmt.Println("1. Aggiungi oggetto")
	fmt.Println("2. Cerca un oggetto")
	fmt.Println("3. Rimuovi un nodo")

	for {
		fmt.Print("\nInserisci scelta: ")
		fmt.Scanln(&resp) //acquisisco la risposta da tastiera
		switch int(resp) {
		case 1:
			fmt.Print("Oggetto da inserire: ")
			scanner := bufio.NewScanner(os.Stdin) //meglio di scanln, se devo acquisire frasi o altro.
			scanner.Scan()                        //leggo ciò che è stato inserito
			keyboardArgoment.Value = scanner.Text()

			//devo connettermi per inserire questo oggetto
			client, err := rpc.DialHTTP("tcp", "0.0.0.0:1234")
			if err != nil {
				log.Fatal("Errore connessione client registry", err)
			}
			/*Nell'esempio sopra, DialHTTP viene utilizzato per creare una connessione RPC utilizzando il protocollo TCP e l'indirizzo ":1234" come server RPC di destinazione */
			/*Una volta stabilita la connessione, puoi utilizzare l'oggetto client per chiamare i metodi esposti dal server RPC utilizzando client.Call o altre funzioni di rpc.Client. */

			err = client.Call("Registry.ReturnRandomNode", keyboardArgoment, &result) //chiamo metodo, passando come argomento "keyboardArgoment" ed ottengo "result", che è il nodo scelto random.
			if err != nil {
				// Gestisci l'errore se si verifica
				log.Fatal("Errore nella chiamata di metodo RPC: ", err)
			}

			fmt.Printf("contatto: %s", result)

			client, err = rpc.DialHTTP("tcp", result) //contatto il nodo che ho trovato prima.
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

			err = client.Call("Successor.AddObject", keyboardArgoment, &result) //chiamo metodo, passando come argomento "keyboardArgoment" ed ottengo "result"
			if err != nil {
				// Gestisci l'errore se si verifica
				log.Fatal("Errore nella chiamata di metodo RPC: ", err)
			}

			fmt.Println(result)
			fmt.Printf("\n")

		case 2:
			fmt.Print("Digita l'id dell'oggetto da cercare: ")

			fmt.Scanln(&keyboardArgoment.Id)

			//devo connettermi per inserire questo oggetto
			client, err := rpc.DialHTTP("tcp", "0.0.0.0:1234")
			if err != nil {
				log.Fatal("Errore connessione client registry", err)
			}
			/*Nell'esempio sopra, DialHTTP viene utilizzato per creare una connessione RPC utilizzando il protocollo TCP e l'indirizzo ":1234" come server RPC di destinazione */
			/*Una volta stabilita la connessione, puoi utilizzare l'oggetto client per chiamare i metodi esposti dal server RPC utilizzando client.Call o altre funzioni di rpc.Client. */

			err = client.Call("Registry.ReturnRandomNode", keyboardArgoment, &result) //chiamo metodo, passando come argomento "keyboardArgoment" ed ottengo "result", che è il nodo scelto random.
			if err != nil {
				// Gestisci l'errore se si verifica
				log.Fatal("Errore nella chiamata di metodo RPC: ", err)
			}

			client, err = rpc.DialHTTP("tcp", result) //contatto il nodo che ho trovato prima.
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

			err = client.Call("Successor.SearchObject", keyboardArgoment, &result) //iterativamente parte una ricerca tra i nodi usando le FT per trovare la risorsa.
			if err != nil {
				log.Fatal("Client invocation error: ", err)
			}

			fmt.Println(result)
			fmt.Printf("\n")

		case 3:
			fmt.Print("Digita l'id del nodo da rimuovere: ")

			fmt.Scanln(&keyboardArgoment.Id)

			client, err := rpc.DialHTTP("tcp", "0.0.0.0:1234")
			if err != nil {
				log.Fatal("Client connection error: ", err)
			}

			err = client.Call("Registry.RemoveNode", keyboardArgoment, &result)
			if err != nil {
				log.Fatal("Client invocation error: ", err)
			}

			fmt.Println("Eliminazione completata.")
			fmt.Printf("\n")

		default:
			println("Scegliere una delle quattro opzioni, digitando '1','2','3' o '4'.")
		}
	}
}
