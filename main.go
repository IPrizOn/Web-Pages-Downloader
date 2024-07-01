package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	s "strings"
	"sync"
)

func main() {
	threadsCount, pageURLList, errFlags := readConsolFlags()
	if errFlags != nil {
		panic(errFlags)
	}

	waitGroup := sync.WaitGroup{}
	jobs := make(chan string, 100)

	for thread := int8(1); thread <= threadsCount; thread++ {
		go worker(jobs, &waitGroup)
	}

	for _, pageURL := range pageURLList {
		waitGroup.Add(1)
		jobs <- pageURL
	}

	waitGroup.Wait()
	close(jobs)
}

func worker(jobs <-chan string, waitGroup *sync.WaitGroup) {
	for job := range jobs {
		pageData, errDownload := downloadByURL(job)
		if errDownload != nil {
			panic(errDownload)
		}

		fileName := convPageURLToFileName(job)

		errStore := storeFile(fileName, pageData)
		if errStore != nil {
			panic(errStore)
		}

		waitGroup.Done()
	}
}

func readConsolFlags() (threadsNum int8, pageURLList []string, err error) {
	var threads int
	var urls string
	flag.IntVar(&threads, "threads", 1, "threads count")
	flag.StringVar(&urls, "urls", "", "urls list")

	flag.Parse()
	if threads == 0 || urls == "" {
		err = fmt.Errorf("flags: missing arguments or commands")
	}

	splitedUrlList := s.Split(urls, ",")

	return int8(threads), splitedUrlList, err
}

func downloadByURL(pageURL string) (pageData []byte, err error) {
	response, errResponse := http.Get(pageURL)
	if errResponse != nil {
		err = errResponse
	}
	defer response.Body.Close()

	scanner := bufio.NewScanner(response.Body)

	for scanner.Scan() {
		pageData = append(pageData, scanner.Bytes()...)
	}
	if errScanner := scanner.Err(); errScanner != nil {
		err = errScanner
	}

	return pageData, err
}

func convPageURLToFileName(pageURL string) (fileName string) {
	fileName = s.ReplaceAll(pageURL, "://", "-")
	fileName = s.ReplaceAll(fileName, ".", "_")

	return fileName + ".txt"
}

func storeFile(fileName string, pageData []byte) (err error) {
	file, errCreate := os.Create(fileName)
	if errCreate != nil {
		return errCreate
	}
	defer file.Close()

	_, errWrite := file.Write(pageData)
	if errWrite != nil {
		return errWrite
	}

	return nil
}
