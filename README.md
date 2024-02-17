# Yesterday

Yesterday prints file names from the dump.

Usage:

    yesterday [-c | -C | -d] [-n daysago | -t [[yy]yy]mm]dd] file...

Yesterday prints the names of the files from the most recent dump. Since
dumps are done early in the morning, yesterday's files are really in today's
dump. For example, if today is February 11, 2003,

```sh
yesterday /home/am3/rsc/.profile
```

prints

```
/dump/am/2003/0211/home/am3/rsc/.profile
```

In fact, the implementation is to select the most recent dump in the current
year, so the dump selected may not be from today. Yesterday does not
guarantee that the string it prints represents an existing file.

By default, yesterday prints the names of the dump files corresponding to
the named files. The first set of options changes this behavior.

The `-c` flag causes yesterday to copy the dump files over the named files.

The `-C` flag causes yesterday to copy the dump files over the named files
only when they differ.

The `-d` flag causes yesterday to run `diff` to compare the dump files
with the named files.

The `-n` flag causes yesterday to select the dump `daysago` prior to the current
day.

The `-t` flag causes yesterday to select other day’s dumps, with a format of
1, 2, 4, 6, or 8 digits of the form d, dd, mmdd, yymmdd, or yyyymmdd.

## Examples

See what’s changed in the last week in your profile:

```sh
$ yesterday −d −n 7 ~/.profile
diff -c /dump/am/2024/0211/home/mpd/.profile /home/mpd/.profile
```

Restore your profile from yesterday:

```sh
$ yesterday −c ~/.profile
cp /dump/am/2024/0217/home/mpd/.profile /home/mpd/.profile
```
