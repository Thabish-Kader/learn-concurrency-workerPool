// Inspiration - https://blog.devgenius.io/concurrency-with-sample-project-in-golang-297400beb0a4
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Result struct {
	Month		string	`json:"month"`
	Num			int		`json:"num"`
	Link		string	`json:"link"`
	Year		string	`json:"year"`
	News		string	`json:"news"`
	SafeTitle	string 	`json:"safe_title"`
	Transcript	string	`json:"transcript"`
	Alt			string	`json:"alt"`
	Img			string	`json:"img"`
	Title		string	`json:"title"`
	Day			string	`json:"day"`
}

var finalResult []Result

const Url = "https://xkcd.com"

func fetch(n int) (*Result, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET",strings.Join([]string{Url, strconv.Itoa(n),"info.0.json"},"/") , nil)
	if err != nil {
		return nil, fmt.Errorf("Error while creating new request %v", err)
	}

	resp, err := client.Do(req)
	if err !=nil {
		return nil, fmt.Errorf("Error while fetching %v", err)
	}

	data := Result{}

	if resp.StatusCode != http.StatusOK {
		data = Result{}
	}else {
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, fmt.Errorf("Error while decoding %v", err)
		}
	}

	defer resp.Body.Close()

	return &data, nil
}

func allocateJobs(noOfJobs int,jobs chan<- int) {
	for task :=0; task <= noOfJobs;task++ {
		jobs <- task
	} 
	close(jobs)
}

func worker(wg *sync.WaitGroup, jobs <-chan int,results chan<- Result) {
	for job := range jobs {
		data, err := fetch(job)
		fmt.Println("Fetching data -->", data.Num)
		if err != nil {
			fmt.Println("Error while fetching", err)
		}
		results <- *data
	}
	wg.Done()
}

func getResults(results <-chan Result, finalResult *[]Result) {
	for result := range results {
		*finalResult = append(*finalResult, result)
	}
}

func main() {
	elapsedTime := time.Now()

	jobs := make(chan int,100)
	results := make(chan Result,100)
	var finalResult []Result

	var wg sync.WaitGroup

	go allocateJobs(1000,jobs)

	// Create workers
	for w:=0;w <=100 ; w++ {
		wg.Add(1)
		go worker(&wg,jobs,results)
	}

	go getResults(results, &finalResult)

	wg.Wait()
	close(results)

	err := writeToFile(finalResult)
	if err !=nil {
		log.Fatalf("Error while writing to file %v\n", err)
	}
	
	fmt.Printf("Time taken: %s", time.Since(elapsedTime))
}

func writeToFile(data []Result) error  {
	file , err := os.Create("result.json")
	if err != nil {
		log.Fatalf("Error while creating file %v\n", err)
		return err
	}
	defer file.Close()

	jsonData, err := json.MarshalIndent(data,"","")
	if err !=nil {
		log.Fatalf("Error while marshalling %v\n", err)
		return err 
	}

	_, err = file.Write(jsonData)
	if err != nil {
		log.Fatalf("Error while writing to file %v\n", err)
		return err
	}
	return nil
}
