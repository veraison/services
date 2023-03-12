Files in this directory:

**checkers.py**

Contains definitions the checker functions used to validate results for test
stages. They are referenced by `verify_response_with` entries in tests.

**confteset.py**

This is the configuration file for py.test, and, therefore, Tavern (which is a
py.test plugin). This is used to set up the hooks mechanism (see below).

**generators.py**

Functions used to generate test case input data. Generated data is stored under
`../__generated__` directory.

**hooks.py**

Test setup hooks. These rely on [Tavern's hooks
mechanism](https://tavern.readthedocs.io/en/latest/basics.html#hooks). A hook
whose name matches `setup_<test name>` gets invoked before the corresponding
test is run. Unlike fixtures, hooks have access to test's variables, allowing
test-specific operation.

**util.py**

Miscellaneous utilities used by the other modules.
