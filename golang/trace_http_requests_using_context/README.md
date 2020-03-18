# Trace HTTP requests using context



# Starting Position

You wrote a REST API in go which is processing thousands of requests in one seconds. After a while you have recognized errors in the log which you want to fix. This errors are thrown by a methods deep in your code. In order to isolate the issue you have to trace the requests.
So that the question is which request is creating this issue. For sure you can dump all requests and try to correlate the timestamps of the incoming requests with the ocurred errors but this seems to be a hugh effort.



# Some Example Code

Your code could look like this.
Lets assume we know the error will be thrown always in our [http.HandlerFunc](https://golang.org/src/net/http/server.go?s=75533:75604#L2441) during executing the method *doSomeMagic* but you cannot find the cause in your big code base.

```
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
	// decoding json to struct
	var someObj SomeObj
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&someObj); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()

	// Do some ...
	doSomeMagic(someObj)

	// respond
	response := Response{
		Message: "Ok",
	}
    respondJSON(w, http.StatusOK, response)
}

func doSomeMagic(magicObj SomeObj) {
	// magically processing ....
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/post", handlePostRequest)

	log.Println("Start server on port :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
```



# Basic Idea

Adding log output everywhere inside the method *doSomeMagic* is not the best idea. Esspecially if the problem only occurs in production.
We could track the request ids for all incoming request and errors are containing this request ids too. So that we need to somehow pass the request id to the method.



# Context Package

The [context package](https://golang.org/pkg/context/) is defined as follows:

```
Package context defines the Context type, which carries deadlines, cancellation signals, and other request-scoped values across API boundaries and between processes.

Incoming requests to a server should create a Context, and outgoing calls to servers should accept a Context. The chain of function calls between them must propagate the Context, optionally replacing it with a derived Context created using WithCancel, WithDeadline, WithTimeout, or WithValue. When a Context is canceled, all Contexts derived from it are also canceled.
```

Sounds great.


# The approach

To track the request ids I have created a new package called *appcontext*. Using this package we are able to track requests using request ids. You can extend it for your own purposes if you like e.g. tracking sessions, cookies, etc.



## appcontext.go

This is a snipped of the package appcontext. This packace is in charge to provide the necessary context for the application plus logging. 
The full package can be found in the repository [github.com/bboortz/goborg/pkg/appcontext](https://github.com/bboortz/goborg/tree/master/pkg/appcontext)

```
package appcontext

import (
        "context"
)

type correlationIdType int

const (
        requestIdKey correlationIdType = iota
        sessionIdKey
        pkgNameKey
)

// WithRqId returns a context which knows its request ID
func WithRqId(ctx context.Context, requestId string) context.Context {
        return context.WithValue(ctx, requestIdKey, requestId)
}

// Logger returns a zap logger with as much context as possible
func Logger(ctx context.Context) *zap.Logger {
	newLogger := logger
	if ctx != nil {
		if ctxRqId, ok := ctx.Value(requestIdKey).(string); ok {
			newLogger = newLogger.With(zap.String("rqId", ctxRqId))
		}
		if ctxSessionId, ok := ctx.Value(sessionIdKey).(string); ok {
			newLogger = newLogger.With(zap.String("sessionId", ctxSessionId))
		}
		if ctxPkgName, ok := ctx.Value(pkgNameKey).(string); ok {
			newLogger = newLogger.With(zap.String("pkgName", ctxPkgName))
		}
	}
	return newLogger
}
```



## middleware.go

You will also need a middleware like this in order to assign centrally the request id. Our request id is just a random generated uuid.

```
package server

import (
        "context"
        "github.com/bboortz/goborg/pkg/appcontext"
        "github.com/google/uuid"
        "net/http"
)

func ContextMiddleware(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rqId, _ := uuid.NewRandom()
		rqCtx := appcontext.WithRqId(r.Context(), rqId.String())
		r = r.WithContext(rqCtx)

		inner.ServeHTTP(w, r)
	})
}
```



## Your Upated Code

To ensure you are using the code, you have to add the middleware method to you request handler.

```
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/post", handlePostRequest)

	log.Println("Start server on port :8080")
	cmux = ContextMiddleware(mux)
	log.Fatal(http.ListenAndServe(":8080", cmux))
}
```

We have to add the context to the method *handPostRequest* and are able to pass it to every other method.
Using `r.Context()` we are receiving the context from the http.Request.


```
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
```


Now we can also refine the method *doSomeMagic* and log the requests during processing

```
func doSomeMagic(ctx context.Context, magicObj SomeObj) {
	logger := appcontext.Logger(ctx)
	logger.Info("WooHooo")
}
```



## Example Log Output 

The log output is including the request id and will look like this. In case you need a deep tracing in your code base just pass the context to the necessary methods and enable the logging. You can also add the context to all errors.

```
2020-03-18T11:33:30.370+0100	INFO	trace_http_requests_using_context/main.go:71	Start server on port :8080
2020-03-18T11:33:36.600+0100	INFO	trace_http_requests_using_context/main.go:63	WooHooo	{"rqId": "42ab6373-ea58-4f86-9a4f-21f1bdea7d5d"}
2020-03-18T11:33:37.084+0100	INFO	trace_http_requests_using_context/main.go:63	WooHooo	{"rqId": "a203736b-4b8c-49e4-8af8-34b3742468d5"}
2020-03-18T11:33:37.534+0100	INFO	trace_http_requests_using_context/main.go:63	WooHooo	{"rqId": "2c48fe39-e28c-412c-9e75-3b13af5af2fd"}
```



# Conclusion

We made some learnings:
* We have learned how to use the *context* package in golang.
* We have learned how to track request ids from incoming request in your methods and can use this information for tracing issues in a bigger code base.

Please also watch the source code in my tipsntricks repository on github: [https://github.com/bboortz/tipsntricks/blob/master/go/trace_http_requests_using_context].



```
# Links

These links if have used for my own studies:

* [Using Go's 'context' library for making your logs make sense](https://blog.gopheracademy.com/advent-2016/context-logging/)
* [Context propagation over HTTP in Go](https://medium.com/@rakyll/context-propagation-over-http-in-go-d4540996e9b0)
* [How to pass context in golang request to middleware -stackoverflow](https://stackoverflow.com/questions/39946583/how-to-pass-context-in-golang-request-to-middleware)
* [Simple Golang HTTP Request Context Example](https://gocodecloud.com/blog/2016/11/15/simple-golang-http-request-context-example/)
* [package context - Gorilla web toolkit](http://www.gorillatoolkit.org/pkg/context)
