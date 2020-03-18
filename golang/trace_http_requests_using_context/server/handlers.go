package server

import (
	"encoding/json"
	"fmt"
	"github.com/bboortz/goborg/internal/borg"
	"net/http"
)

type PingRequest struct {
	BorgId string `json:"borgid"`
	Addr   string `json:"addr"`
}

type Response struct {
	BorgId  string `json:"borgid"`
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

func getIndex(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Message: "Welcome!",
	}
	respondJSON(w, http.StatusOK, response)
}

func getHeaders(w http.ResponseWriter, r *http.Request) {
	for name, headers := range r.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func postEcho(w http.ResponseWriter, r *http.Request) {
	r.Write(w)
}

func getBorgs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(borg.BorgRepo); err != nil {
		panic(err)
	}
}

func postPing(w http.ResponseWriter, r *http.Request) {
	// retrieve the current context
	rqCtx := r.Context()

	// decoding json to struct
	var pingRequest PingRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&pingRequest); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// post verifying struct
	if pingRequest.BorgId == "" {
		respondError(w, http.StatusBadRequest, "borgId in passed json is missing")
		return
	}
	if pingRequest.Addr == "" {
		respondError(w, http.StatusBadRequest, "addr in passed json is missing")
		return
	}

	// storing the borg
	borg.RepoAddBorg(rqCtx, borg.NewBorg(pingRequest.BorgId, pingRequest.Addr))

	// responding
	response := Response{
		BorgId:  pingRequest.BorgId,
		Message: "pong",
	}
	respondJSON(w, http.StatusCreated, response)
}
