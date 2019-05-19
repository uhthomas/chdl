# chdl
A <4,7,8>chan downloader written in Go.

## Currently supported chans:
* 4chan
* 7chan
* 8chan

## Usage
```
Usage of chdl:
  -exclude-extras
        Don't download extra files
  -limit int
        Concurrent download limit (default 10)
  -out string
        Output directory for files. (default "chdl")
  -url string
        Board or thread URL.
url can also be set by passing it after the flags like: 
  chdl -limit 5 https://4chan.org/b
```