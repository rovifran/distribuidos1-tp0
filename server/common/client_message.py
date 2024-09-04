from common.utils import Bet

BET_LEN_SIZE = 1

class ClientMessageDecoder:
    def __init__(self):
        self.bets = []
        self.waiting_for_lottery = False
        self.client_agency = None

    @classmethod
    def decode_client_message(self, data: bytearray) -> 'Self':
        """
        Receives a bytearray object and returns a ClientMessageDecoder
        instance containing the information in the bytes object.

        The decoding is done following this rules: 
        - If the amount of bytes is 1, the client is waiting for the lottery results
        - If the amount of bytes is greater than 1, the client is sending bets
        
        It is assumed that a message can't be empty.
        """ 
        
        if len(data) == 1:
            return ClientMessageDecoder._decode_waiting_for_lottery(data)
        
        return ClientMessageDecoder._decode_sending_bets(data)
    
    @classmethod
    def _decode_waiting_for_lottery(self, data: bytearray) -> 'Self':
        """
        Receives a bytearray object with length 1 and returns a ClientMessageDecoder
        instance with the waiting_for_lottery attribute set to True.
        """
        # There is no need to decode data, as it is only one byte and it is the agency
        cmd = ClientMessageDecoder()
        cmd.client_agency = int(data[0])
        cmd.set_waiting_for_lottery(True)
        return cmd
    
    @classmethod
    def _decode_sending_bets(self, data: bytearray) -> 'Self':
        """
        Receives a bytearray object with length greater than 1 and
        returns a ClientMessageDecoder instance with the bets attribute
        set to the bets in the bytearray.
        """
        bets = []

        total_bets_len = len(data)
        total_read = 0

        while total_read < total_bets_len:
            bet_len = int.from_bytes(data[total_read : total_read + BET_LEN_SIZE], 'little')
            total_read += BET_LEN_SIZE
            bet = Bet.decodeBytes(data[total_read : total_read + bet_len])
            bets.append(bet)
            total_read += bet_len

        cmd = ClientMessageDecoder()
        cmd.set_bets(bets)
        return cmd
    
    def set_bets(self, bets: list):
        self.bets = bets

    def set_waiting_for_lottery(self, waiting: bool):
        self.waiting_for_lottery = waiting