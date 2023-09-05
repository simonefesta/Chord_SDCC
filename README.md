# Implementazione dell'algoritmo/protocollo di Chord

- Corso: Sistemi Distribuiti e Cloud Computing
- Facoltà di Ingegneria Informatica Magistrale, Università di Roma Tor Vergata.
- Ambiente di sviluppo: Linux, in particolare <i>Fedora Linux 38</i>. Testato con  <i>Docker Desktop 4.22.0</i> e  <i>GO version go1.20.6</i>.
- Autore 👨‍💻: Simone Festa

# Breve Descrizione

Il seguente progetto riproduce una overlay network strutturata basata sul protocollo di Chord. Mediante un client, è possibile, passando per un server registry, memorizzare/ricercare/eliminare stringhe sui nodi componenti l'anello. La rimozione  <i>controllata </i> di un nodo è supportata totalmente. Viene gestito anche il <i>crash </i>di un nodo,  per consentire al sistema di rimanere consistente, senza però attuare un meccanismo di replicazione per mantenere le risorse del nodo caduto.


![ring](https://github.com/simonefesta/Chord_SDCC/assets/55951548/04af223b-d756-4e77-b3b5-c74ed7ffe8d4)

# Esecuzione del programma

### Start-up del sistema

Il programa richiede l'avvio in background del Docker Server.  
Il progetto prevede un <i>file di configurazione </i>: `config.json`  
In questo file è possibile definire due parametri:

- <i>bits</i>: numero di bits usati per identificare un nodo. Permette di creare fino a $2^{bits}$ nodi.
  (se $bits=5$, è possibile creare fino a 32 nodi.)

- <i>nodes</i>: numero di nodi da creare all'avvio del sistema. Devono essere $\leq2^{bits}$.  

<i>Nel caso si modifichino </i>questi parametri, per applicarli al progetto è necessario digitare da terminale, immettendosi nella **directory principale del progetto** (Chord_SDCC): 

```
python generate-compose.py
```

per poter aggiornare il file `docker-compose.yml`

Dalla stessa directory, eseguire il <i>build</i> con:

```
docker-compose build
```

e l'avvio con: 

```
docker-compose up
```

Questo permette la creazione di un numero definito di container e di un server registry nella stessa rete.  


### Avvio del client

Sempre partendo dalla directory principale, è possibile avviare il client recandosi nella cartella <i>client</i> mediante comando 

```
cd client
```

e procedere all'avvio mediante  

```
go run client.go
```

per avere un'interfaccia per la gestione dell'anello.  

### Avvio di un nodo dopo lo start-up

E' supportata l'aggiunta postuma di un nodo singolarmente.  
Dalla directory principale, mediante terminale, basterà eseguire il comando:

```
./start_node.sh
```


E' possibile aggiungere ulteriori nodi, uno alla volta, tramite stesso comando, ma, per permettere una corretta gestione delle porte, bisognerà specificare un flag crescente, che parte da 1.
Ad esempio:

- ``` ./start_node.sh``` istanzia il primo nodo postumo, 
- ``` ./start_node.sh 1``` istanzia il secondo nodo postumo,
- ``` ./start_node.sh 2``` istanzia il terzo nodo postumo.

# Esecuzione del programma con AWS - EC2

<b>Disclaimer</b>: l'istanza richiede l'installazione di Docker, Golang e Python. Per accedere ai servizi Amazon EC2 sono richieste delle credenziali per AWS. La guida fa uso delle credenziali memorizzate nel file ```AWSKeypair.pem``` poste nella cartella ```.ssh```.

Recarsi sul sito: https://awsacademy.instructure.com/courses/28710/modules/items/2385832  
Effettuare il login se necessario.  
Premere 'Start Lab'.  
Quando la scritta 'AWS' presenta un cerchio verde, premere 'AWS' per entrare in 'AWS Management Console'.  
Recarsi nel pannello di controllo EC2.  
Avviare una nuova istanza, o mantenerne una precedentemente creata. (<i>nome di esempio: ec2).</i>  
Da terminale (presso qualsiasi cartella), collegarsi all'istanza mediante:  
``` ssh -i ~/.ssh/AWSKeypair.pem ec2-user@<ipv4_public_address>```    
(nb: l'indirizzo ipv4 pubblico è visibile nel pannello delle istanze di Amazon EC2).  
Per copiare il progetto, recarsi nella directory <i>genitore</i> della directory contenente il progetto. (Se il progetto è nella folder 'Scaricati', posizionarsi in 'Scaricati').    
Da terminale:    
```scp -i ~/.ssh/AWSKeypair.pem -r Chord_SDCC ec2-user@<ipv4_public_address>:/home/ec2-user/```  
Quando il progetto è stato copiato, è possibile avviarlo con i comandi già visti:

- ```"docker-compose build"```,  
- ```"docker-compose up"```.  
  Recandosi nella cartella <i>client</i>, si deve avviare il client come già visto in precedenza.  
  Terminati i test, puliamo l'ambiente dei container tramite il comando:  
  ```docker container prune```  
  Per scollegarsi dall'istanza, usare il comando ```exit```

# NOTE

- E' possibile incorrere in collisioni durante la creazione dei container. Un nodo mappato su in identificativo già usato verrà chiuso. Il sistema funzionerà normalmente.
  Se ciò si verifica durante l'avvio del nodo singolo, mediante lo stesso comando di avvio l'istanza verrà riavviata ed entrerà nel sistema (a meno di un'altra, ma più improbabile, collisione).
