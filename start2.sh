#!/bin/bash

# Compila ed esegui il registro
cd registry
go build
./registry &

# Compila ed esegui i nodi
cd ../node
go build
./node 127.0.0.2:8082 &
sleep 2
./node 127.0.0.3:8083 &
sleep 2
./node 127.0.0.4:8084 &
sleep 2
./node 127.0.0.5:8085 &

# Compila ed esegui il client
cd ../client
go build
./client

# Torna alla cartella principale
cd ..

