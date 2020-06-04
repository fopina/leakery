# leakery

Set of scripts to index (in files) leaks such as the "Break Collection"

## Usage

### parse.py

Extract `email`/`password` combos from the files

```
./parse.py -d /target/directory/ /path/to/BreachCompilation
```

### cleanup.py
`sort -u` (sort and remove duplicates) the created files

```
./cleanup.py -n 4 /target/directory/
```

Use `-h` on any for tweak options (such as `-T` in `cleanup.py` that is passed to `sort` to use specific directory for temporary files)

### stats.py

Build stats (total records and such) and log them

```
./stats.py path/to/db /path/to/statsfile
```

### merge.py

Merge 2 DBs - i.e. after indexing new leak on new (faster) disk, merge it with the old one

```
./merge.py path/to/one path/to/other
```

`other` will be merged into `one`
