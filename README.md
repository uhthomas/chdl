# chdl
A chan downloader written in Go.

## Currently supported chans:
* 4chan
* 7chan
* 8chan

## Usage
```
go-chdl [<flags>] <url>

Flags:
      --help            Show context-sensitive help (also try --help-long and
                        --help-man).
  -l, --limit=10        Concurrent download limit
  -o, --out="chdl"      Output directory for downloaded files
  -e, --exclude-extras  Exclude extra files

Args:
  <url>  Thread URL

```