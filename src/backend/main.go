// src/backend/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	scrapeOnly := flag.Bool("scrapeonly", false, "Run scraping and filtering then exit")
	flag.Parse()

	RunScraping()
	runFilter()
	if *scrapeOnly {
		log.Println("Scraping dan filtering selesai (mode scrapeonly). Aplikasi akan keluar.")
		return
	}
	log.Println("=== MEMULAI SERVER BACKEND ===")
	dataDirPath := "data"
	err := InitData(dataDirPath)
	if err != nil {
		log.Fatalf("FATAL: Gagal memuat data awal aplikasi dari '%s': %v", dataDirPath, err)
	}
	fmt.Println("Data awal berhasil dimuat.")
	BuildGraph(GetRecipeMap())
	fmt.Println("Struktur graf siap digunakan.")

	// Setup Rute API
	http.HandleFunc("/api/search", searchHandler)

	// Jalankan Server
	port := "8080"
	log.Printf("Server backend berjalan di http://localhost:%s\n", port)
	log.Printf("Server frontend berjalan di http://localhost:3000\n")
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("FATAL: Gagal menjalankan server: %v", err)
	}
}
