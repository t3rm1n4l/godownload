// An HTTP parallel downloader library

package download

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	NotStarted = "Not started"
	OnProgress = "On progress"
	Completed  = "Completed"
)

type Status string

type Downloader struct {
	url    string
	conns  int
	file   *os.File
	size   int
	parts  []Part
	start  time.Time
	end    time.Time
	done   chan error
	quit   chan bool
	status string
}

func New() Downloader {
	return Downloader{}
}

func (dl *Downloader) Init(url string, conns int, filename string) (uint64, error) {
	dl.url = url
	dl.conns = conns
	dl.status = NotStarted
	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != 200 {
		return 0, errors.New(resp.Status)
	}

	dl.size, err = strconv.Atoi(resp.Header.Get("Content-Length"))

	if err != nil {
		return 0, errors.New("Not supported for download")
	}

	_, err = os.Stat(filename)
	if os.IsExist(err) {
		os.Remove(filename)
	}

	dl.file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return 0, err
	}

	dl.parts = make([]Part, dl.conns)
	size := dl.size / dl.conns

	for i, part := range dl.parts {
		part.id = i
		part.url = dl.url
		part.offset = i * size
		switch {
		case i == dl.conns-1:
			part.size = dl.size - size*i
			break
		default:
			part.size = size
		}
		part.dlsize = 0
		part.file = dl.file
		dl.parts[i] = part
	}

	return uint64(dl.size), nil
}

func (dl *Downloader) StartDownload() {
	dl.done = make(chan error, dl.conns)
	dl.quit = make(chan bool, dl.conns)
	dl.status = OnProgress
	for i := 0; i < dl.conns; i++ {
		go dl.parts[i].Download(dl.done, dl.quit)
	}
	dl.start = time.Now()
}

func (dl Downloader) GetProgress() (status string, total, downloaded int, elapsed time.Duration) {
	dlsize := 0
	for _, part := range dl.parts {
		dlsize += part.dlsize
	}
	return dl.status, dl.size, dlsize, time.Now().Sub(dl.start)
}

func (dl *Downloader) Wait() error {
	var err error = nil
	for i := 0; i < dl.conns; i++ {
		e := <-dl.done
		if e != nil {
			err = e
			dl.status = err.Error()
			for j := i; j < dl.conns; j++ {
				dl.quit <- true
			}
		}
	}
	close(dl.done)
	dl.end = time.Now()
	dl.file.Close()
	if dl.status == OnProgress {
		dl.status = Completed
	}
	return err
}

func (dl *Downloader) Download() error {
	dl.StartDownload()
	return dl.Wait()
}

type Part struct {
	id     int
	url    string
	offset int
	dlsize int
	size   int
	file   *os.File
}

func (part *Part) Download(done chan error, quit chan bool) error {
	client := http.Client{}
	size := 4096
	buffer := make([]byte, size)
	req, err := http.NewRequest("GET", part.url, nil)
	defer func() {
		done <- err
	}()

	if err != nil {
		return err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", part.offset, part.offset+part.size-1))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	for {
		select {
		case <-quit:
			return nil
		default:
		}

		nbytes, err := resp.Body.Read(buffer[0:size])
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		nbytes, err = part.file.WriteAt(buffer[0:nbytes], int64(part.offset+part.dlsize))
		if err != nil {
			return nil
		}
		part.dlsize += nbytes
		remaining := part.size - part.dlsize
		switch {
		case remaining == 0:
			return nil
		case remaining < 4096:
			size = part.size - part.dlsize
		}
	}

	return nil
}
