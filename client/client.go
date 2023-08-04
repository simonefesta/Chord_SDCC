package main

import (
	"bufio"
	"fmt"
	"log"
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

	fmt.Println("1. Inserisci un nuovo oggetto")
	fmt.Println("2. Cerca un oggetto")
	/* TODO
	fmt.Println("3. Aggiungi un nuovo nodo")
	fmt.Println("4. Rimuovi un nodo") */

	for {
		fmt.Print("Inserisci scelta: ")
		fmt.Scanln(&resp) //acquisisco la risposta da tastiera
		switch int(resp) {
		case 1:
			fmt.Print("Oggetto da inserire: ")
			scanner := bufio.NewScanner(os.Stdin) //meglio di scanln, se devo acquisire frasi o altro.
			scanner.Scan()                        //leggo ciò che è stato inserito
			keyboardArgoment.Value = scanner.Text()

			//devo connettermi per inserire questo oggetto
			client, err := rpc.DialHTTP("tcp", "localhost:1234")
			if err != nil {
				log.Fatal("Errore connessione client ", err)
			}
			/*Nell'esempio sopra, DialHTTP viene utilizzato per creare una connessione RPC utilizzando il protocollo TCP e l'indirizzo "localhost:1234" come server RPC di destinazione */
			/*Una volta stabilita la connessione, puoi utilizzare l'oggetto client per chiamare i metodi esposti dal server RPC utilizzando client.Call o altre funzioni di rpc.Client. */

			err = client.Call("Registry.ReturnChordNode", keyboardArgoment, &result) //chiamo metodo, passando come argomento "keyboardArgoment" ed ottengo "result", che è il nodo scelto random.
			if err != nil {
				// Gestisci l'errore se si verifica
				log.Fatal("Errore nella chiamata di metodo RPC: ", err)
			}

			//fmt.Println("Registry return chord node ha restituito nodo RANDOM ", result)

			//mi riconnetto per chiedere un altro metodo (posso farlo una volta sola?)

			client, err = rpc.DialHTTP("tcp", result) //contatto il nodo che ho trovato prima.
			if err != nil {
				log.Fatal("Errore connessione client ", err)
			}

			//fmt.Println("Ho trovato ", result)

			err = client.Call("Successor.AddObject", keyboardArgoment, &result) //chiamo metodo, passando come argomento "keyboardArgoment" ed ottengo "result"
			if err != nil {
				// Gestisci l'errore se si verifica
				log.Fatal("Errore nella chiamata di metodo RPC: ", err)
			}

			fmt.Println(result)

		case 2:
			fmt.Print("Digita l'id dell'oggetto da cercare: ")

			fmt.Scanln(&keyboardArgoment.Id)

			client, err := rpc.DialHTTP("tcp", "localhost:1234")
			if err != nil {
				log.Fatal("Client connection error: ", err)
			}

			err = client.Call("Registry.ReturnChordNode", keyboardArgoment, &result) //chiamo un nodo random
			if err != nil {
				log.Fatal("Client invocation error: ", err)
			}

			client, err = rpc.DialHTTP("tcp", result) //mi connetto a questo nodo random
			if err != nil {
				log.Fatal("Client connection error2: ", err)
			}

			err = client.Call("Successor.SearchObject", keyboardArgoment, &result)
			if err != nil {
				log.Fatal("Client invocation error: ", err)
			}

			fmt.Println(result)

		default:
			println("Devi selezionare una delle due scelte digitando 1 o 2")
		}
	}
}
