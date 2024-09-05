import socket
import logging
from typing import Dict, List, Tuple
import common
from common.sigterm_binding import SigTermSignalBinder, SigTermError
import common.sigterm_binding
from common.utils import Bet, store_bets, store_bets_for_lottery
from common.central_lottery_agency import CentralLotteryAgency
from common.client_message import ClientMessageDecoder
from multiprocessing import Manager, Process, Pool, Lock, Queue
from time import sleep

BET_LEN_SIZE = 1
MSG_LEN_SIZE = 2
FIRST_FIELD_SIZE = 2
MAX_AGENCIES = 5


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
        self.bet_writing_queue = Queue(1)
        self.agency_socket_queue = Queue(MAX_AGENCIES)
        self.worker_queue = Queue(MAX_AGENCIES)
        self.lock = Lock()

    """
    Removes the agency socket from the dictionary of agency sockets and
    returns it.
    """
    def delete_agency_socket(self, agency_id):
        self._agency_sockets.pop(agency_id)

    """
    Receives the agency id and the agency socket and stores it in the
    dictionary of agency sockets, with the key being the agency_id and
    the vaklue being the agency_socket.
    """
    def store_agency_socket(self, agency_id, agency_socket):
        self._agency_sockets[agency_id] = agency_socket

    """
    Fires a process in the process pool that sends the winners to the
    agency. The process receives the winners list, the agency socket and
    the agency id as arguments, and closes the corresponding socket.
    """
    def _announce_winners_to_agency(self, agency_socket, agency_id, winners: List[int]):
        self.safe_send(agency_socket, self.create_winners_message(winners))
        logging.info(f'action: winners_announced | result: success | agency: {agency_id} | winners: {len(winners)}')
        agency_socket.close()

    """
    Iterates through the clients that are waiting for the lottery and
    sends them the winners of the lottery. The winners are passed by parameter.
    """
    def announce_winners_to_agencies(self, winners: Dict[int, List[int]]):
        logging.info(self.agency_socket_queue.empty())
        while not self.agency_socket_queue.empty():
            res = self.agency_socket_queue.get()
            agency_id, agency_socket = res

            self.worker_queue.put((agency_socket, agency_id, winners.get(agency_id, [])))
    """
    Starts the lottery and announces the winners to the agencies. The
    function blocks until the writing process finishes writing the bets
    to the file, so all bets are considered for the lottery
    """
    def _start_lottery(self):
        # self.end_threadpool_workers()

        logging.info(f'action: lottery_time | result: bets_received')
        self.central_lottery_agency.determine_winners()
        
        logging.info(f'action: lottery_time | result: winners_determined')
        winners = self.central_lottery_agency.get_winners()
        logging.info(f"winners: {winners}")
        self.announce_winners_to_agencies(winners)

    """
    Waits for all processes to finish and closes eventual open sockets
    """
    def finish_gracefully(self):
        self._server_socket.close()
        self.end_threadpool_workers()

        # Pool is closed, so the sockets that were waiting for the lottery
        # and somehow didn't receive the winners will be closed here
        # This can happen when SIGTERM is raised when the server is waiting
        # for the lottery to start. Also the lock isn't needed anymore
        # because the pool has already been joined
        # for agency_socket in self._agency_sockets.values():
        #     agency_socket.close()

    def _init_thread_pool(self):
        self.workers = []
        for i in range(MAX_AGENCIES):
            p = Process(target=self.__handle_client_connection, args=(self.lock,))
            self.workers.append(p)
            p.start()

    def _join_workers(self):
        for p in self.workers:
            p.join()

        self.workers = []

    def end_threadpool_workers(self):
        for _ in range(len(self.workers)):
            self.worker_queue.put(None)
        
        self._join_workers()
    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        self._init_thread_pool()
        sigterm_binder = SigTermSignalBinder()
        client_sock = None

        with Manager() as manager:
            # self.agency_socket_queue = manager.Queue(MAX_AGENCIES)
            # lock = manager.Lock()
            while True:
                try:
                    client_sock = self.__accept_new_connection()
                    # self.pool.apply(self.__handle_client_connection, args=(client_sock, queue,))
                    self.worker_queue.put([client_sock])
                    logging.info(f'action: client_connection | result: success')

                except common.sigterm_binding.SigTermError:
                    logging.info(f'action: SIGTERM received | result: finishing early')
                    break # It's not needed here because the signal triggers the sigterm_received flag, but it is more explicit this way.
                except socket.timeout:
                    # This case is assumed to be the lottery time
                    self._start_lottery()
                    break
                except Exception as e:
                    logging.error(f'action: finishing | result: fail | message: unknown error: {e.with_traceback()}')
                    break
            
            # In multiprocessing each process should close its own client socket before finishing
            self.finish_gracefully()

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

    def create_winners_message(self, winners: List[int] ) -> bytearray:
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
        msg = int(FIRST_FIELD_SIZE).to_bytes(2, 'little') + int(quantity).to_bytes(2, 'little')
        return msg

    def __handle_client_connection(self, lock):#, process_set, process_set_lock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        while True:
            client_sock = self.worker_queue.get()
            if client_sock is None:
                logging.info(f'action: client_connection | result: finishing')
                return # This is the signal to finish the process
            
            elif len(client_sock) > 1:
                client_sock, agency,  winners = client_sock
                self._announce_winners_to_agency(client_sock, agency, winners)
                continue

            client_message = None
            client_sock = client_sock[0]
            try:
                client_data = self.safe_receive(client_sock)
                client_message = ClientMessageDecoder.decode_client_message(client_data)

                if not client_message.waiting_for_lottery:
                    bets = client_message.bets                
                    with lock:
                        store_bets(bets)

                    logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
                    self.safe_send(client_sock, self.create_bets_answer(len(bets)))
                    
                else:
                    agency = client_message.client_agency
                    self.agency_socket_queue.put((agency, client_sock))
                    logging.info(f'action: agencia_esperando_sorteo | result: success | agencia: ${agency}')

            except OSError as e:
                logging.error(f"action: receive_message | result: fail | error: {e}")
            except ReadingMessageError as e:
                logging.error(f'action: apuesta_recibida | result: fail | cantidad: 0')
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
