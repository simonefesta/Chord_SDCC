#!/bin/bash

# Avvia una sessione di tmux
tmux new-session -d -s mysession

# Compila ed esegui il registro nella sessione
tmux send-keys -t mysession "cd registry" C-m
tmux send-keys -t mysession "go build" C-m
tmux send-keys -t mysession "./registry" C-m

# Crea una finestra per ogni nodo
tmux new-window -t mysession:1 -n "Node 1"
tmux send-keys -t mysession:1 "cd ../node" C-m
tmux send-keys -t mysession:1 "go build" C-m
tmux send-keys -t mysession:1 "./node 127.0.0.2:8082" C-m

tmux new-window -t mysession:2 -n "Node 2"
tmux send-keys -t mysession:2 "cd ../node" C-m
tmux send-keys -t mysession:2 "go build" C-m
tmux send-keys -t mysession:2 "./node 127.0.0.3:8083" C-m

tmux new-window -t mysession:3 -n "Node 3"
tmux send-keys -t mysession:3 "cd ../node" C-m
tmux send-keys -t mysession:3 "go build" C-m
tmux send-keys -t mysession:3 "./node 127.0.0.4:8084" C-m

tmux new-window -t mysession:4 -n "Node 4"
tmux send-keys -t mysession:4 "cd ../node" C-m
tmux send-keys -t mysession:4 "go build" C-m
tmux send-keys -t mysession:4 "./node 127.0.0.5:8085" C-m

# Crea una finestra per il client
tmux new-window -t mysession:5 -n "Client"
tmux send-keys -t mysession:5 "cd ../client" C-m
tmux send-keys -t mysession:5 "go build" C-m
tmux send-keys -t mysession:5 "./client" C-m

# Attiva la finestra del registro
tmux select-window -t mysession:0

# Attiva la sessione di tmux
tmux attach-session -t mysession
