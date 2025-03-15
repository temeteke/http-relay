package main

import (
	"compress/gzip"
	"fmt"
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
		lowerName := strings.ToLower(name)
		if lowerName == "host" {
			continue
		}
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to reach target", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var bodyBytes []byte
	encoding := resp.Header.Get("Content-Encoding")
	if encoding == "gzip" {
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
	} else {
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read response body", http.StatusInternalServerError)
			return
		}
	}

	// URL の書き換え（Content-Type チェック）
	modifiedBody := bodyBytes
	contentType := resp.Header.Get("Content-Type")
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
	} else if contentType == "application/x-mpegurl" || contentType == "application/vnd.apple.mpegur" {
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

	// ヘッダーを設定してレスポンスを返す
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(modifiedBody)))
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(modifiedBody)
}

func main() {
	r := mux.NewRouter()
	// SkipClean prevents the router from cleaning the URL path, which is necessary for handling proxied URLs correctly.
	r.SkipClean(true)
	r.PathPrefix("/").HandlerFunc(handler)
	log.Fatal(http.ListenAndServe(":8080", r))
}
