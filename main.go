package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/eventials/go-tus"
	"github.com/schollz/progressbar/v3"
)

func main() {
	var fp, uplPath, headers, metadata string

	flag.StringVar(&fp, "f", "", "full path to your file")
	flag.StringVar(&uplPath, "u", "", "TUS upload server path")
	flag.StringVar(&headers, "H", "", "custom request headers in JSON format")
	flag.StringVar(&metadata, "m", "", "metadata in JSON format")
	flag.Parse()

	if fp == "" || uplPath == "" {
		flag.Usage()
		return
	}

	// listen to kill signal
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	defer signal.Stop(signals)

	// create the config.
	config := tus.DefaultConfig()
	if headers != "" {
		headersMap := make(map[string]string)
		if err := json.Unmarshal([]byte(headers), &headersMap); err != nil {
			panic(err)
		}
		clientHeaders := http.Header{}
		for k, v := range headersMap {
			clientHeaders.Add(k, v)
		}
		config.Header = clientHeaders
	}

	metadataMap := tus.Metadata{}
	if metadata != "" {
		if err := json.Unmarshal([]byte(metadata), &metadataMap); err != nil {
			panic(err)
		}
	}

	if _, err := url.Parse(uplPath); err != nil {
		panic(err)
	}

	// open the file.
	f, err := os.Open(fp)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// create the tus client.
	client, err := tus.NewClient(uplPath, config)
	if err != nil {
		panic(err)
	}

	// create an upload from a file.
	upload, err := tus.NewUploadFromFile(f)
	if err != nil {
		panic(err)
	}
	uplMeta := upload.Metadata
	for k, v := range metadataMap {
		uplMeta[k] = v
	}

	bar := progressbar.Default(100)
	bar.Describe(fmt.Sprintf("Uploading %s", filepath.Base(fp)))

	// create the uploader.
	uploader, err := client.CreateUpload(upload)
	if err != nil {
		panic(err)
	}

	// upload progress channel
	prg := make(chan tus.Upload)
	go func() {
		for data := range prg {
			p := float64(data.Offset()) / float64(data.Size())
			p = math.Ceil(p * 100)
			bar.Set(int(p))
		}
	}()

	uploader.NotifyUploadProgress(prg)

	// start the uploading process.
	if err := uploader.Upload(); err != nil {
		panic(err)
	} else {
		log.Println("Upload finished successfully! 🏆")
		os.Exit(0)
	}

	<-signals
	go func() {
		<-signals

		f.Close()
		uploader.Abort()
		bar.Finish()
		close(prg)

		os.Exit(1)
	}()

	log.Println("Gracefully shutting down...")

	close(prg)
	// stop the uploading process.
	uploader.Abort()
}
