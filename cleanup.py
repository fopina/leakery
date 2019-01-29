#!/usr/bin/env python

import argparse
import os
import re
import multiprocessing
from parse import walkem
import subprocess
from tqdm import tqdm


def main(argv=None):
    parser = argparse.ArgumentParser(description='Sort and uniq each index file.')
    parser.add_argument('input', help='input directory')
    parser.add_argument('-n', '--workers', type=int,
                        help='number of workers (default single-process)')
    parser.add_argument('-T', '--temporary-directory',
                        help='use DIR for temporaries, not $TMPDIR or /tmp')

    args = parser.parse_args(argv)

    totsize = 0
    totfiles = 0
    for f in walkem(args.input):
        if f[-4:] != '.txt':
            continue
        totsize += os.path.getsize(f)
        totfiles += 1
    print('Files to be processed: %d' % totfiles)
    print('Total data size: %s' % tqdm.format_sizeof(totsize))

    cmd = [
        'sort',
        '-u', '',
        '-o', '',
    ]

    if args.workers:
        cmd.extend(['--parallel', str(args.workers)])
    if args.temporary_directory:
        cmd.extend(['-T', args.temporary_directory])
    
    sb = sa = 0
    ratio = 100

    with tqdm(total=totsize, unit_scale=True) as pbar:
        for f in walkem(args.input):
            if f[-4:] != '.txt':
                continue
            size = os.path.getsize(f)
            sb += size
            fn = f[len(args.input):]
            fo = f + '.sort'
            pbar.set_postfix(
                file=fn,
                size=tqdm.format_sizeof(size),
                ratio=ratio,
            )
            cmd[2] = f
            cmd[4] = fo
            subprocess.check_call(cmd)
            pbar.update(size)
            sa += os.path.getsize(fo)
            ratio = sa * 100 / sb
            os.rename(fo, f)


if __name__ == "__main__":
    main()
