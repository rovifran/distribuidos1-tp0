import socket
import logging
from typing import Dict, List
from common.graceful_finisher import GracefulFinisher, SigTermError
from common.utils import Bet
from common.central_lottery_agency import CentralLotteryAgency
from common.client_message import ClientMessageDecoder

BET_LEN_SIZE = 1
MSG_LEN_SIZE = 2
FIRST_FIELD_SIZE = 2

class ReadingMessageError(Exception):
    pass

class Server:
    """
    Initializes the server socket. The server socket is a TCP socket
    with a timeout associated that serves as a lottery winner syncronizer.
    """
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._agency_sockets = {}
        self._server_socket.settimeout(10)
        self._server_socket.listen(listen_backlog)
        self.central_lottery_agency = CentralLotteryAgency()

    def store_agency_socket(self, agency_id, agency_socket):
        self._agency_sockets[agency_id] = agency_socket

    def announce_winners_to_agencies(self, winners: Dict[int, List[int]]):
        for agency_id, winners in winners.items():
            agency_socket = self._agency_sockets.get(agency_id)
            if agency_socket:
                self.safe_send(agency_socket, self.create_bets_answer(len(winners)))
                logging.info(f'action: winners_announced | result: success | agency: ${agency_id} | winners: ${len(winners)}')
                agency_socket.close()

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        graceful_finisher = GracefulFinisher()
        client_sock = None

        while not graceful_finisher.finished:
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except SigTermError:
                logging.info(f'action: SIGTERM received | result: finishing early')
            except socket.timeout:
                # This case is assumed to be the lottery time
                self.central_lottery_agency.determine_winners()
                logging.info(f'action: lottery_time | result: winners_determined')
                self.central_lottery_agency.get_winners()
                winners = self.central_lottery_agency.get_winners()
                self.announce_winners_to_agencies(winners)
                return
                
            finally:
                pass
                # if client_sock and not self.:
                #     client_sock.close()

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
                raise ReadingMessageError("Error reading from client")

        msg_len_bytes = _receive_all(MSG_LEN_SIZE)
        msg_len = int.from_bytes(msg_len_bytes, 'little')

        return _receive_all(msg_len)

    def safe_send(self, client_sock, data):
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

        _send_all(data)

    def create_winners_message(self, agency: int, winners: List[int]) -> bytearray:
        """
        Creates a bytearray with the length of the winners. This bytearray is
        sent to the agency to acknowledge the winners of the lottery.
        """
        winners_bytes = b''
        for winner in winners:
            winners_bytes += int(winner).to_bytes(4, 'little')
        
        winners_bytes = int(len(winners)).to_bytes(2, 'little') + winners_bytes

        return int(len(winners_bytes)).to_bytes(2, 'little') + winners_bytes

    def create_bets_answer(self, quantity: int) -> bytearray:
        """
        Creates a bytearray with the length of the bets. This bytearray is
        sent to the client to acknowledge the bets received.
        """
        return int(FIRST_FIELD_SIZE).to_bytes(2, 'little') + int(quantity).to_bytes(2, 'little')

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        client_message = None
        try:
            client_data = self.safe_receive(client_sock)
            client_message = ClientMessageDecoder.decode_client_message(client_data)

            if not client_message.waiting_for_lottery:
                bets = client_message.bets
                self.central_lottery_agency.add_bets(bets)
                logging.info(f'action: apuesta_recibida | result: success | cantidad: ${len(bets)}')
                self.safe_send(client_sock, self.create_bets_answer(len(bets)))
                return
                
            else:
                agency = client_message.client_agency
                self.store_agency_socket(agency, client_sock)
                logging.info(f'action: agencia_esperando_sorteo | result: success | agencia: ${agency}')
                return

        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        except ReadingMessageError as e:
            logging.error(f'action: apuesta_recibida | result: fail | cantidad: $0')
            self.safe_send(client_sock, self.create_bets_answer(-1))

        finally:
            if client_message and not client_message.waiting_for_lottery:
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
