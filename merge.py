#!/usr/bin/env python

import argparse
import os
import shutil
from tqdm import tqdm

from parse import walkem


def valid_db(path):
    files = os.listdir(path)
    for i in range(97, 123):
        if f'{chr(i)}.txt' in files:
            return True
    return False


def main(argv=None):
    parser = argparse.ArgumentParser(description='Build stats of a leakery database.')
    parser.add_argument('main', help='main DB location')
    parser.add_argument('new', help='newer, smaller, DB location (to merge into)')

    args = parser.parse_args(argv)

    for x in (args.main, args.new):
        if not valid_db(x):
            print(f'{x} does not seem to be a leakery db')
            exit(1)

    # normalize paths
    args.main = os.path.abspath(args.main)
    args.new = os.path.abspath(args.new)

    totsize = 0
    totfiles = 0

    for f in walkem([args.new]):
        totsize += os.path.getsize(f)
        totfiles += 1

    print('Files to be processed: %d' % totfiles)
    print('Total data size: %s' % tqdm.format_sizeof(totsize))

    sn = len(args.new)

    fileno = 0
    with tqdm(total=totsize, unit_scale=True) as pbar:
        for f in walkem([args.new]):
            fileno += 1
            pbar.update(os.path.getsize(f))
            pbar.set_postfix(fileno=fileno, refresh=False)
            with open(f, 'rb') as fsrc, open(args.main + f[sn:], 'ab') as fdst:
                shutil.copyfileobj(fsrc, fdst)


if __name__ == '__main__':
    main()
