# Script per generare il file docker-compose.yml dinamicamente

import sys
import json



def get_m_from_config():
    with open('config.json') as config_file:
        config_data = json.load(config_file)
        return config_data['bits']

def generate_docker_compose(num_containers):
    # Apri il file docker-compose.yml in modalità scrittura
    with open('docker-compose.yml', 'w') as f:
        f.write(f'version: "3"\n\n')
        f.write('services:\n')

        # Configurazione del servizio registry
        f.write('  registry:\n')
        f.write('    build:\n')
        f.write('      context: ./  #Parto dalla root del progetto\n')
        f.write('      dockerfile: ./DockerFiles/registry/Dockerfile  #path del docker file associato al registry\n')
        f.write('    ports:\n')
        f.write('      - "1234:1234"  # Mappa la porta 1234 del server del container all\'host\n')
        f.write('    networks:\n')
        f.write('      - my_network\n\n')

        # Crea i servizi node con nomi e porte distinti
        for i in range(1, num_containers + 1):
            f.write(f'  node{i}:\n')
            f.write(f'    build:\n')
            f.write(f'      context: ./ \n')
            f.write(f'      dockerfile: ./DockerFiles/node/Dockerfile  \n')
            f.write(f'    ports:\n')
            f.write(f'      - "{8004 + i}:{8004 + i}"  # Mappa la porta 8005 del server del container all\'host\n')
            f.write(f'    environment:\n')
            f.write(f'      - NODE_PORT={8004 + i}\n')  # mantengo info su porta esposta
            f.write(f'    networks:\n')
            f.write(f'      - my_network\n')
            f.write(f'    depends_on:\n')
            f.write(f'      - registry\n\n')

        f.write('networks:\n')
        f.write('  my_network:\n')
        f.write('    driver: bridge\n')

if __name__ == "__main__":
    num_containers = get_m_from_config()
    generate_docker_compose(num_containers)
