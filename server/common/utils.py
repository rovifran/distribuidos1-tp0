import csv
import datetime
import time
from typing import Tuple


""" Bets storage location. """
STORAGE_FILEPATH = "./bets.csv"
""" Simulated winner number in the lottery contest. """
LOTTERY_WINNER_NUMBER = 7574

BYTES_OF_UINT8 = 1
BYTES_OF_UINT16 = 2
BYTES_OF_UINT32 = 4


""" A lottery bet registry. """
class Bet:
    def __init__(self, agency: str, first_name: str, last_name: str, document: str, birthdate: str, number: str):
        """
        agency must be passed with integer format.
        birthdate must be passed with format: 'YYYY-MM-DD'.
        number must be passed with integer format.
        """
        self.agency = int(agency)
        self.first_name = first_name
        self.last_name = last_name
        self.document = document
        self.birthdate = datetime.date.fromisoformat(birthdate)
        self.number = int(number)

    @classmethod
    def decodeBytes(cls, bytes: bytes) -> 'Bet':
        """
        Receives a bytes object and returns a Bet instance containing 
        the information in the bytes object.

        The decoding is done following the format that the client uses
        to encode the information: for the string fields, the
        length of the string is the first byte of the field. The numeric
        fields are encoded as 4 bytes and 2 bytes for the dni and number
        fields respectively.
        """
        pos = 0
        agency, pos = _decodeNumericField(bytes, pos, BYTES_OF_UINT8)
        name, pos = _decodeStringField(bytes, pos)
        surname, pos = _decodeStringField(bytes, pos)
        dni, pos = _decodeNumericField(bytes, pos, BYTES_OF_UINT32)
        birthday, pos = _decodeStringField(bytes, pos)
        number, pos = _decodeNumericField(bytes, pos, BYTES_OF_UINT16)

        return Bet(agency, name, surname, dni, birthday, number)

def _decodeNumericField(bytes: bytes, pos: int, bytes_len: int) -> Tuple[int, int]:
    """
    Receives a bytes object, a position in the bytes object and the
    number of bytes that the numeric field occupies. Returns a tuple
    with the numeric field and the next position in the bytes object.
    """
    return int.from_bytes(bytes[pos : pos + bytes_len], 'little'), pos + bytes_len

def _decodeStringField(bytes: bytes, pos: int) -> Tuple[str, int]:
    """
    Receives a bytes object and a position in the bytes object and
    returns a tuple with the string and the next position in the bytes
    object.
    """
    length = int.from_bytes(bytes[pos:pos+1], 'big')
    pos += 1
    return bytes[pos : pos + length].decode('utf-8'), pos + length
    

""" Checks whether a bet won the prize or not. """
def has_won(bet: Bet) -> bool:
    return bet.number == LOTTERY_WINNER_NUMBER

"""
Persist the information of each bet in the STORAGE_FILEPATH file.
Not thread-safe/process-safe.
"""
def store_bets(bets: list[Bet]) -> None:
    with open(STORAGE_FILEPATH, 'a+') as file:
        writer = csv.writer(file, quoting=csv.QUOTE_MINIMAL)
        for bet in bets:
            writer.writerow([bet.agency, bet.first_name, bet.last_name,
                             bet.document, bet.birthdate, bet.number])

"""
Loads the information all the bets in the STORAGE_FILEPATH file.
Not thread-safe/process-safe.
"""
def load_bets() -> list[Bet]:
    with open(STORAGE_FILEPATH, 'r') as file:
        reader = csv.reader(file, quoting=csv.QUOTE_MINIMAL)
        for row in reader:
            yield Bet(row[0], row[1], row[2], row[3], row[4], row[5])

Bet.decodeBytes(bytes([1,4,74,111,104,110,4,80,111,114,107,21,205,91,7,10,49,57,56,48,45,48,49,45,48,49,42,0]))