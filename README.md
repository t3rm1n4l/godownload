# Parallel Downloader

A simple parallel downloader package written in go

Benchmark shows that it is faster than aria2c with equal number of splits

## Usage example
    $ cd cmd
    $ go get github.com/t3rm1n4l/godownload
    $ make
    $ ./downloader -c=5 -o=outputfile -u=http://releases.ubuntu.com/precise/ubuntu-12.04.1-alternate-amd64.iso
    File size is 734MB
    Downloaded 0.23% of 734MB, at 121KB/s

## Library usage
Import `"github/com/t3rm1n4l/godownload"` in your code.
See `cmd/main.go` for implementation of downloader

    d := download.New()
    size, err := d.Init(url, connections, filename)
    // Used for blocked wait for full file download
    err = d.Download()
        or
    // Just returns by starting download. We must call d.Wait() to block wait
    // till download is complete.
    d.StartDownload()

    // Used to gather information about status of on-going download
    d.GetProgress()


