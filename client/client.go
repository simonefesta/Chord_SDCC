package main

import (
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strconv"
)

type KeyboardArgoment struct { //Struct per gli argomenti passati da tastiera.
	Id     int
	Value  string
	Choice int
	Type   bool //true se cerco una risorsa per eliminarla, false se la certo per ottenere il valore associato.
}

var RegistryFromOutside string = "0.0.0.0:1234" //indirizzo per contattare il registry dal client.

func main() {

	keyboardArgoment := new(KeyboardArgoment) //creo nuova istanza di keyboardArgoment
	var result string
	var resp int
	var input string

	fmt.Println("1. Aggiungi oggetto, [PUT]")
	fmt.Println("2. Cerca un oggetto, [GET]")
	fmt.Println("3. Rimuovi un oggetto, [REMOVE]")
	fmt.Println("4. Rimuovi un nodo")

	for {
		fmt.Print("\nInserisci scelta: ")
		fmt.Scanln(&resp) //acquisisco la risposta da tastiera
		switch int(resp) {
		case 1:
			fmt.Print("Oggetto da inserire: ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan() //leggo ciò che è stato inserito
			keyboardArgoment.Value = scanner.Text()
			keyboardArgoment.Choice = 1
			_ = scanner.Text() //pulizia buffer

			client, err := rpc.DialHTTP("tcp", RegistryFromOutside)
			if err != nil {
				log.Fatal("Errore nel client: non riesco a connettermi al registry, ", err)
			}

			err = client.Call("Registry.EnterRing", keyboardArgoment, &result) //chiamo metodo, passando come argomento "keyboardArgoment" ed ottengo "result", che è il nodo scelto random.
			if err != nil {
				log.Fatal("Errore nel client: non riesco a chiamare la funzione 'EnterRing' del registry, ", err)
			}

			client.Close()

			client, err = rpc.DialHTTP("tcp", result) //contatto il nodo che ho trovato prima.
			if err != nil {
				log.Fatal("Errore nel client caso 1: non riesco a contattere il nodo necessario, ", err)
			}

			err = client.Call("OtherNode.AddObject", keyboardArgoment, &result)
			if err != nil {
				// Gestisci l'errore se si verifica
				log.Fatal("Errore nel client caso 1: non riesco a chiamare la funzione 'AddObject', ", err)
			}

			fmt.Println(result)
			client.Close()

		case 2:
			fmt.Print("Digita l'id dell'oggetto da cercare: ")

			fmt.Scanln(&input)
			keyboardArgoment.Choice = 2

			id, err := strconv.Atoi(input)
			if err != nil {
				fmt.Println("Input non valido. Devi inserire un numero intero.")

			} else {
				keyboardArgoment.Id = id

				client, err := rpc.DialHTTP("tcp", RegistryFromOutside)
				if err != nil {
					log.Fatal("Errore nel client: non riesco a connettermi al registry, ", err)
				}

				err = client.Call("Registry.EnterRing", keyboardArgoment, &result)
				if err != nil {
					log.Fatal("Errore nel client: non riesco a chiamare la funzione 'EnterRing' del registry, ", err)
				}
				client.Close()

				client, err = rpc.DialHTTP("tcp", result) //contatto il nodo che ho trovato prima.
				if err != nil {
					log.Fatal("Errore nel client caso 2: non riesco a contattere il nodo necessario, ", err)

				}

				err = client.Call("OtherNode.SearchObject", keyboardArgoment, &result) //iterativamente parte una ricerca tra i nodi usando le FT per trovare la risorsa.
				if err != nil {
					log.Fatal("Errore nel client caso 2: non riesco a chiamare la funzione 'SearchObject' (caso 2), ", err)
				}

				fmt.Println(result)
				client.Close()

			}

		case 3:
			fmt.Print("Digita l'id dell'oggetto da rimuovere: ")

			fmt.Scanln(&input)
			keyboardArgoment.Choice = 3 //per eliminare una chiave, devo cercarla, ma questo già lo faccio nel caso due.

			id, err := strconv.Atoi(input)
			if err != nil {
				fmt.Println("Input non valido. Devi inserire un numero intero.")

			} else {
				keyboardArgoment.Id = id

				client, err := rpc.DialHTTP("tcp", RegistryFromOutside)
				if err != nil {
					log.Fatal("Errore nel client: non riesco a connettermi al registry, ", err)
				}

				err = client.Call("Registry.EnterRing", keyboardArgoment, &result)
				if err != nil {
					log.Fatal("Errore nel client: non riesco a chiamare la funzione 'EnterRing' del registry, ", err)
				}
				client.Close()

				keyboardArgoment.Type = true //true se voglio cercare l'oggetto per rimuoverlo.
				client, err = rpc.DialHTTP("tcp", result)
				if err != nil {
					log.Fatal("Errore nel client caso 3: non riesco a contattere il nodo necessario, ", err)

				}

				err = client.Call("OtherNode.SearchObject", keyboardArgoment, &result) //iterativamente parte una ricerca tra i nodi usando le FT per trovare la risorsa.
				if err != nil {
					log.Fatal("Errore nel client caso 3: non riesco a chiamare la funzione 'SearchObject' (caso 3), ", err)
				}

				fmt.Println(result)
				client.Close()

			}

		case 4:
			fmt.Print("Digita l'id del nodo da rimuovere: ")
			fmt.Scanln(&input)

			id, err := strconv.Atoi(input)
			if err != nil {
				fmt.Println("Input non valido. Devi inserire un numero intero.")

			} else {
				keyboardArgoment.Id = id
				client, err := rpc.DialHTTP("tcp", RegistryFromOutside)
				if err != nil {
					log.Fatal("Errore nel client: non riesco a connettermi al registry, ", err)
				}

				err = client.Call("Registry.RemoveNode", keyboardArgoment, &result)
				if err != nil {
					log.Fatal("Errore nel client: non riesco a chiamare la funzione 'RemoveNode' del registry, ", err)
				}

				fmt.Println(result)
				client.Close()

			}

		default:
			println("Scegliere una delle quattro opzioni, digitando '1','2','3'.")
		}
	}
}
