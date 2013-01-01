// An HTTP parallel downloader client

package main

import (
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/t3rm1n4l/godownload"
	"os"
	"time"
)

func DisplayProgress(dl *download.Downloader) {
	for {
		status, total, downloaded, elapsed := dl.GetProgress()
		fmt.Fprintf(os.Stdout, "Downloaded %.2f%% of %s, at %s/s\r", float64(downloaded)*100/float64(total), humanize.Bytes(uint64(total)), humanize.Bytes(uint64(float64(downloaded)/elapsed.Seconds())))
		switch {
		case status == download.Completed:
			fmt.Println("\nSuccessfully completed download in", elapsed)
			return
		case status == download.OnProgress:
		case status == download.NotStarted:
		default:
			fmt.Printf("\nFailed: %s\n", status)
			os.Exit(1)
		}
		time.Sleep(time.Second)
	}
}

func main() {

	var url = flag.String("u", "", "Download file url")
	var conns = flag.Int("c", 1, "Number of connections")
	var outfile = flag.String("o", "", "Output filename")
	flag.Parse()

	if *outfile == "" || *url == "" {
		flag.Usage()
		os.Exit(1)
	}

	d := download.New()
	size, err := d.Init(*url, *conns, *outfile)
	fmt.Printf("File size is %s\n", humanize.Bytes(size))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	d.StartDownload()
	go d.Wait()
	DisplayProgress(&d)

}
