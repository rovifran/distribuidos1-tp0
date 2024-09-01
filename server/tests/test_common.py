from common.utils import *
import os
import unittest

class TestUtils(unittest.TestCase):

    def test_bet_decodeBytes_must_return_correct_bet_instance(self):
        encodedBet = bytearray([1,4,74,111,104,110,4,80,111,114,107,21,205,91,7,10,49,57,56,48,45,48,49,45,48,49,42,0])
        b = Bet.decodeBytes(encodedBet)
        self.assertEqual(1, b.agency)
        self.assertEqual('John', b.first_name)
        self.assertEqual('Pork', b.last_name)
        self.assertEqual(123456789, b.document)
        self.assertEqual(datetime.date(1980, 1, 1), b.birthdate)
        self.assertEqual(42, b.number)

    def tearDown(self):
        if os.path.exists(STORAGE_FILEPATH):
            os.remove(STORAGE_FILEPATH)

    def test_bet_init_must_keep_fields(self):
        b = Bet('1', 'first', 'last', '10000000','2000-12-20', 7500)
        self.assertEqual(1, b.agency)
        self.assertEqual('first', b.first_name)
        self.assertEqual('last', b.last_name)
        self.assertEqual('10000000', b.document)
        self.assertEqual(datetime.date(2000, 12, 20), b.birthdate)
        self.assertEqual(7500, b.number)

    def test_has_won_with_winner_number_must_be_true(self):
        b = bet('1', 'first', 'last', 10000000,'2000-12-20', LOTTERY_WINNER_NUMBER)
        self.assertTrue(has_won(b))

    def test_has_won_with_winner_number_must_be_true(self):
        b = Bet('1', 'first', 'last', 10000000,'2000-12-20', LOTTERY_WINNER_NUMBER + 1)
        self.assertFalse(has_won(b))

    def test_store_bets_and_load_bets_keeps_fields_data(self):
        to_store = [Bet('1', 'first', 'last', '10000000','2000-12-20', 7500)]
        store_bets(to_store)
        from_load = list(load_bets())

        self.assertEqual(1, len(from_load))
        self._assert_equal_bets(to_store[0], from_load[0])


    def test_store_bets_and_load_bets_keeps_registry_order(self):
        to_store = [
            Bet('0', 'first_0', 'last_0', '10000000','2000-12-20', 7500),
            Bet('1', 'first_1', 'last_1', '10000001','2000-12-21', 7501),
        ]
        store_bets(to_store)
        from_load = list(load_bets())

        self.assertEqual(2, len(from_load))
        self._assert_equal_bets(to_store[0], from_load[0])
        self._assert_equal_bets(to_store[1], from_load[1])

    def _assert_equal_bets(self, b1, b2):
        self.assertEqual(b1.agency, b2.agency)
        self.assertEqual(b1.first_name, b2.first_name)
        self.assertEqual(b1.last_name, b2.last_name)
        self.assertEqual(b1.document, b2.document)
        self.assertEqual(b1.birthdate, b2.birthdate)
        self.assertEqual(b1.number, b2.number)

if __name__ == '__main__':
    unittest.main()

