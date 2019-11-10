# Torbula

Torbula is file system based torrent client. It detects `torrent` files from
specific directory and download them. When seeding is completed, Torbula then
moves files to a final directory.

## Installation

```
go install github.com/isutare412/torbula/cmd/...
```

## How to use

```
$ torbula [settings.ini]
```

Torbula needs a setting file to execute. A example of the setting file is in
`cmd/torbula/settings.ini`. Belows are brief steps of how Torbula works.

1. Put `torrent` files into `src_dir`, which can be set in `settings.ini` file.
2. The files are downloaded into `tmp_dir`.
3. If the files are downloaded completely, Torbula then starts to seed
    the torrents for `seed_time` duration, which also can be set in
    `settings.ini` file.
4. When the seeding is done, downloaded files are moved from `tmp_dir` to
    `dst_dir`.
5. The `torrent` files you put into `src_dir` is removed.

While Torbula is downloading, you can see the download status in a `status.txt`
file.
