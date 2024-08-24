import sys

DEST_FILE = "docker-compose-dev.yaml"
BEGINNING_STR = """name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net

"""

CLIENT_STR = """  client%AMOUNT%:
    container_name: client%AMOUNT%
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=%AMOUNT%
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server

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
    clients_amount = parseClientAmountInput(sys.argv[2])
    writeDockerComposeFile(clients_amount, sys.argv[1])

if __name__ == "__main__":
    main()