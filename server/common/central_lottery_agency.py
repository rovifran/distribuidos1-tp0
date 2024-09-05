from common.utils import *
from multiprocessing import Queue

class CentralLotteryAgency:
    def __init__(self):
        self.winners_per_agency = {}

    def _add_winner(self, agency: int, dni: int):
        if agency not in self.winners_per_agency:
            self.winners_per_agency[agency] = set()
        self.winners_per_agency.get(agency, set()).add(dni)

    def get_winners(self):
        return self.winners_per_agency
    
    def determine_winners(self):
        for bet in load_bets():
            if has_won(bet):
                self._add_winner(bet.agency, bet.document)
