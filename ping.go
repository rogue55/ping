package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/parkr/ping/analytics"
	"github.com/zenazn/goji"
)

const returnedJavaScript = "(function(){})();"
const lengthOfJavaScript = "17"

func javascriptRespond(w http.ResponseWriter, code int, err string) {
	w.WriteHeader(code)

	var content string
	if err == "" {
		content = returnedJavaScript
	} else {
		content = fmt.Sprintf(`(function(){console.error("%s")})();`, err)
	}

	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	fmt.Fprintf(w, content)
}

func ping(w http.ResponseWriter, r *http.Request) {
	referrer := r.Referer()
	if referrer == "" {
		log.Println("empty referrer")
		javascriptRespond(w, http.StatusBadRequest, "empty referrer")
		return
	}

	url, err := url.Parse(referrer)

	if err != nil {
		log.Println("invalid referrer:", referrer)
		javascriptRespond(w, 500, "Couldn't parse referrer: "+err.Error())
		return
	}

	var ip string
	if res := r.Header.Get("X-Forwarded-For"); res != "" {
		log.Println("Fetching IP from proxy:", res)
		ip = res
	} else {
		ip = r.RemoteAddr
	}

	visit := &Visit{
		IP:        ip,
		Host:      url.Host,
		Path:      url.Path,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	log.Println("Logging visit:", visit.String())

	err = visit.Save()

	if err != nil {
		javascriptRespond(w, 500, err.Error())
		return
	}

	javascriptRespond(w, 201, "")
}

func counts(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")

	var err error
	var views, visitors int
	if path == "" {
		http.Error(w, "Missing param", 400)
	} else {
		views, err = analytics.ViewsForPath(db, path)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		visitors, err = analytics.VisitorsForPath(db, path)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		writeJsonResponse(w, map[string]int{
			"views":    views,
			"visitors": visitors,
		})
	}
}

func all(w http.ResponseWriter, r *http.Request) {
	thing := r.FormValue("type")

	if thing == "path" || thing == "host" {
		entries, err := analytics.ListDistinctColumn(db, thing)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		writeJsonResponse(w, map[string][]string{"entries": entries})
	} else {
		http.Error(w, "Missing param", 400)
		return
	}
}

func main() {
	goji.Get("/ping", ping)
	goji.Get("/ping.js", ping)
	goji.Get("/counts", counts)
	goji.Get("/all", all)
	goji.Serve()
}
