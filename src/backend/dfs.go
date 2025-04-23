// src/backend/dfs.go
package main // Harus package main

import (
	"container/list" // Digunakan oleh reconstructPathRevised
	"errors"
	"fmt"
)

// --- Fungsi DFS Utama ---

// FindPathDFS mencari sebuah jalur resep dari elemen dasar ke targetElement menggunakan DFS.
// TIDAK menjamin jalur terpendek.
// Mengembalikan slice resep yang membentuk jalur, jumlah node yang dikunjungi, dan error jika tidak ditemukan.
func FindPathDFS(targetElement string) ([]Recipe, int, error) {
	fmt.Printf("Mencari jalur DFS ke: %s\n", targetElement)

	graph := GetAlchemyGraph()
	if graph == nil {
		return nil, 0, errors.New("graf alkimia belum diinisialisasi")
	}

	// 1. Inisialisasi struktur data DFS
	stack := make([]string, 0)                     // Stack untuk elemen yang akan dikunjungi (pakai slice)
	visited := make(map[string]bool)               // Set untuk melacak elemen yang sudah ditemukan/dikunjungi
	parent := make(map[string]Recipe)              // Map untuk merekonstruksi path
	nodesVisitedCount := 0                         // Penghitung node yang dieksplorasi dari stack

	baseElements := []string{"Air", "Earth", "Fire", "Water"}

	// 2. Mulai dari elemen dasar (masukkan ke stack)
	for _, base := range baseElements {
		if base == targetElement {
			return []Recipe{}, 0, nil // Target adalah elemen dasar
		}
		// Hanya tambahkan jika elemen dasar relevan/ada di data kita
		if _, exists := graph[base]; exists || len(GetRecipeMap()[base]) > 0 || isBaseElementDFS(base) {
			if !visited[base] { // Hanya push sekali jika belum visited
				stack = append(stack, base) // Push ke stack
				visited[base] = true
			}
		}
	}

	// 3. Loop DFS selama stack tidak kosong
	for len(stack) > 0 {
		// Pop elemen terakhir dari stack (LIFO)
		n := len(stack) - 1
		currentElement := stack[n]
		stack = stack[:n] // Hapus elemen terakhir
		nodesVisitedCount++

		// Debug: Cetak elemen yg sedang diproses
		// fmt.Printf("  DFS: Memproses -> %s\n", currentElement)

		// 4. Cek apakah ini target kita (cek setelah pop)
		//    Pindah cek ke sini agar path reconstruction bekerja benar jika target adalah salah satu start node yg relevan
		if currentElement == targetElement {
			fmt.Printf("DFS: Target '%s' ditemukan!\n", targetElement)
			path := reconstructPathRevisedDFS(parent, targetElement) // Gunakan fungsi rekonstruksi
			return path, nodesVisitedCount, nil                     // Sukses!
		}

		// 5. Dapatkan semua resep yang melibatkan currentElement sebagai bahan
		recipesInvolvingCurrent := graph[currentElement]

		// 6. Iterasi melalui resep-resep tersebut
		for _, recipe := range recipesInvolvingCurrent {
			var otherIngredient string
			if recipe.Ingredient1 == currentElement {
				otherIngredient = recipe.Ingredient2
			} else {
				otherIngredient = recipe.Ingredient1
			}

			// 7. Cek apakah bahan lainnya sudah ditemukan (visited)
			if visited[otherIngredient] {
				result := recipe.Result

				// 8. Cek apakah hasil ini belum pernah ditemukan sebelumnya
				if !visited[result] {
					// Debug: Cetak elemen baru yg ditemukan
					// fmt.Printf("    DFS: Menemukan -> %s (dari %s + %s)\n", result, recipe.Ingredient1, recipe.Ingredient2)

					visited[result] = true    // Tandai hasil sudah ditemukan
					parent[result] = recipe   // Simpan resep pembuatnya
					stack = append(stack, result) // Push hasil ke stack untuk dijelajahi lebih dulu (Depth-First)

					// TIDAK langsung return di sini seperti BFS, biarkan loop lanjut
					// sampai stack kosong atau target di-pop dari stack (di langkah 4)
				}
			}
		} // Akhir loop resep
	} // Akhir loop stack

	// 9. Jika stack kosong dan target tidak ditemukan
	fmt.Printf("DFS: Target '%s' tidak dapat ditemukan dari elemen dasar.\n", targetElement)
	return nil, nodesVisitedCount, fmt.Errorf("jalur ke elemen '%s' tidak ditemukan", targetElement)
}

// --- Fungsi Helper untuk Rekonstruksi Path (Sama dengan BFS) ---
// Kita bisa pindahkan ini ke file utilitas nanti agar tidak duplikat,
// tapi untuk sekarang kita copy saja dulu.

// reconstructPathRevisedDFS membuat ulang urutan resep dari map parent.
func reconstructPathRevisedDFS(parent map[string]Recipe, target string) []Recipe {
	path := list.New()
	queue := list.New()
	queue.PushBack(target)
	processedForPath := make(map[string]bool)
	processedForPath[target] = true

	for queue.Len() > 0 {
		queueEl := queue.Front()
		current := queueEl.Value.(string)
		queue.Remove(queueEl)

		recipe, exists := parent[current]
		if !exists {
			continue
		}

		path.PushFront(recipe)

		if _, processed := processedForPath[recipe.Ingredient1]; !processed {
			if _, parentExists := parent[recipe.Ingredient1]; parentExists || isBaseElementDFS(recipe.Ingredient1) {
				queue.PushBack(recipe.Ingredient1)
				processedForPath[recipe.Ingredient1] = true
			}
		}
		if _, processed := processedForPath[recipe.Ingredient2]; !processed {
			if _, parentExists := parent[recipe.Ingredient2]; parentExists || isBaseElementDFS(recipe.Ingredient2) {
				queue.PushBack(recipe.Ingredient2)
				processedForPath[recipe.Ingredient2] = true
			}
		}
	}

	finalPath := make([]Recipe, 0, path.Len())
	for e := path.Front(); e != nil; e = e.Next() {
		finalPath = append(finalPath, e.Value.(Recipe))
	}
	return finalPath
}

func isBaseElementDFS(name string) bool {
	baseElements := []string{"Air", "Earth", "Fire", "Water"}
	for _, base := range baseElements {
		if name == base {
			return true
		}
	}
	return false
}