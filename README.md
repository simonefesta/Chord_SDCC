# Implementazione dell'algoritmo/protocollo di Chord
- Corso: Sistemi Distribuiti e Cloud Computing
- Facoltà di Ingegneria Informatica Magistrale, Università di Roma Tor Vergata.
- Ambiente di sviluppo: Linux, in particolare <i>Fedora Linux 38</i>. Testato con  <i>Docker Desktop 4.22.0</i> e  <i>GO version go1.20.6</i>.
  
# Breve Descrizione
Il seguente progetto riproduce una overlay network strutturata basata sul protocollo di Chord. Mediante un client, è possibile, passando per un server registry, memorizzare/ricercare/eliminare stringhe sui nodi componenti l'anello. La rimozione  <i>controllata </i> di un nodo è supportata.
![ring](https://github.com/simonefesta/Chord_SDCC/assets/55951548/04af223b-d756-4e77-b3b5-c74ed7ffe8d4)



# Esecuzione del programma
Il programa richiede l'avvio in background del Docker Server. 
Successivamente, recandosi nella directory principale del progetto, da terminale, eseguire:
```
docker-compose build
```
e in seguito
```
docker-compose up
```
Questo permette la creazione di un numero definito di container e di un server registry nella stessa rete.
E' possibile variare il numero di nodi da creare. Per far questo bisognerà modificare il file

``` config.json``` e, sempre dalla directory principale, ricompilare il file docker-compose.yml tramite il comando, da terminale:
``` python generate-compose.py ```
Questa operazione richiederà tuttavia una nuova esecuzione dei comandi build e run visti sopra.

Sempre partendo dalla directory principale, è possibile avviare il client recandosi nella cartella <i>client</i> mediante comando
```cd client``` e procedere all'avvio mediante ```go run client.go``` per avere un'interfaccia per la gestione dell'anello.

E' supportata l'aggiunta postuma di un nodo.
Dalla directory principale, mediante terminale, basterà eseguire il comando ``` ./start_node.sh ```.
E' possibile aggiungere ulteriori nodi, uno alla volta, tramite stesso comando, ma, per permettere una corretta gestione delle porte, bisognerà specificare un flag crescente, che parte da 1.
Ad esempio, ``` ./start_node.sh ``` istanzia il primo nodo postumo, ``` ./start_node.sh 1``` istanzia il secondo nodo postumo, ``` ./start_node.sh 2``` istanzia il terzo nodo postumo.
