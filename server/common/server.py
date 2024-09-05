import socket
import logging
from common.sigterm_binding import SigTermSignalBinder, SigTermError
from common.utils import Bet, store_bets

BET_LEN_SIZE = 1
TOTAL_BETS_LEN_SIZE = 2

class ReadingBetsError(Exception):
    pass

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

    def finish_gracefully(self):
        self._server_socket.close()

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        sigterm_binder = SigTermSignalBinder()
        client_sock = None

        while not sigterm_binder.finished:
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except SigTermError:
                logging.info(f'action: program finishing early | result: success')
                break
                
            finally:
                if client_sock != None:
                    client_sock.close()
                    client_sock = None

    def safe_receive(self, client_sock) -> Bet:
        """
        Receives the message from the client socket following the protocol:
        - The first 2 bytes are the length of the message
        - The next N bytes are the message itself
        """

        def _receive_all(size):
            """
            This receive is tolerant to short reads, meaning that it will keep
            reading from the socket until the full message is received
            """
            try:
                msg = b''
                while len(msg) < size:
                    received = client_sock.recv(size - len(msg))
                    if len(received) == 0:
                        raise OSError("Client disconnected")
                    msg += received
                return msg
            except OSError as e:
                raise e
            except:
                raise ReadingBetsError("Error reading from client")
        
        bets = []

        total_bets_len_bytes = _receive_all(TOTAL_BETS_LEN_SIZE)
        total_bets_len = int.from_bytes(total_bets_len_bytes, 'little')

        total_read = 0
        while total_read < total_bets_len:
            bet_len_bytes = _receive_all(BET_LEN_SIZE)
            bet_len = int.from_bytes(bet_len_bytes, 'little')
            bet_bytes = _receive_all(bet_len)
            bet = Bet.decodeBytes(bet_bytes)
            bets.append(bet)
            total_read += bet_len + BET_LEN_SIZE

        return bets

    def safe_send(self, client_sock, msg: int):
        """
        Sends a SUCCESS message to the client socket, tolerant to short writes 
        """

        def _send_all(msg):
            """
            This send is tolerant to short writes, meaning that it will keep
            writing to the socket until the full message is sent
            """
            while len(msg) > 0:
                sent = client_sock.send(msg)
                msg = msg[sent:]

        msg = int(msg).to_bytes(2, 'little')
        _send_all(msg)

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            bets = self.safe_receive(client_sock)
            store_bets(bets)
            logging.info(f'action: apuesta_recibida | result: success | cantidad: ${len(bets)}')
            self.safe_send(client_sock, len(bets))
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        except ReadingBetsError as e:
            logging.error(f'action: apuesta_recibida | result: fail | cantidad: $0')
            self.safe_send(client_sock, -1)

        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
