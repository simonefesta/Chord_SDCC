# Script per generare il file docker-compose.yml dinamicamente

def generate_docker_compose(num_containers):
    # Apri il file docker-compose.yml in modalità scrittura
    with open('docker-compose.yml', 'w') as f:
        f.write(f'version: "3"\n\n')
        f.write('services:\n')

        # Configurazione del servizio registry
        f.write('  registry:\n')
        f.write('    build:\n')
        f.write('      context: ./  # Percorso corretto per il contesto del Dockerfile del server\n')
        f.write('      dockerfile: ./DockerFiles/registry/Dockerfile  # Percorso corretto per il Dockerfile del server\n')
        f.write('    ports:\n')
        f.write('      - "1234:1234"  # Mappa la porta 1234 del server del container all\'host\n')
        f.write('    networks:\n')
        f.write('      - my_network\n\n')

        # Crea i servizi node con nomi e porte distinti
        for i in range(1, num_containers + 1):
            f.write(f'  node{i}:\n')
            f.write(f'    build:\n')
            f.write(f'      context: ./  # Percorso corretto per il contesto del Dockerfile del server\n')
            f.write(f'      dockerfile: ./DockerFiles/node/Dockerfile  # Percorto corretto per il Dockerfile del server\n')
            f.write(f'    ports:\n')
            f.write(f'      - "{8004 + i}:8005"  # Mappa la porta 8005 del server del container all\'host\n')
            f.write(f'    networks:\n')
            f.write(f'      - my_network\n\n')

        f.write('networks:\n')
        f.write('  my_network:\n')
        f.write('    driver: bridge\n')

if __name__ == "__main__":
    import sys

    if len(sys.argv) != 2:
        print("Usage: python generate_compose.py <num_containers>")
        sys.exit(1)

    try:
        num_containers = int(sys.argv[1])
    except ValueError:
        print("Invalid input for num_containers. Please provide a valid integer.")
        sys.exit(1)

    generate_docker_compose(num_containers)
