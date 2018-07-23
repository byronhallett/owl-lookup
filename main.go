package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

const owlURL = "https://owlbot.info/api/v2/dictionary/%s"

// Definition is used to unmarshal owlbot responses
type Definition struct {
	Type       string `json:"type,omitempty"`
	Definition string `json:"definition,omitempty"`
	Example    string `json:"example,omitempty"`
}

// Looks up one word on owlbot
func lookupWord(word string, results *map[string]string, wg *sync.WaitGroup, queue chan int) {
	defer wg.Done()
	var definitions []Definition
	wordURL := fmt.Sprintf(owlURL, word)
	// This will block until the queue is short enough
	queue <- 1
	response, err := http.Get(wordURL)
	// Clear our requests from the queue
	<-queue
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(body, &definitions)
	// Insert the empty string if no def found
	if len(definitions) < 1 {
		(*results)[word] = ""
	} else {
		// Insert a ||| delimited string for one or more def
		defs := make([]string, len(definitions))
		for i, def := range definitions {
			defs[i] = def.Definition
		}
		(*results)[word] = strings.Join(defs, "|||")
	}
}

func getKeys(fromMap *map[string]string) *[]string {
	// Preallocate keys slice
	keys := make([]string, len(*fromMap))
	// Extract each key
	i := 0
	for k := range *fromMap {
		keys[i] = k
		i++
	}
	return &keys
}

func main() {
	// Prepare to read from stdin
	scanner := bufio.NewScanner(os.Stdin)
	dict := make(map[string]string)
	var wg sync.WaitGroup
	// lets us throttle requests
	queue := make(chan int, 50)
	for scanner.Scan() {
		wg.Add(1)
		w := scanner.Text()
		go lookupWord(w, &dict, &wg, queue)
	}
	if scanner.Err() != nil {
		log.Fatal(scanner.Err())
	}
	wg.Wait()
	words := getKeys(&dict)
	for _, word := range *words {
		fmt.Printf("%s: %s\n", word, dict[word])
	}
}
