package main

import "crypto/sha1"

const (
	N int = 32
)

func sha_adapted(key string) int {
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
