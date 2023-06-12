package main

import (
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
)

type KeyboardArgoment struct { //ciò che viene preso da tastiera se vogliamo inserire un oggetto
	id    int
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
		fmt.Print("Inserisci scelta")
		fmt.Scanln(&resp) //acquisisco la risposta da tastiera
		switch resp {
		case 1:
			fmt.Print("Oggetto da inserire:")
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
			err = client.Call("Registry.ReturnChordNode", keyboardArgoment, &result) //chiamo metodo, passando come argomento "keyboardArgoment" ed ottengo "result"
			if err != nil {
				// Gestisci l'errore se si verifica
				log.Fatal("Errore nella chiamata di metodo RPC: ", err)
			}

			//mi riconnetto per chiedere un altro metodo (posso farlo una volta sola?)

			client, err = rpc.DialHTTP("tcp", "localhost:1234")
			if err != nil {
				log.Fatal("Errore connessione client ", err)
			}

			err = client.Call("Successor.AddObject", keyboardArgoment, &result) //chiamo metodo, passando come argomento "keyboardArgoment" ed ottengo "result"
			if err != nil {
				// Gestisci l'errore se si verifica
				log.Fatal("Errore nella chiamata di metodo RPC: ", err)
			}

			fmt.Println(result)
			//break

			//case2 TODO
		}

	}
}
