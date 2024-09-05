class ReadingMessageError(Exception):
    pass

MSG_LEN_SIZE = 2
def safe_receive(client_sock) -> bytes:
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

def safe_send(client_sock, data):
    """
    Sends the message passed by parameter to the socket, tolerant to short writes 
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