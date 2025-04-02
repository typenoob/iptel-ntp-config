package log

import (
	"encoding/json"
	"log"
	"os"
)

func saveToJSON(entries []Entry) {
	file, err := os.Create("result.json")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(entries)
	if err != nil {
		log.Fatalln(err)
	}
}

func readFromJSON() []Entry {
	var entries []Entry
	file, err := os.Open("result.json")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		log.Fatalln(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&entries)
	if err != nil {
		log.Fatalln(err)
	}
	return entries
}
