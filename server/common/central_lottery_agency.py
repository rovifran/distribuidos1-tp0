from common.utils import *

class CentralLotteryAgency:
    def __init__(self):
        self.winners_per_agency = {}

    def add_bets(self, bets: Bet):
        store_bets(bets)

    def _add_winner(self, agency: int, dni: int):
        if agency not in self.winners_per_agency:
            self.winners_per_agency[agency] = []
        self.winners_per_agency.get(agency, []).append(dni)

    def get_winners(self):
        return self.winners_per_agency
    
    def determine_winners(self):
        for bet in load_bets():
            if has_won(bet):
                self._add_winner(bet.agency, bet.document)
