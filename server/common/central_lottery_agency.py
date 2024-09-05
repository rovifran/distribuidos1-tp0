from common.utils import *
from multiprocessing import Queue

"""
Class that keeps track of the winners of the lottery.
"""
class CentralLotteryAgency:
    def __init__(self):
        self.winners_per_agency = {}

    """
    Adds the corresponding winner to the winners_per_agency dictionary.
    """
    def _add_winner(self, agency: int, dni: int):
        if agency not in self.winners_per_agency:
            self.winners_per_agency[agency] = set()
        self.winners_per_agency.get(agency, set()).add(dni)

    """
    Returns a dictionary with the agency as the key and the value
    being all the distinct winners of that agency.
    """
    def get_winners(self):
        return self.winners_per_agency
    
    """
    Reads the bets from the file and determines the winners.
    """
    def determine_winners(self):
        for bet in load_bets():
            if has_won(bet):
                self._add_winner(bet.agency, bet.document)
