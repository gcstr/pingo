package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

//go:embed templates/*
var templatesFS embed.FS

func startWebServer(db *sql.DB, port string) {
	tmpl := template.Must(template.ParseFS(templatesFS, "templates/index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, nil); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		startDate := r.URL.Query().Get("start")
		endDate := r.URL.Query().Get("end")

		var stats []PingStats
		var err error

		if startDate != "" && endDate != "" {
			stats, err = getStatsByDateRange(db, startDate, endDate)
		} else {
			stats, err = getRecentStats(db, 50)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	log.Printf("Web server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}
