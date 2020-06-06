#!/usr/bin/env python

import argparse
import os
import subprocess
from tqdm import tqdm
import time

from parse import walkem, Leakery


def grep(email, db_file):
    # TODO: sorted grep / bissect search
    email = email.lower() + b':'
    r = False
    try:
        for line in open(db_file, 'rb'):
            if line.lower().startswith(email):
                print(line.rstrip(b'\r\n'))
                r = True
    except FileNotFoundError:
        return False
    return r


def main(argv=None):
    parser = argparse.ArgumentParser(description='Query leakery database.')
    parser.add_argument('path', help='leakery database')
    parser.add_argument('email', help='email to search for')
    parser.add_argument(
        '-f',
        '--file',
        action='store_true',
        help='EMAIL is a file containing emails instead of an email',
    )

    args = parser.parse_args(argv)

    if args.file:
        emails = (x.rstrip() for x in open(args.email, 'rb'))
    else:
        emails = [args.email.encode()]

    for email_b in emails:
        path, fname = Leakery.email_path(email_b)
        db_file = os.path.join(args.path, *path, fname + '.txt')
        grep(email_b, db_file)


if __name__ == '__main__':
    main()
