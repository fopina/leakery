#!/usr/bin/env python

import argparse
import os
import re
import sys
import multiprocessing
from tqdm import tqdm


class Leakery:
    default = 'all'

    def __init__(self, mode, output):
        m = getattr(self, 'mode_' + mode, None)
        if not m:
            raise Exception('invalid mode')
        self._m = m
        self._output = output

        self._mail_re = re.compile(rb'^([a-zA-Z]{1,3})')
        self._domain_clean = re.compile(rb'[^a-zA-Z0-9\_\-\.]')
        s = getattr(self, 'setup_' + mode, None)
        if s:
            s()

    def setup_plain(self):
        self._re = re.compile(rb'^(.*?)[:;\s](.*)$')

    def mode_plain(self, line):
        m = self._re.findall(line.strip())
        if m:
            return m[0]
        return None

    def setup_user_and_ip(self):
        self._re = re.compile(rb'^[^:]*:([^:]*):\d+\.\d+\.\d+\.\d+:(.*)$')

    def mode_user_and_ip(self, line):
        return self.mode_plain(line)

    def setup_user_included(self):
        self._re = re.compile(rb'^[^:]*:([^:]*):(.*)$')

    def mode_user_included(self, line):
        return self.mode_plain(line)

    def setup_all(self):
        res = []
        for mode in ('plain', 'user_and_ip', 'user_included'):
            met = getattr(self, 'setup_' + mode)
            met()
            res.append(self._re)
        self._re = res

    def mode_all(self, line):
        for _re in self._re:
            m = _re.findall(line.strip())
            if m and b'@' in m[0][0]:
                return m[0]
        return None

    def handle(self, line):
        r = len(line)
        line = line.rstrip()
        c = self._m(line)
        if c is None or b'@' not in c[0]:
            return (False, r, line)
        carr = c[0].split(b'@')
        m = self._mail_re.match(carr[0])
        kk = []
        jc = b':'.join(c)
        if m:
            _u = m.group(1).decode().lower()
            lu = len(_u)
            if lu > 1:
                kk = list(_u)[:-1]
                _u = _u[-1]
            # is it worth shaving off these few bytes?
            # jc = jc[lu:]
        else:
            _u = '_'
        return (True, r, _my_os_path_join(self._output, *kk), _u + '.txt', jc, carr[0])

    @classmethod
    def modes(cls):
        return [m[5:] for m in dir(cls) if m[:5] == 'mode_']


def _my_os_path_join(*args):
    return '/'.join(args)


def walkem(dirs):
    for d in dirs:
        if os.path.isdir(d):
            for p, _, files in os.walk(d):
                for f in files:
                    yield os.path.join(p, f)
        else:
            yield d


class FDCache:
    def __init__(self, limit=30000):
        self._cache = {}
        self._limit = limit

    def open(self, kd1, kd2):
        fname = _my_os_path_join(kd1, kd2)
        if fname not in self._cache:
            try:
                _f = open(fname, 'ab')
            except FileNotFoundError:
                os.makedirs(kd1)
                _f = open(fname, 'ab')
            if len(self._cache) >= self._limit:
                self.close_all()
            self._cache[fname] = _f
        return self._cache[fname]

    def close_all(self):
        for v in self._cache.values():
            v.close()
        self._cache = {}


def main(argv=None):
    parser = argparse.ArgumentParser(description='Index random dumps.')
    parser.add_argument(
        'input',
        action='append',
        help='input directories or files: if directory, subdirectories are processed as well',
    )
    parser.add_argument('-d', '--output', help='output directory')
    parser.add_argument(
        '-p',
        '--parser',
        choices=Leakery.modes(),
        default=Leakery.default,
        help='email/password parser to use',
    )
    parser.add_argument(
        '-l',
        '--fd-cache',
        default=30000,
        type=int,
        help='file descriptor cache (to reduce open/close calls) - set ulimit maxfiles accordingly',
    )
    parser.add_argument(
        '-n', '--workers', type=int, help='number of workers (default single-process)'
    )

    args = parser.parse_args(argv)

    # strip / for the 'dumber' _my_os_path_join
    args.output = args.output.rstrip('/')
    prs = Leakery(args.parser, args.output)
    fd_cache = FDCache(limit=args.fd_cache)

    files_done = set()
    session = _my_os_path_join(args.output, '.session.sav')
    if os.path.exists(session):
        session_data = open(session, 'r').read().splitlines()
        split_data = session_data.index('=')
        old_args = session_data[:split_data]
        files_done = set(session_data[split_data + 1 :])
        if old_args == sys.argv[1:]:
            msg = (
                'Session save was found and matches parameters. %d files already done. Resume?'
                % len(files_done)
            )
            shall = input('%s (y/n) ' % msg).lower().strip()
            if shall != 'y':
                print('Remove file %s and re-run' % session)
                exit(1)
        else:
            print('Session exists but args do not match')
            print('%s != %s' % (old_args, sys.argv[1:]))
            print('Remove file %s and re-run to start new session' % session)
            exit(1)
    os.makedirs(args.output, exist_ok=True)
    if not files_done:
        session_fd = open(session, 'w')
        for arg in sys.argv[1:]:
            session_fd.write(arg)
            session_fd.write('\n')
        session_fd.write('=\n')
    else:
        session_fd = open(session, 'a')
    totsize = 0
    totfiles = 0
    donefiles = 0
    for f in walkem(args.input):
        if f in files_done:
            donefiles += 1
        else:
            totsize += os.path.getsize(f)
            totfiles += 1
    print('Files to be processed: %d' % totfiles)
    if donefiles:
        print('  %d files processed in previous run(s)' % donefiles)
    print('Total data size: %s' % tqdm.format_sizeof(totsize))

    if args.workers:
        p = multiprocessing.Pool(args.workers)
        _map = p.imap_unordered
    else:
        _map = map

    error = os.path.join(args.output, 'errors.log')
    error_fd = open(error, 'ab')
    stats = [0, 0]
    fileno = 0

    with tqdm(total=totsize, unit_scale=True) as pbar:
        for f in walkem(args.input):
            if f in files_done:
                continue
            fileno += 1
            with open(f, 'rb') as fh:
                for d in _map(prs.handle, fh):
                    pbar.update(d[1])
                    stats[int(d[0])] += 1
                    if d[0]:
                        of = fd_cache.open(d[2], d[3])
                        of.write(d[4])
                        of.write(b'\n')
                    else:
                        error_fd.write(d[2])
                        error_fd.write(b'\n')
                    pbar.set_postfix(
                        hits=pbar.format_sizeof(stats[1]),
                        err=pbar.format_sizeof(stats[0]),
                        fileno=fileno,
                        refresh=False,
                    )
            session_fd.write(f)
            session_fd.write('\n')
            session_fd.flush()  # make sure progress is not lost

    fd_cache.close_all()
    print('Records found: %d' % stats[1])
    print('Errors: %d' % stats[0])


if __name__ == '__main__':
    main()
