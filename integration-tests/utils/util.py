import json
import re
import subprocess


def to_identifier(text):
     return re.sub(r'\W','_', text)


def update_json(infile, new_data, outfile):
    with open(infile) as fh:
        data = json.load(fh)

    update_dict(data, new_data)

    with open(outfile, 'w') as wfh:
        json.dump(data, wfh)


def update_dict(base, other):
    for k, ov in other.items():
        bv = base.get(k)
        if bv is None:
            continue  # nothing to update

        if isinstance(bv, dict) and isinstance(ov, dict):
            update_dict(bv, ov)
        elif isinstance(bv, dict) or isinstance(ov, dict):
            raise ValueError(f'Value mismatch for "{k}": only one is a dict ({bv}, {ov})')
        else:
            base[k] = ov


def run_command(command: str, action: str) -> int:
    print(command)
    proc = subprocess.run(command, capture_output=True, shell=True)
    print(f'---STDOUT---\n{proc.stdout}')
    print(f'---STDERR---\n{proc.stderr}')
    print(f'------------')
    if proc.returncode:
        executable = command.split()[0]
        raise RuntimeError(f'Could not {action}; {executable} returned {proc.returncode}')

