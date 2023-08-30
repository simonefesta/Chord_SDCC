package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"time"
)

func sha_adapted(key string) int {

	m, err := ReadFromConfig() //leggo "m" dal json
	if err != nil {
		fmt.Println("Error reading m:", err)
		return -1
	}
	N := (1 << m)
	//l'hash deve essere nell'intorno (0, risorse rete)
	h := sha1.New()         //oggetto hash
	h.Write([]byte(key))    //scrivo key in byte in h
	hashedKey := h.Sum(nil) //calcolo hash
	res := byte(N - 1)      //scrivo in byte N-1
	for i := 0; i < len(hashedKey); i++ {
		res = res ^ (hashedKey[i] % byte(N)) // applico la riduzione modulo N
		//la parte tra parentesi ritorna valore tra 0 e N-1
		//res ^ (...) fa uno XOR  per ogni bit, se uno dei due bit corrispondenti è impostato, il bit nel risultato sarà impostato; altrimenti, il bit nel risultato sarà azzerato.
	}
	return int(res)
}

// test per ip e porta
func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

type Config struct {
	M int `json:"bits"`
}

func ReadFromConfig() (int, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return 0, err
	}
	filePath := filepath.Join(currentDir, "config.json")

	file, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return 0, err
	}

	return config.M, nil
}

func PrintFingerTable(me *Node) {

	fmt.Printf("FT[%d]: ", me.Id)
	for i := 1; i <= len(me.Finger)-1; i++ {
		fmt.Printf("<%d,%d> ", i, me.Finger[i])
	}
	fmt.Printf("\n")
	time.Sleep(15 * time.Second)

}

// se il nodo non risponde, chiedo al registry di verificare se sia attivo
func ContactRegistryAliveNode(nodetocheck string, idnodetocheck int) int {
	var reply int
	register, err := rpc.DialHTTP("tcp", RegistryFromInside)
	if err != nil {
		log.Fatal("Errore nodo Finger: non riesco a contattare il registry dall'interno per vedere se un nodo è vivo ", err)
	}
	arg := new(Arg)
	arg.Value = nodetocheck
	arg.Id = idnodetocheck
	err = register.Call("Registry.IsNodeAlive", arg, &reply)
	if err != nil {
		log.Fatal("Errore nodo Finger: non riesco a chiamare Registry.isNodeAlive ", err)
		return -1

	}
	register.Close()

	return 0
}
