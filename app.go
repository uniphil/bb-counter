package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

//go:embed templates/*
var resources embed.FS

var t = template.Must(template.ParseFS(resources, "templates/*"))

var ONE_DAY, _ = time.ParseDuration("24h");

// var IS_PROD = os.Getenv("DEV") == ""



func inThreeishWeeks(from time.Time) time.Time {
	return from.Add(ONE_DAY * 7 * 4).Truncate(ONE_DAY * 7)
}

func beep(w http.ResponseWriter, r *http.Request) {
	site := strings.TrimPrefix(r.URL.Path, "/beep/")
	if site == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "400 site missing from path")
		return
	}

	rawOrigin := r.Header.Get("Origin")
	if rawOrigin == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "400 origin missing")
		return
	}
	origin, err := url.Parse(rawOrigin)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "400 bad origin")
		return
	}
	if origin.Host != site {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "400 wrong origin for site")
		log.Println(origin.Host, site)
		return
	}

	agent := r.UserAgent()
	if !strings.HasPrefix(agent, "Mozilla/5.0") &&
	   !strings.HasPrefix(agent, "Opera/") {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "400 seems non-browser :/")
		return
    }

    referer := r.Referer()
    if !strings.HasPrefix(referer, rawOrigin) {
    	w.WriteHeader(http.StatusBadRequest)
    	fmt.Fprintf(w, "400 seems wrong referer")
    	return
    }

	w.Header().Set("Access-Control-Allow-Origin", rawOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	if r.Method == "OPTIONS" {
		return
	}

	gpc := r.Header.Get("Sec-GPC") == "1" || r.FormValue("gpc") == "1"

	if gpc {
		log.Println("gpc")
		fmt.Fprintf(w, "who dis")
		return
	}

	path := r.FormValue("path")
	if !strings.HasPrefix(path, "/") {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "400 bad path")
		return
	}
	if len(path) > 80 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "400 path too long")
		return
	}

	cookie, no_cookie_err := r.Cookie("returning");
	returning := no_cookie_err == nil
	log.Println("path", path, "returning?", returning)

	if returning && cookie.Value != "true" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "400 bad cookie content")
		return
	}
	http.SetCookie(w, &http.Cookie {
		Name: "returning",
		Value: "true",
		Path: r.URL.Path,
		Expires: inThreeishWeeks(time.Now()),
		Secure: true,
		HttpOnly: true,
	})

	message := "welcome"
	if returning {
		message += " back"
	}
	fmt.Fprintf(w, message)
}


func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/beep/", beep)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]string{
			"Region": os.Getenv("FLY_REGION"),
		}
		t.ExecuteTemplate(w, "index.html.tmpl", data)
		log.Println("boooo")
	})

	log.Println("listening on", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
