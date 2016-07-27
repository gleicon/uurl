package main

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
)

type HTTPResources struct {
	uu                *UURL
	serverMux         *http.ServeMux
	staticFileHandler *http.Handler
}

func NewHTTPResources(uu *UURL, staticFiles string) *HTTPResources {
	sm := http.NewServeMux()
	sf := http.FileServer(http.Dir(staticFiles))
	hr := HTTPResources{uu: uu, serverMux: sm, staticFileHandler: &sf}

	hr.serverMux.HandleFunc("/", hr.RootHandler)
	hr.serverMux.HandleFunc("/api/v1/url", hr.URLApiHandler)
	hr.serverMux.HandleFunc("/api/v1/url/", hr.URLApiHandler)
	hr.serverMux.HandleFunc("/api/v1/quicklink", hr.QuickLinkHandler)
	hr.serverMux.HandleFunc("/api/v1/stats/", hr.StatsHandler)
	return &hr
}

func (hr *HTTPResources) RootHandler(w http.ResponseWriter, r *http.Request) {
	var redirURL string
	var err error
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	uid := r.URL.Path[len("/"):]
	ext := filepath.Ext(uid)
	// serve static file
	if len(uid) < 1 || ext != "" {
		f := *hr.staticFileHandler
		f.ServeHTTP(w, r)
		return
	}
	// serve redirect
	ref := r.Header.Get("REFERER")
	if redirURL, err = hr.uu.GetURLByUID(uid, r.RemoteAddr, ref); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if redirURL == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, redirURL, http.StatusSeeOther)
	return

}

// /api/v1/url GET/POST
func (hr *HTTPResources) URLApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	appendix := r.URL.Path[len("/api/v1/url"):]

	switch r.Method {
	case "GET":
		var redirURL string
		var err error

		if len(appendix) < 2 {
			http.Error(w, "Invalid parameter range, not enough data", http.StatusBadRequest)
			return
		}
		uid := appendix[1:]
		ref := r.Header.Get("REFERER")

		if redirURL, err = hr.uu.GetURLByUID(uid, r.RemoteAddr, ref); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if redirURL == "" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		http.Redirect(w, r, redirURL, http.StatusSeeOther)
		return
		break
	case "POST":
		var cc, uid string
		var customUrl, oUrl []string
		var err error

		if len(appendix) > 1 {
			http.Error(w, "Invalid parameter range ", http.StatusBadRequest)
			return
		}
		if err = r.ParseForm(); err != nil {
			http.Error(w, "Error parsing parameters", http.StatusInternalServerError)
			return
		}
		if oUrl = r.Form["url"]; oUrl == nil {
			http.Error(w, "No url found", http.StatusBadRequest)
			return
		}
		// the place to check if url is valid, if it's malware etc
		if _, err = url.Parse(oUrl[0]); err != nil {
			http.Error(w, "Invalid url", http.StatusBadRequest)
			return
		}
		if customUrl = r.Form["custom_url"]; customUrl != nil {
			cc = customUrl[0]
		}
		if uid, err = hr.uu.UpdateURLData(oUrl[0], cc); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if uid == "" {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// '{"uurl":"%s", "url": "%s", "base_url": "%s"}' % (uurl, url, BASE_URL)
		fmt.Fprintf(w, "{\"uurl\":\"%s\", \"url\": \"%s\", \"base_url\": \"%s\"}\n", uid, oUrl, r.Header.Get("HOST"))
		return
		break
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return
}

func (hr *HTTPResources) QuickLinkHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var uid string
		var oUrl string
		var err error

		if oUrl := r.FormValue("url"); oUrl == "" {
			http.Error(w, "No url found", http.StatusBadRequest)
			return
		}
		// the place to check if url is valid, if it's malware etc
		if _, err = url.Parse(oUrl); err != nil {
			http.Error(w, "Invalid url", http.StatusBadRequest)
			return
		}
		if uid, err = hr.uu.UpdateURLData(oUrl, ""); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if uid == "" {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// redirect to stats page
		return
		break
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return
}

func (hr *HTTPResources) StatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	uid := r.URL.Path[len("/api/v1/stats/"):]

	switch r.Method {
	case "GET":
		var sts *URLStats
		var err error
		var pl []byte

		if len(uid) < 1 {
			http.Error(w, "Invalid url", http.StatusBadRequest)
			return
		}
		if sts, err = hr.uu.GetURLStatsByUID(uid); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if pl, err = sts.toJson(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(pl)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return
}
