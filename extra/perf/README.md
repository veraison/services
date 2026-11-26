## Requirements

This requires `locust` to generate loads and `jupyter-lab` to analyse the
results. Both are Python applications, and so may be installed via `pip`, or
through your system's package manager.

## Basic flow

1. Run loads and collect data:
   ```
   ./do-runs 20 10 30 raw/$(hostname)-MYCONFIG-set1
   ```
   replace MYCONFIG with a descriptive name of the configuration you're
   testing.

2. Run the Jupyter lab
   ```
   cd notebooks
   jupyter lab
   ```
   Use the link in the output to open the lab page in your browser

3. In the browser Jupyter interface, duplicate one of the existing notebooks
   (right-click -> "duplicate", in the left bar). Rename the notebook to
   reflect your test, then adjust the values in cell `10` to reflect the
   results you've collected (likely, you'd just need to just `config` to be
   MYCONFIG above). Run the notebook.


## Notes

Manual master locust command (you don't need to do this if using the flow
above):
```
locust --locustfile scripts/run-loads.py --master --headless --expect-workers 5 -u5  --run-time 1m  --stop-timeout 10s --csv /tmp/testrun
```
(You'd then need to start `5` workers using `locust --locustfile
scripts/run-loads.py --worker`.)

