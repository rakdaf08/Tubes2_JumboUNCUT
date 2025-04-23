// src/backend/main.go
package main

import (
	"fmt"
	"log"
	"time" // Import time untuk mengukur durasi (opsional)
)

func main() {
	log.Println("=== MEMULAI APLIKASI BACKEND (MODE PENGUJIAN ALGORITMA) ===")

	// 1. Muat Data Awal (dari data.go)
	dataDirPath := "data"
	err := InitData(dataDirPath)
	if err != nil {
		log.Fatalf("FATAL: Gagal memuat data awal dari '%s': %v", dataDirPath, err)
	}
	fmt.Println("Data awal berhasil dimuat.")

	// 2. Bangun Struktur Graf (dari graph.go)
	BuildGraph(GetRecipeMap())
	fmt.Println("Struktur graf siap digunakan.")

	// 3. Definisikan Elemen Target untuk Diuji
	testTargets := []string{
		"Mud",          // Sederhana
		"Steam",        // Sederhana
		"Clay",         // Sedikit lebih kompleks
		"Brick",        // Sedikit lebih kompleks
		"Life",         // Cukup kompleks
		"Human",        // Kompleks
		"Time",         // Sangat Kompleks (jika ada datanya)
		"Unobtanium",   // Elemen tidak ada (uji error)
		"Water",        // Elemen dasar (uji kasus dasar)
	}

	fmt.Println("\n=== MEMULAI PENGUJIAN BFS & DFS ===")

	// 4. Loop Melalui Setiap Target dan Uji Kedua Algoritma
	for _, target := range testTargets {
		fmt.Printf("\n--- Menguji Target: [%s] ---\n", target)

		// -- Uji BFS --
		fmt.Println("[BFS]")
		startTimeBFS := time.Now() // Ukur waktu mulai
		bfsPath, bfsNodesVisited, bfsErr := FindPathBFS(target)
		durationBFS := time.Since(startTimeBFS) // Ukur durasi

		if bfsErr != nil {
			fmt.Printf("  Error: %v\n", bfsErr)
			fmt.Printf("  Node Dikunjungi: %d\n", bfsNodesVisited)
		} else {
			fmt.Printf("  Jalur Ditemukan (%d langkah) - %d node dikunjungi - Durasi: %v\n", len(bfsPath), bfsNodesVisited, durationBFS)
			if len(bfsPath) == 0 && target != "Air" && target != "Earth" && target != "Fire" && target != "Water" {
				// Ini seharusnya tidak terjadi jika error nil, tapi sebagai jaga-jaga
				fmt.Println("  (Target ditemukan tapi path kosong?)")
			} else if len(bfsPath) == 0 {
				fmt.Println("  (Target adalah elemen dasar)")
			} else {
				// Cetak beberapa langkah awal/akhir saja agar tidak terlalu panjang
				maxStepsToShow := 5
				for i, step := range bfsPath {
					if i < maxStepsToShow || i >= len(bfsPath)-maxStepsToShow {
						fmt.Printf("    Langkah %d: %s + %s => %s\n", i+1, step.Ingredient1, step.Ingredient2, step.Result)
					} else if i == maxStepsToShow {
						fmt.Printf("    ...\n")
					}
				}
			}
		}
		fmt.Println("[Akhir BFS]")

		// Beri sedikit pemisah
		fmt.Println("---")

		// -- Uji DFS --
		fmt.Println("[DFS]")
		startTimeDFS := time.Now() // Ukur waktu mulai
		dfsPath, dfsNodesVisited, dfsErr := FindPathDFS(target)
		durationDFS := time.Since(startTimeDFS) // Ukur durasi

		if dfsErr != nil {
			fmt.Printf("  Error: %v\n", dfsErr)
			fmt.Printf("  Node Dikunjungi: %d\n", dfsNodesVisited)
		} else {
			fmt.Printf("  Jalur Ditemukan (%d langkah) - %d node dikunjungi - Durasi: %v\n", len(dfsPath), dfsNodesVisited, durationDFS)
			if len(dfsPath) == 0 && target != "Air" && target != "Earth" && target != "Fire" && target != "Water" {
				fmt.Println("  (Target ditemukan tapi path kosong?)")
			} else if len(dfsPath) == 0 {
				fmt.Println("  (Target adalah elemen dasar)")
			} else {
				// Cetak beberapa langkah awal/akhir saja agar tidak terlalu panjang
				maxStepsToShow := 5
				for i, step := range dfsPath {
					if i < maxStepsToShow || i >= len(dfsPath)-maxStepsToShow {
						fmt.Printf("    Langkah %d: %s + %s => %s\n", i+1, step.Ingredient1, step.Ingredient2, step.Result)
					} else if i == maxStepsToShow {
						fmt.Printf("    ...\n")
					}
				}
			}
		}
		fmt.Println("[Akhir DFS]")
		fmt.Printf("--- Selesai Target: [%s] ---\n", target)

	} // Akhir loop testTargets

	fmt.Println("\n=== PENGUJIAN SELESAI ===")

	// Komentari atau hapus bagian setup server untuk sementara
	// fmt.Println("Setup server HTTP (placeholder)...")
	// setupRoutes()
	// log.Println("Starting server on port 8080...")
	// err = http.ListenAndServe(":8080", nil)
	// if err != nil {
	// 	log.Fatalf("FATAL: Gagal menjalankan server: %v", err)
	// }
}

// Anda TIDAK perlu menyalin fungsi handler API atau setupRoutes ke sini
// karena kita hanya fokus pada pengujian algoritma.