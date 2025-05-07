// src/backend/dfs.go
package main

import (
	// "container/list" // Digunakan oleh reconstructPathRevised (jika FindPathDFS masih pakai)
	"errors"
	"fmt"
	"sort" // Diperlukan untuk generatePathIdentifier jika dipindah ke sini
	"strings" // Diperlukan untuk generatePathIdentifier jika dipindah ke sini
	"sync" // Import sync untuk Mutex dan WaitGroup
)

// FindPathDFS (fungsi DFS single path yang sudah ada)
// Anda bisa membiarkannya atau menghapusnya jika tidak dipakai lagi.
// Pastikan reconstructPathRevised bisa diakses jika fungsi ini masih ada.
func FindPathDFS(targetElement string) ([]Recipe, int, error) {
	fmt.Printf("Mencari jalur DFS (single) ke: %s\n", targetElement)

	graph := GetAlchemyGraph()
	if graph == nil {
		return nil, 0, errors.New("graf alkimia belum diinisialisasi")
	}

	stack := make([]string, 0)
	visited := make(map[string]bool)
	parent := make(map[string]Recipe) // Map parent untuk rekonstruksi
	nodesVisitedCount := 0

	baseElements := []string{"Air", "Earth", "Fire", "Water"}

	// Inisialisasi stack dengan elemen dasar
	for _, base := range baseElements {
		if base == targetElement {
			return []Recipe{}, 0, nil
		}
		if _, exists := graph[base]; exists || isBaseElementDFS(base) { // Gunakan isBaseElementDFS
			if !visited[base] {
				stack = append(stack, base)
				visited[base] = true
			}
		}
	}

	// Loop DFS Iteratif
	for len(stack) > 0 {
		n := len(stack) - 1
		currentElement := stack[n]
		stack = stack[:n] // Pop
		nodesVisitedCount++

		// Cek target *setelah* pop agar parent map benar
		if currentElement == targetElement {
			fmt.Printf("DFS (single): Target '%s' ditemukan!\n", targetElement)
			// Gunakan fungsi rekonstruksi yang sama dengan BFS (dari bfs.go)
			path := reconstructPathRevised(parent, targetElement)
			return path, nodesVisitedCount, nil
		}

		// Dapatkan resep yang melibatkan elemen ini sebagai *bahan*
		recipesInvolvingCurrent := graph[currentElement]

		for _, recipe := range recipesInvolvingCurrent {
			var otherIngredient string
			if recipe.Ingredient1 == currentElement {
				otherIngredient = recipe.Ingredient2
			} else {
				otherIngredient = recipe.Ingredient1
			}

			// Jika bahan lain sudah dikunjungi (bisa membuat hasil)
			if visited[otherIngredient] {
				result := recipe.Result
				// Jika hasil belum dikunjungi, tandai, simpan parent, push ke stack
				if !visited[result] {
					visited[result] = true
					parent[result] = recipe // Simpan resep pembuatnya
					stack = append(stack, result) // Push hasil (DFS)
				}
			}
		}
	}

	fmt.Printf("DFS (single): Target '%s' tidak dapat ditemukan.\n", targetElement)
	return nil, nodesVisitedCount, fmt.Errorf("jalur ke elemen '%s' tidak ditemukan", targetElement)
}


// --- Implementasi DFS Multiple Path ---

// FindMultiplePathsDFS mencari sejumlah 'maxRecipes' jalur resep berbeda ke targetElement menggunakan DFS.
// Menggunakan multithreading untuk memulai pencarian dari resep awal.
func FindMultiplePathsDFS(targetElement string, maxRecipes int) ([][]Recipe, int, error) {
	fmt.Printf("Mencari %d jalur DFS BERBEDA ke: %s (menggunakan multithreading)\n", maxRecipes, targetElement)

	// Akses data yang diperlukan
	recipeMap := GetRecipeMap() // Map[Hasil] -> []ResepPembuat
	if recipeMap == nil {
		return nil, 0, errors.New("map resep belum diinisialisasi")
	}
	if maxRecipes <= 0 {
		return nil, 0, errors.New("jumlah resep minimal harus 1")
	}
    if isBaseElementDFS(targetElement) {
        return [][]Recipe{}, 0, nil // Target adalah elemen dasar
    }

	// Shared variables
	var allFoundPaths [][]Recipe
	var mu sync.Mutex
	var wg sync.WaitGroup
	foundCount := 0
	nodesVisitedCount := -1 // Sulit dihitung akurat, set -1
    addedPathIdentifiers := make(map[string]bool) // Cek duplikasi

    // Channel untuk sinyal berhenti
    quitChan := make(chan struct{})
    var quitOnce sync.Once

	// Fungsi rekursif internal DFS
	var findPathRecursive func(elementToFind string, currentPathRecipes []Recipe, visitedOnBranch map[string]bool)
	findPathRecursive = func(elementToFind string, currentPathRecipes []Recipe, visitedOnBranch map[string]bool) {

        // 1. Cek sinyal berhenti
        select {
        case <-quitChan:
            return
        default: // Lanjutkan
        }

		// 2. Deteksi Siklus
		if visitedOnBranch[elementToFind] {
			return
		}

		// 3. Base Case: Elemen Dasar
		if isBaseElementDFS(elementToFind) {
			return // Tidak perlu cari resep lagi
		}

        // 4. Tandai kunjungan di cabang ini (buat salinan map)
        newVisited := make(map[string]bool)
        for k, v := range visitedOnBranch { newVisited[k] = v }
        newVisited[elementToFind] = true

		// 5. Dapatkan resep pembuat
		recipes := recipeMap[elementToFind]
		if len(recipes) == 0 { return /* Buntu */ }

		// 6. Jelajahi setiap resep
		for _, recipe := range recipes {
            // Tambahkan resep ini ke depan path saat ini
            pathIncludingCurrent := append([]Recipe{recipe}, currentPathRecipes...)

            // Cek apakah kedua bahan adalah elemen dasar
            ing1IsBase := isBaseElementDFS(recipe.Ingredient1)
            ing2IsBase := isBaseElementDFS(recipe.Ingredient2)

            if ing1IsBase && ing2IsBase {
                // --- Jalur Lengkap Ditemukan ---
                pathID := generatePathIdentifierDFS(pathIncludingCurrent) // Buat ID unik
                mu.Lock()
                if !addedPathIdentifiers[pathID] && foundCount < maxRecipes {
                    // Buat salinan path sebelum disimpan
                    completePath := make([]Recipe, len(pathIncludingCurrent))
                    copy(completePath, pathIncludingCurrent)
                    allFoundPaths = append(allFoundPaths, completePath)
                    addedPathIdentifiers[pathID] = true
                    foundCount++
                    fmt.Printf("DFS Multiple: Path UNIK ke-%d ditemukan (target: %s)\n", foundCount, targetElement)

                    if foundCount >= maxRecipes {
                        quitOnce.Do(func() { close(quitChan) }) // Kirim sinyal berhenti
                    }
                }
                mu.Unlock()
                // Lanjutkan cek resep lain (jangan return)

            } else {
                // --- Lanjutkan Rekursi ---
                // Rekursi untuk bahan 1 (jika bukan base)
                if !ing1IsBase {
                    findPathRecursive(recipe.Ingredient1, pathIncludingCurrent, newVisited)
                }

                // Cek sinyal berhenti sebelum rekursi kedua
                select {
                case <-quitChan: return
                default: // Lanjutkan
                    // Rekursi untuk bahan 2 (jika bukan base)
                    if !ing2IsBase {
                        findPathRecursive(recipe.Ingredient2, pathIncludingCurrent, newVisited)
                    }
                }
            }

             // Cek sinyal berhenti setelah selesai satu resep
             select {
             case <-quitChan: return
             default: // Lanjutkan
             }
		} // Akhir loop resep
	} // Akhir findPathRecursive

	// --- Mulai Pencarian Paralel ---
	initialRecipes := recipeMap[targetElement]
	if len(initialRecipes) == 0 {
		return nil, nodesVisitedCount, fmt.Errorf("tidak ada resep yang diketahui untuk membuat %s", targetElement)
	}

	fmt.Printf("Memulai %d pencarian DFS paralel awal untuk %s...\n", len(initialRecipes), targetElement)
	for _, recipe := range initialRecipes {
		wg.Add(1) // Tambah counter WaitGroup
		go func(startRecipe Recipe) {
			defer wg.Done() // Pastikan Done dipanggil

            initialPath := []Recipe{startRecipe} // Path awal hanya resep ini
            initialVisited := make(map[string]bool) // Visited map awal
            initialVisited[targetElement] = true // Tandai target agar tidak kembali

            ing1IsBase := isBaseElementDFS(startRecipe.Ingredient1)
            ing2IsBase := isBaseElementDFS(startRecipe.Ingredient2)

            // Jika resep awal sudah lengkap
            if ing1IsBase && ing2IsBase {
                 pathID := generatePathIdentifierDFS(initialPath)
                 mu.Lock()
                 if !addedPathIdentifiers[pathID] && foundCount < maxRecipes {
                     finalPath := make([]Recipe, len(initialPath))
                     copy(finalPath, initialPath)
                     allFoundPaths = append(allFoundPaths, finalPath)
                     addedPathIdentifiers[pathID] = true
                     foundCount++
                     fmt.Printf("DFS Multiple: Path UNIK ke-%d ditemukan (target: %s) - Initial Recipe\n", foundCount, targetElement)
                     if foundCount >= maxRecipes {
                          quitOnce.Do(func() { close(quitChan) })
                     }
                 }
                 mu.Unlock()
            } else {
                // Mulai rekursi untuk bahan non-dasar
                if !ing1IsBase {
                    findPathRecursive(startRecipe.Ingredient1, initialPath, initialVisited)
                }
                select { // Cek quit signal
                case <-quitChan: return
                default:
                     if !ing2IsBase {
                         findPathRecursive(startRecipe.Ingredient2, initialPath, initialVisited)
                     }
                }
            }
		}(recipe) // Jalankan goroutine dengan salinan resep
	}

	// Tunggu semua goroutine awal selesai
	wg.Wait()
	fmt.Println("Semua goroutine DFS awal selesai.")
    quitOnce.Do(func() { close(quitChan) }) // Pastikan channel ditutup

	// Cek hasil akhir
	if len(allFoundPaths) == 0 {
        if !isBaseElementDFS(targetElement) {
		    return nil, nodesVisitedCount, fmt.Errorf("jalur ke elemen '%s' tidak ditemukan", targetElement)
        }
	}

	// Pastikan tidak mengembalikan lebih dari maxRecipes
    finalPathsToReturn := allFoundPaths
    mu.Lock()
    if len(allFoundPaths) > maxRecipes {
        finalPathsToReturn = allFoundPaths[:maxRecipes]
    }
    mu.Unlock()

	return finalPathsToReturn, nodesVisitedCount, nil
}


// Fungsi isBaseElementDFS (pastikan ada dan bisa diakses)
func isBaseElementDFS(name string) bool {
	baseElements := []string{"Air", "Earth", "Fire", "Water"}
	for _, base := range baseElements {
		if name == base {
			return true
		}
	}
	return false
}

// Fungsi generatePathIdentifierDFS (mirip dengan versi BFS, bisa juga dipindah ke file util)
func generatePathIdentifierDFS(path []Recipe) string {
	recipesCopy := make([]Recipe, len(path))
	copy(recipesCopy, path)
	sort.Slice(recipesCopy, func(i, j int) bool {
		if recipesCopy[i].Result != recipesCopy[j].Result {
			return recipesCopy[i].Result < recipesCopy[j].Result
		}
        ing1i, ing2i := recipesCopy[i].Ingredient1, recipesCopy[i].Ingredient2
        if ing1i > ing2i { ing1i, ing2i = ing2i, ing1i }
        ing1j, ing2j := recipesCopy[j].Ingredient1, recipesCopy[j].Ingredient2
        if ing1j > ing2j { ing1j, ing2j = ing2j, ing1j }
		if ing1i != ing1j {
			return ing1i < ing1j
		}
		return ing2i < ing2j
	})
	var parts []string
	for _, r := range recipesCopy {
        ing1, ing2 := r.Ingredient1, r.Ingredient2
        if ing1 > ing2 { ing1, ing2 = ing2, ing1 }
		parts = append(parts, fmt.Sprintf("%s+%s=>%s", ing1, ing2, r.Result))
	}
	return strings.Join(parts, "|")
}

// Fungsi reconstructPathRevised (dari bfs.go) diperlukan jika FindPathDFS masih menggunakannya
// Pastikan bisa diakses dari package main
// func reconstructPathRevised(parent map[string]Recipe, target string) []Recipe { ... }

