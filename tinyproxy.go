package main

import (
    "flag"
    "io"
    "log"
    "net/http"
	"os"
	"strings"
	"net/url"
)

// Helper function to copy headers from one map to another
func copyHeaders(dst, src http.Header) {
    for k, vv := range src {
        for _, v := range vv {
            dst.Add(k, v)
        }
    }
}

func main() {
    port := flag.String("port", "", "Local port to accept HTTP connections")
    remote := flag.String("remote", "", "Remote host and port to forward requests to")
    outfile := flag.String("out", "", "Output file for logging (optional)")

    flag.Parse()

    if *port == "" || *remote == "" {
        log.Fatal("Port and remote must be specified.")
    }

    // Parse the remote scheme, host and port
    parsedRemote, err := url.Parse(*remote)
    if err != nil {
        log.Fatalf("Failed to parse remote URL: %v", err)
    }
    if parsedRemote.Scheme != "http" && parsedRemote.Scheme != "https" {
        log.Fatal("Remote URL must start with http:// or https://")
    }		
		
    var logOut io.Writer
    if *outfile != "" {
        file, err := os.Create(*outfile)
        if err != nil {
            log.Fatalf("Failed to open output file: %v", err)
        }
        defer file.Close()
        logOut = file
    } else {
        logOut = os.Stdout
    }

		logger := log.New(logOut, "", log.LstdFlags)

    handler := func(w http.ResponseWriter, r *http.Request) {
        // Log the incoming request
        logger.Printf("\n=== Incoming Request ===\n")
        logRequest(r, logger)

        // Forward the request to remote
        client := &http.Client{}
		req := cloneRequest(r)
        //targetURL := url.URL{
        //        Scheme:   parsedRemote.Scheme,
        //        Host:     parsedRemote.Host,
        //        Path:     r.URL.Path,
        //        RawQuery: r.URL.RawQuery,
			//    }
		req.URL.Scheme = parsedRemote.Scheme
		req.URL.Host = parsedRemote.Host

        resp, err := client.Do(req)
        if err != nil {
            logger.Printf("Error forwarding request: %v\n", err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
			defer resp.Body.Close()

		// Copy headers from the remote response to the client response
		copyHeaders(w.Header(), resp.Header)
			

        // Log the response from remote
        logger.Printf("\n=== Response from Remote ===\n")
        logResponse(resp, logger)

        // Copy response to client
        io.Copy(w, resp.Body)
    }

    server := &http.Server{
        Addr:    ":" + *port,
        Handler: http.HandlerFunc(handler),
    }

    logger.Printf("Starting proxy server on port %s, forwarding to %s...\n", *port, *remote)
    if err := server.ListenAndServe(); err != nil {
        logger.Fatalf("Failed to start server: %v\n", err)
    }
}

func cloneRequest(r *http.Request) *http.Request {
    return &http.Request{
        Method:     r.Method,
        URL:        r.URL,
        Proto:      r.Proto,
        Header:     r.Header.Clone(),
        Body:       r.Body, // Note: This is read-only once; in a real-world scenario, you may need to clone the body
        ContentLength: r.ContentLength,
        Close:      r.Close,
    }
}

func logRequest(r *http.Request, logger *log.Logger) {
    logger.Printf("Method: %s\n", r.Method)
    logger.Printf("Path: %s\n", r.URL.Path)
    logger.Printf("Headers:\n")
    for k, v := range r.Header {
        logger.Printf("- %s: %v\n", k, v)
    }
    if len(r.Form) > 0 {
        logger.Printf("Form Data:\n")
        for k, v := range r.Form {
            logger.Printf("- %s: %v\n", k, v)
        }
    }
}

func logResponse(resp *http.Response, logger *log.Logger) {
    logger.Printf("Status Code: %d\n", resp.StatusCode)
    logger.Printf("Headers:\n")
    for k, v := range resp.Header {
        logger.Printf("- %s: %v\n", k, v)
    }
    body, _ := io.ReadAll(resp.Body)
    logger.Printf("Body: %s\n", string(body))
    resp.Body = io.NopCloser(stringToReader(string(body)))
}


func stringToReader(s string) *strings.Reader {
    return strings.NewReader(s)
}
