package server

import (
	"net/http"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
	// HandlerFunc ServerHandlerFunc
	// HandlerFunc ServerHandlerFunc2
}

type Routes []Route

var routes = Routes{
	Route{
		"Getndex",
		"GET",
		"/",
		getIndex,
	},
	Route{
		"GetHeaders",
		"GET",
		"/headers",
		getHeaders,
	},
	Route{
		"PostEcho",
		"POST",
		"/echo",
		postEcho,
	},
	Route{
		"GetBorgs",
		"GET",
		"/borgs",
		getBorgs,
	},
	Route{
		"PostPing",
		"POST",
		"/ping",
		postPing,
	},
}
