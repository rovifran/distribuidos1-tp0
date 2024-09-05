from typing import List

FIRST_FIELD_SIZE = 2
SIZE_UINT32 = 4
SIZE_INT16 = 2
def create_winners_message(winners: List[int] ) -> bytearray:
    """
    Creates a bytearray with the length of the winners. This bytearray is
    sent to the agency to acknowledge the winners of the lottery.
    """
    winners_bytes = b''
    for winner in winners:
        winners_bytes += int(winner).to_bytes(SIZE_UINT32, 'little')
    
    winners_bytes = int(len(winners)).to_bytes(SIZE_INT16, 'little') + winners_bytes
    
    return int(len(winners_bytes)).to_bytes(SIZE_INT16, 'little') + winners_bytes

def create_bets_answer(quantity: int) -> bytearray:
    """
    Creates a bytearray with the length of the bets. This bytearray is
    sent to the client to acknowledge the bets received.
    """
    msg = int(FIRST_FIELD_SIZE).to_bytes(2, 'little') + int(quantity).to_bytes(2, 'little')
    return msg