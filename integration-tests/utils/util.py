import json
import re


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

        if isinstance(bv, dict) and isinstance(ov, dict):
            update_dict(bv, ov)
        elif isinstance(bv, dict) or isinstance(ov, dict):
            raise ValueError(f'Value mismatch for "{k}": only one is a dict ({bv}, {ov})')
        else:
            base[k] = ov


