# leakery

Set of scripts to index (in files) leaks such as the "Break Collection"

## Usage

Use `parse.py` to extract `email`/`password` combos from the files

```
./parse.py -d /target/directory/ /path/to/BreachCompilation
```

Use `cleanup.py` to `sort -u` (sort and remove duplicates) the created files

```
./cleanup.py -n 4 /target/directory/
```

Use `-h` on any for tweak options (such as `-T` in `cleanup.py` that is passed to `sort` to use specific directory for temporary files)
