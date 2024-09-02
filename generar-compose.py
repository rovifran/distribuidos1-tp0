import sys

DEST_FILE = "docker-compose-dev.yaml"
BEGINNING_STR = """name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    networks:
      - testing_net
    volumes:
      - type: bind
        source: ./server/config.ini
        target: /config.ini

"""

CLIENT_STR = """  client%AMOUNT%:
    container_name: client%AMOUNT%
    image: client:latest
    entrypoint: /client
    networks:
      - testing_net
    depends_on:
      - server
    volumes:
      - type: bind
        source: ./client/config.yaml
        target: /config.yaml

"""

NET_STR = """networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""

def parseClientAmountInput(amount: str) -> int:
    """
    Takes the amount of clients as input and returns it as an integer.
    If the amount of clients is not a number or is less than 1, raises a ValueError.
    """

    if not amount.isnumeric() or int(amount) < 1:
        raise ValueError("Amount entered not a number or invalid amount, program closing.")
    return int(amount)

def writeDockerComposeFile(clients_amount, dest_file):
    """
    Receives the name of the file and the amount of clients to write in the file.
    Writes the docker-compose file with the given amount of clients.
    If the file is not found, raises a FileNotFoundError.
    """

    try:
      with open(dest_file, 'w') as f:
          f.write(BEGINNING_STR)
          for i in range(0, clients_amount):
              f.write(CLIENT_STR.replace('%AMOUNT%', str(i + 1)))
          f.write(NET_STR)
    except FileNotFoundError as e:
        raise FileNotFoundError(f"File {dest_file} not found, program closing.") from e 

def main():
    if len(sys.argv) != 3:
        print("Usage: python generar-compose.py <file_name> <clients_amount>")
        sys.exit(1)
    clients_amount = parseClientAmountInput(sys.argv[2])
    writeDockerComposeFile(clients_amount, sys.argv[1])

if __name__ == "__main__":
    main()