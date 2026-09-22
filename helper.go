package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type returnVals struct {
		Error string `json:"error"`
	}

	resBody := returnVals{
		Error: msg,
	}
	data, err := json.Marshal(resBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}

func replaceBadWords(content string) string {
	const replaceString = "****"
	badWords := []string{ "kerfuffle", "sharbert", "fornax" }

	words := strings.Split(content, " ")

	for idx, word := range words {
		for _, badWord := range badWords {
			if strings.ToLower(word) == badWord {
				words[idx] = strings.Replace(word, word, replaceString, 1)
			}
		}
	}

	return strings.Join(words, " ")
}