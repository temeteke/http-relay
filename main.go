package main

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
)

var urlHtmlPattern = regexp.MustCompile(`(")(https?://[^\s"'<>\)]+|/[^\s"'<>\)]+)(")`)
var urlListPattern = regexp.MustCompile(`(?m)^(https?://[^\s"'<>\)]+|/[^\s"'<>\)]+)$`)

func handler(w http.ResponseWriter, r *http.Request) {
	urlPrefix := "http://" + r.Host + "/"

	target := strings.TrimPrefix(r.RequestURI, "/")
	if target == "" {
		http.Error(w, "Invalid target URL", http.StatusBadRequest)
		return
	}

	log.Println(target)

	targetURL, err := url.Parse(target)
	if err != nil || (targetURL.Scheme != "http" && targetURL.Scheme != "https") {
		http.Error(w, "Invalid target URL", http.StatusBadRequest)
		return
	}

	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	for name, values := range r.Header {
		canonicalName := http.CanonicalHeaderKey(name)
		if canonicalName == "Host" {
			continue
		}
		for _, value := range values {
			req.Header.Add(canonicalName, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to reach target", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	// Rewrite the response body if it is HTML or M3U8
	if contentType == "text/html" || contentType == "application/x-mpegurl" || contentType == "application/vnd.apple.mpegurl" {
		// Decompress the response body if it is compressed
		var bodyBytes []byte
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gzReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				http.Error(w, "Failed to create gzip reader", http.StatusInternalServerError)
				return
			}
			defer gzReader.Close()
			bodyBytes, err = io.ReadAll(gzReader)
			if err != nil {
				http.Error(w, "Failed to read gzip body", http.StatusInternalServerError)
				return
			}
			resp.Header.Del("Content-Encoding")
		} else {
			bodyBytes, err = io.ReadAll(resp.Body)
			if err != nil {
				http.Error(w, "Failed to read response body", http.StatusInternalServerError)
				return
			}
		}

		// Modify URLs in the response body
		modifiedBody := bodyBytes
		if contentType == "text/html" {
			modifiedBody = urlHtmlPattern.ReplaceAllFunc(bodyBytes, func(match []byte) []byte {
				submatch := urlHtmlPattern.FindStringSubmatch(string(match))
				parsedURL, err := url.Parse(submatch[2])
				if err != nil {
					return match
				}
				if !parsedURL.IsAbs() {
					parsedURL = targetURL.ResolveReference(parsedURL)
				}
				return []byte(submatch[1] + urlPrefix + parsedURL.String() + submatch[3])
			})
		} else if contentType == "application/x-mpegurl" || contentType == "application/vnd.apple.mpegurl" {
			modifiedBody = urlListPattern.ReplaceAllFunc(bodyBytes, func(match []byte) []byte {
				parsedURL, err := url.Parse(string(match))
				if err != nil {
					return match
				}
				if !parsedURL.IsAbs() {
					parsedURL = targetURL.ResolveReference(parsedURL)
				}
				return []byte(urlPrefix + parsedURL.String())
			})
		}

		// Write the modified response
		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}
		// Let http.ResponseWriter handle Content-Length automatically
		w.WriteHeader(resp.StatusCode)
		w.Write(modifiedBody)
	} else {
		// Write the response as is
		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

func main() {
	r := mux.NewRouter()
	// SkipClean prevents the router from cleaning the URL path, which is necessary for handling proxied URLs correctly.
	r.SkipClean(true)
	r.PathPrefix("/").HandlerFunc(handler)
	log.Fatal(http.ListenAndServe(":8080", r))
}
