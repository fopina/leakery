#!/usr/bin/env python

import argparse
import os
import subprocess
from tqdm import tqdm
import time

from parse import walkem


def main(argv=None):
    parser = argparse.ArgumentParser(description='Build stats of a leakery database.')
    parser.add_argument('input', help='leakery database')
    parser.add_argument('statsfile', help='file to update with stats')

    args = parser.parse_args(argv)

    totsize = 0
    totfiles = 0
    records = 0
    for f in walkem([args.input]):
        totsize += os.path.getsize(f)
        totfiles += 1
    print('Files to be processed: %d' % totfiles)
    print('Total data size: %s' % tqdm.format_sizeof(totsize))

    fileno = 0
    with tqdm(total=totsize, unit_scale=True) as pbar:
        for f in walkem([args.input]):
            fileno += 1
            pbar.update(os.path.getsize(f))
            records += int(subprocess.check_output(['wc', '-l', f]).split()[0])
            pbar.set_postfix(recs=pbar.format_sizeof(records), fileno=fileno, refresh=False)

    print('Records found: %d' % records)

    with open(args.statsfile, 'a') as f:
        f.write('Stats at ')
        f.write(time.strftime('%d/%m/%y %H:%M'))
        f.write('\n')
        f.write('\nFiles:     ')
        f.write(str(totfiles))
        f.write('\nSize:      ')
        f.write(str(totsize))
        f.write('\nRecords:   ')
        f.write(str(records))
        f.write('\n\n')


if __name__ == "__main__":
    main()
