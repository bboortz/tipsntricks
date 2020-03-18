package main

import (
	"context"
	"net/http"
	"github.com/bboortz/goborg/pkg/appcontext"
	"encoding/json"
	"log"
	)

type SomeObj struct {
        Foo string `json:"foo"`
        Bar string `json:"bar"`
}

type Response struct {
        Message string `json:"message"`
}

// respondJSON makes the response with payload as json format
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
        response, err := json.Marshal(payload)
        if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                w.Write([]byte(err.Error()))
                return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(status)
        w.Write([]byte(response))
}

// respondError makes the error response with payload as json format
func respondError(w http.ResponseWriter, code int, message string) {
        respondJSON(w, code, map[string]string{"error": message})
	}	

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	// retrieve the current context
	ctx := r.Context()

	// decoding json to struct
	var someObj SomeObj
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&someObj); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// Do some ...
	doSomeMagic(ctx, someObj)

	// respond
	response := Response{
		Message: "Ok",
	}
    respondJSON(w, http.StatusOK, response)
}

func doSomeMagic(ctx context.Context, magicObj SomeObj) {
	logger := appcontext.Logger(ctx)
	logger.Info("WooHooo")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/post", handlePostRequest)

	logger := appcontext.Logger(context.Background())
	logger.Info("Start server on port :8080")
	cmux := ContextMiddleware(mux)
	log.Fatal(http.ListenAndServe(":8080", cmux))
}
