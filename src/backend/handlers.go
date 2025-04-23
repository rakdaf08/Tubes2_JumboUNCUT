// src/backend/handlers.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	// "strconv" // Untuk konversi string ke integer (max recipes)
	"strings"
	"time"
)

// --- Struct untuk Response JSON ---

// SearchResponse adalah struktur data yang akan dikirim sebagai JSON response.
type SearchResponse struct {
	SearchTarget   string              `json:"searchTarget"`
	Algorithm      string              `json:"algorithm"`
	Mode           string              `json:"mode"`
	PathFound      bool                `json:"pathFound"`
	Path           []Recipe            `json:"path,omitempty"`           // Daftar resep (hanya ada jika pathFound=true)
	ImageURLs      map[string]string   `json:"imageURLs,omitempty"`      // URL gambar untuk elemen di path
	NodesVisited   int                 `json:"nodesVisited"`
	DurationMillis int64               `json:"durationMillis"`
	Error          string              `json:"error,omitempty"`          // Pesan error jika terjadi
}

// --- Fungsi Handler API ---

// searchHandler menangani request ke endpoint pencarian resep.
func searchHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS header (berguna saat development jika FE & BE beda port/domain)
	w.Header().Set("Access-Control-Allow-Origin", "*") // Izinkan akses dari mana saja
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Hanya izinkan metode GET
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// 1. Ambil Query Parameters
	targetElement := strings.TrimSpace(r.URL.Query().Get("target"))
	algo := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("algo")))
	mode := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mode")))
	// maxRecipesStr := r.URL.Query().Get("max") // Belum digunakan sekarang

	// Default values jika parameter kosong
	if algo == "" {
		algo = "bfs" // Default ke BFS
	}
	if mode == "" {
		mode = "shortest" // Default ke shortest
	}

	// // Konversi maxRecipes (belum dipakai sekarang)
	// maxRecipes := 0 // Default
	// if maxRecipesStr != "" {
	// 	maxRecipes, _ = strconv.Atoi(maxRecipesStr) // Abaikan error konversi untuk sementara
	// }

	// 2. Validasi Input Dasar
	if targetElement == "" {
		http.Error(w, "Parameter 'target' diperlukan", http.StatusBadRequest)
		return
	}

	// Gunakan fungsi dari data.go untuk cek elemen valid
	if !IsElementExists(targetElement) {
		http.Error(w, fmt.Sprintf("Elemen target '%s' tidak valid atau tidak ditemukan", targetElement), http.StatusBadRequest)
		return
	}

	if algo != "bfs" && algo != "dfs" {
		http.Error(w, "Parameter 'algo' harus 'bfs' atau 'dfs'", http.StatusBadRequest)
		return
	}

	// Validasi mode (sementara hanya shortest yg relevan)
	if mode != "shortest" && mode != "multiple" {
		http.Error(w, "Parameter 'mode' harus 'shortest' atau 'multiple'", http.StatusBadRequest)
		return
	}
	// Jika mode multiple tapi algo BFS (tidak cocok) atau fungsi multi belum ada
	if mode == "multiple" {
		// TODO: Panggil fungsi FindMultiplePaths nanti di sini
		log.Printf("Peringatan: Mode 'multiple' belum diimplementasikan, menjalankan mode single path DFS sebagai gantinya.")
		// Untuk sementara, kita bisa jalankan DFS single path jika mode multiple diminta
        algo = "dfs"
	}
    // Jika mode shortest tapi algo DFS (kurang optimal)
	if mode == "shortest" && algo == "dfs" {
        log.Printf("Peringatan: Mode 'shortest' diminta dengan algo 'dfs'. Hasil mungkin bukan yang terpendek.")
    }


	// 3. Panggil Fungsi Algoritma & Ukur Waktu
	startTime := time.Now()
	var path []Recipe
	var nodesVisited int
	var err error

	log.Printf("Memulai pencarian: Target=%s, Algo=%s, Mode=%s\n", targetElement, algo, mode)

	if algo == "bfs" {
		path, nodesVisited, err = FindPathBFS(targetElement)
	} else { // algo == "dfs"
		path, nodesVisited, err = FindPathDFS(targetElement)
	}

	duration := time.Since(startTime)
	log.Printf("Pencarian selesai: Durasi=%v, Nodes=%d, Error=%v\n", duration, nodesVisited, err)


	// 4. Siapkan Struktur Response
	response := SearchResponse{
		SearchTarget:   targetElement,
		Algorithm:      algo,
		Mode:           mode,
		PathFound:      err == nil && len(path) > 0, // Dianggap found jika tidak error & path tidak kosong (kecuali base element)
		Path:           path,
		NodesVisited:   nodesVisited,
		DurationMillis: duration.Milliseconds(),
		ImageURLs:      make(map[string]string), // Inisialisasi map gambar
	}

    // Handle kasus target adalah elemen dasar (path kosong tapi tidak error)
    if err == nil && len(path) == 0 && IsElementExists(targetElement) && isBaseElementDFS(targetElement) { // Gunakan helper isBaseElement
         response.PathFound = true // Tetap tandai found
    }


	// Jika ada error dari algoritma
	if err != nil {
		response.Error = err.Error()
		// Set status code? Bisa 404 jika errornya path not found
		// w.WriteHeader(http.StatusNotFound) // Mungkin? Atau biarkan 200 OK tapi dengan error di body?
	}

	// 5. Ambil URL Gambar untuk elemen di path (jika path ditemukan)
	if response.PathFound && len(path) > 0 {
		imgMap := GetImageMap() // Ambil map gambar dari data.go
		elementsInPath := make(map[string]bool) // Lacak elemen unik di path

		// Tambahkan elemen hasil akhir
		elementsInPath[targetElement] = true

		// Tambahkan semua bahan unik
		for _, step := range path {
			elementsInPath[step.Ingredient1] = true
			elementsInPath[step.Ingredient2] = true
			// Result perantara juga bisa ditambahkan jika perlu
			// elementsInPath[step.Result] = true
		}

		// Ambil URL untuk setiap elemen unik
		for elementName := range elementsInPath {
			if imgUrl, ok := imgMap[elementName]; ok && imgUrl != "" {
				response.ImageURLs[elementName] = imgUrl
			}
		}
        // Jangan lupa tambahkan juga URL gambar untuk elemen dasar jika belum ada
        baseElements := []string{"Air", "Earth", "Fire", "Water"}
        for _, base := range baseElements {
             if _, inPath := elementsInPath[base]; inPath { // Hanya jika elemen dasar relevan di path
                 if imgUrl, ok := imgMap[base]; ok && imgUrl != "" {
                     if _, exists := response.ImageURLs[base]; !exists {
                         response.ImageURLs[base] = imgUrl
                     }
                 }
             }
        }
	}


	// 6. Encode Response ke JSON dan Kirim
	w.Header().Set("Content-Type", "application/json")
	jsonResponse, jsonErr := json.MarshalIndent(response, "", "  ") // Pakai indentasi agar mudah dibaca saat debug
	if jsonErr != nil {
		log.Printf("Error saat marshal JSON response: %v", jsonErr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, writeErr := w.Write(jsonResponse)
	if writeErr != nil {
		log.Printf("Error saat menulis JSON response: %v", writeErr)
	}
}