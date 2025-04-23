// src/backend/handlers.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// MultiSearchResponse struct (as defined before)
type MultiSearchResponse struct {
	SearchTarget   string              `json:"searchTarget"`
	Algorithm      string              `json:"algorithm"`
	Mode           string              `json:"mode"`
	MaxRecipes     int                 `json:"maxRecipes,omitempty"`
	PathFound      bool                `json:"pathFound"`
	Path           []Recipe            `json:"path,omitempty"` // For shortest mode
	Paths          [][]Recipe          `json:"paths,omitempty"` // For multiple mode
	ImageURLs      map[string]string   `json:"imageURLs,omitempty"`
	NodesVisited   int                 `json:"nodesVisited"`
	DurationMillis int64               `json:"durationMillis"`
	Error          string              `json:"error,omitempty"`
}


func searchHandler(w http.ResponseWriter, r *http.Request) {
	// ... (CORS headers, Method check) ...
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method != http.MethodGet {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// ... (Parsing target, algo, mode, maxRecipesStr) ...
	targetElement := strings.TrimSpace(r.URL.Query().Get("target"))
	algo := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("algo")))
	mode := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mode")))
	maxRecipesStr := r.URL.Query().Get("max")

	// ... (Default values) ...
	if algo == "" { algo = "bfs" }
	if mode == "" { mode = "shortest" }

	// ... (Validasi targetElement, algo, mode) ...
    if targetElement == "" { http.Error(w, "Parameter 'target' diperlukan", http.StatusBadRequest); return }
    if !IsElementExists(targetElement) { http.Error(w, fmt.Sprintf("Elemen target '%s' tidak valid", targetElement), http.StatusBadRequest); return }
    if algo != "bfs" && algo != "dfs" { http.Error(w, "Parameter 'algo' harus 'bfs' atau 'dfs'", http.StatusBadRequest); return }
    if mode != "shortest" && mode != "multiple" { http.Error(w, "Parameter 'mode' harus 'shortest' atau 'multiple'", http.StatusBadRequest); return }

	// ... (Proses parameter 'max' jika mode 'multiple') ...
    maxRecipes := 1
	if mode == "multiple" {
		if maxRecipesStr != "" {
			var convErr error
			maxRecipes, convErr = strconv.Atoi(maxRecipesStr)
			if convErr != nil || maxRecipes <= 0 {
				http.Error(w, "Parameter 'max' harus angka positif untuk mode 'multiple'", http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "Parameter 'max' diperlukan untuk mode 'multiple'", http.StatusBadRequest)
			return
		}
	}


	// Panggil Fungsi Algoritma & Ukur Waktu
	startTime := time.Now()

	var singlePath []Recipe
	var multiplePaths [][]Recipe
	var nodesVisited int
	var err error
	var pathFound bool

	log.Printf("Memulai pencarian: Target=%s, Algo=%s, Mode=%s, Max=%d\n", targetElement, algo, mode, maxRecipes)

	if mode == "shortest" {
		// --- Mode Shortest Path ---
		if algo == "bfs" {
			singlePath, nodesVisited, err = FindPathBFS(targetElement)
		} else { // algo == "dfs"
			log.Printf("Peringatan: Mode 'shortest' dengan algo 'dfs' belum diimplementasikan sepenuhnya. Menjalankan BFS.")
			singlePath, nodesVisited, err = FindPathDFS(targetElement) // Fallback ke BFS
		}
		pathFound = err == nil && (len(singlePath) > 0 || (len(singlePath) == 0 && isBaseElement(targetElement))) // Cek juga base element

	} else {
		// --- Mode Multiple Paths ---
		if algo == "bfs" {
			multiplePaths, nodesVisited, err = FindMultiplePathsBFS(targetElement, maxRecipes)
		} else { // algo == "dfs"
			// --- PERBAIKAN FALLBACK DI SINI ---
			log.Printf("Peringatan: Mode 'multiple' dengan algo 'dfs' belum diimplementasikan. Menjalankan mode single path DFS sebagai gantinya.")
			// Jalankan DFS standar
			singlePathFallback, nodesVisitedFallback, errFallback := FindPathDFS(targetElement)
			err = errFallback // Gunakan error dari fallback
			nodesVisited = nodesVisitedFallback // Gunakan nodes visited dari fallback

			// Jika DFS standar berhasil menemukan path, format hasilnya sebagai multiple path (dengan 1 elemen)
			if err == nil && len(singlePathFallback) > 0 {
				multiplePaths = [][]Recipe{singlePathFallback} // Bungkus single path dalam slice of slice
			} else {
                // Jika DFS standar gagal atau target adalah base element, multiplePaths tetap kosong
                multiplePaths = [][]Recipe{}
            }
            // ------------------------------------
		}
		pathFound = err == nil && len(multiplePaths) > 0 // Path found jika tidak ada error dan ada setidaknya 1 path
        // Handle kasus target adalah base element (tidak akan masuk ke multiplePaths dari fallback DFS)
        if err == nil && len(multiplePaths) == 0 && isBaseElement(targetElement) {
             pathFound = true // Tetap anggap found, tapi paths akan kosong
        }
	}

	duration := time.Since(startTime)
	log.Printf("Pencarian selesai: Durasi=%v, Nodes=%d, Ditemukan=%t, Error=%v\n", duration, nodesVisited, pathFound, err)

	// Siapkan Struktur Response
	response := MultiSearchResponse{
		SearchTarget:   targetElement,
		Algorithm:      algo,
		Mode:           mode,
		PathFound:      pathFound,
		NodesVisited:   nodesVisited,
		DurationMillis: duration.Milliseconds(),
		ImageURLs:      make(map[string]string),
	}

	// Isi field path/paths berdasarkan mode
	if mode == "shortest" {
		response.Path = singlePath
	} else {
		response.Paths = multiplePaths // Sekarang ini akan diisi bahkan saat fallback DFS
		response.MaxRecipes = maxRecipes
	}

	if err != nil {
		response.Error = err.Error()
	}

	// Ambil URL Gambar (logika ini seharusnya sudah benar untuk multiplePaths)
	if response.PathFound {
        imgMap := GetImageMap()
        elementsInPaths := make(map[string]bool)
        elementsInPaths[targetElement] = true

        pathsToProcess := [][]Recipe{} // Kumpulkan semua path untuk diproses
        if mode == "shortest" {
            if len(singlePath) > 0 {
                pathsToProcess = append(pathsToProcess, singlePath)
            }
        } else { // mode == "multiple"
            pathsToProcess = multiplePaths // Gunakan hasil dari multiplePaths
        }

        for _, path := range pathsToProcess {
            for _, step := range path {
                elementsInPaths[step.Ingredient1] = true
                elementsInPaths[step.Ingredient2] = true
            }
        }

		// Tambahkan elemen dasar yang mungkin relevan
		baseElements := []string{"Air", "Earth", "Fire", "Water"}
		for _, base := range baseElements {
			if _, exists := imgMap[base]; exists {
				if (isBaseElement(targetElement) && targetElement == base) || elementsInPaths[base] {
					elementsInPaths[base] = true
				}
			}
		}

		// Ambil URL untuk setiap elemen unik
		for elementName := range elementsInPaths {
			if imgUrl, ok := imgMap[elementName]; ok && imgUrl != "" {
				response.ImageURLs[elementName] = imgUrl
			}
		}
	}

	// Encode Response ke JSON dan Kirim
	w.Header().Set("Content-Type", "application/json")
	jsonResponse, jsonErr := json.MarshalIndent(response, "", "  ")
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

// // Pastikan fungsi isBaseElement ada dan bisa diakses
// func isBaseElement(name string) bool {
// 	baseElements := []string{"Air", "Earth", "Fire", "Water"}
// 	for _, base := range baseElements {
// 		if name == base {
// 			return true
// 		}
// 	}
// 	return false
// }

