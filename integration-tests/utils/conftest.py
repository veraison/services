import os
import sys

import hooks
from util import to_identifier, clear_stores

class TavernTest:

    def __init__(self, test_dict):
        self.full_name = test_dict['test_name']
        self.name = self.full_name.split('[')[0]

        self.common_vars, self.test_vars = {}, {}
        for entry in test_dict['includes']:
            if entry['name'] == 'common':
                self.common_vars = entry['variables']
            elif entry['name'].startswith('parametrized'):
                self.test_vars = entry['variables']


def pytest_tavern_beta_before_every_test_run(test_dict, variables):
    test = TavernTest(test_dict)
    setup(test, variables)

def setup(test, variables):
    clear_stores()
    test_id = to_identifier(test.name)
    handler = getattr(hooks, f'setup_{test_id}', None)
    if handler:
        handler(test, variables)


