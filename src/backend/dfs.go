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
// func FindPathDFS(targetElement string) ([]Recipe, int, error) {
// 	fmt.Printf("Mencari jalur DFS (single) ke: %s\n", targetElement)

// 	graph := GetAlchemyGraph()
// 	if graph == nil {
// 		return nil, 0, errors.New("graf alkimia belum diinisialisasi")
// 	}

// 	stack := make([]string, 0)
// 	visited := make(map[string]bool)
// 	parent := make(map[string]Recipe) // Map parent untuk rekonstruksi
// 	nodesVisitedCount := 0

// 	baseElements := []string{"Air", "Earth", "Fire", "Water"}

// 	// Inisialisasi stack dengan elemen dasar
// 	for _, base := range baseElements {
// 		if base == targetElement {
// 			return []Recipe{}, 0, nil
// 		}
// 		if _, exists := graph[base]; exists || isBaseElementDFS(base) { // Gunakan isBaseElementDFS
// 			if !visited[base] {
// 				stack = append(stack, base)
// 				visited[base] = true
// 			}
// 		}
// 	}

// 	// Loop DFS Iteratif
// 	for len(stack) > 0 {
// 		n := len(stack) - 1
// 		currentElement := stack[n]
// 		stack = stack[:n] // Pop
// 		nodesVisitedCount++

// 		// Cek target *setelah* pop agar parent map benar
// 		if currentElement == targetElement {
// 			fmt.Printf("DFS (single): Target '%s' ditemukan!\n", targetElement)
// 			// Gunakan fungsi rekonstruksi yang sama dengan BFS (dari bfs.go)
// 			path := reconstructPathRevised(parent, targetElement)
// 			return path, nodesVisitedCount, nil
// 		}

// 		// Dapatkan resep yang melibatkan elemen ini sebagai *bahan*
// 		recipesInvolvingCurrent := graph[currentElement]

// 		for _, recipe := range recipesInvolvingCurrent {
// 			var otherIngredient string
// 			if recipe.Ingredient1 == currentElement {
// 				otherIngredient = recipe.Ingredient2
// 			} else {
// 				otherIngredient = recipe.Ingredient1
// 			}

// 			// Jika bahan lain sudah dikunjungi (bisa membuat hasil)
// 			if visited[otherIngredient] {
// 				result := recipe.Result
// 				// Jika hasil belum dikunjungi, tandai, simpan parent, push ke stack
// 				if !visited[result] {
// 					visited[result] = true
// 					parent[result] = recipe // Simpan resep pembuatnya
// 					stack = append(stack, result) // Push hasil (DFS)
// 				}
// 			}
// 		}
// 	}

// 	fmt.Printf("DFS (single): Target '%s' tidak dapat ditemukan.\n", targetElement)
// 	return nil, nodesVisitedCount, fmt.Errorf("jalur ke elemen '%s' tidak ditemukan", targetElement)
// }

func FindPathDFS(targetElement string) ([]Recipe, int, error) {
    fmt.Printf("Mencari jalur DFS (single) ke: %s\n", targetElement)

    // Persiapan
    recipeMap := GetRecipeMap()
    if recipeMap == nil {
        return nil, 0, errors.New("map resep belum diinisialisasi")
    }
    
    if isBaseElementDFS(targetElement) {
        return []Recipe{}, 0, nil // Target adalah elemen dasar
    }
    
    nodesVisitedCount := 0
    
    // Cache untuk elemen yang bisa dibuat
    knownCreatableElements := make(map[string]bool)
    for _, base := range []string{"Air", "Earth", "Fire", "Water"} {
        knownCreatableElements[base] = true
    }
    
    // Fungsi untuk memeriksa apakah elemen bisa dibuat
    var isCreatable func(element string, visited map[string]bool) bool
    isCreatable = func(element string, visited map[string]bool) bool {
        nodesVisitedCount++
        
        // Base case 1: Jika elemen dasar
        if isBaseElementDFS(element) {
            return true
        }
        
        // Base case 2: Sudah kita ketahui bisa dibuat
        if known, exists := knownCreatableElements[element]; exists {
            return known
        }
        
        // Base case 3: Loop deteksi
        if visited[element] {
            return false
        }
        
        // Tandai sudah dikunjungi untuk cabang ini
        newVisited := make(map[string]bool)
        for k, v := range visited {
            newVisited[k] = v
        }
        newVisited[element] = true
        
        // Cek setiap resep yang menghasilkan elemen ini
        recipes := recipeMap[element]
        if len(recipes) == 0 {
            // Simpan hasil: tidak bisa dibuat
            knownCreatableElements[element] = false
            return false
        }
        
        // Cek apakah ada resep yang valid (kedua bahannya bisa dibuat)
        for _, recipe := range recipes {
            ing1Creatable := isCreatable(recipe.Ingredient1, newVisited)
            ing2Creatable := isCreatable(recipe.Ingredient2, newVisited)
            
            if ing1Creatable && ing2Creatable {
                // Simpan hasil: bisa dibuat
                knownCreatableElements[element] = true
                return true
            }
        }
        
        // Simpan hasil: tidak bisa dibuat
        knownCreatableElements[element] = false
        return false
    }
    
    // Fungsi untuk menemukan jalur terpendek ke suatu elemen
    var findShortestPath func(target string, visited map[string]bool) []Recipe
    findShortestPath = func(target string, visited map[string]bool) []Recipe {
        nodesVisitedCount++
        
        // Base case 1: Elemen dasar
        if isBaseElementDFS(target) {
            return []Recipe{}
        }
        
        // Base case 2: Loop deteksi
        if visited[target] {
            return nil
        }
        
        // Tandai dikunjungi
        newVisited := make(map[string]bool)
        for k, v := range visited {
            newVisited[k] = v
        }
        newVisited[target] = true
        
        var shortestPathSoFar []Recipe
        
        // Cari resep untuk target
        recipes := recipeMap[target]
        for _, recipe := range recipes {
            // Periksa apakah bahan-bahan bisa dibuat
            emptyVisited := make(map[string]bool)
            if !isCreatable(recipe.Ingredient1, emptyVisited) || !isCreatable(recipe.Ingredient2, emptyVisited) {
                continue // Lewati resep yang tidak valid
            }
            
            // Cari path untuk kedua bahan
            path1 := []Recipe{}
            if !isBaseElementDFS(recipe.Ingredient1) {
                path1 = findShortestPath(recipe.Ingredient1, newVisited)
                if path1 == nil {
                    continue // Tidak bisa membuat bahan 1
                }
            }
            
            path2 := []Recipe{}
            if !isBaseElementDFS(recipe.Ingredient2) {
                path2 = findShortestPath(recipe.Ingredient2, newVisited)
                if path2 == nil {
                    continue // Tidak bisa membuat bahan 2
                }
            }
            
            // Gabungkan path
            currentPath := []Recipe{recipe}
            currentPath = append(currentPath, path1...)
            currentPath = append(currentPath, path2...)
            
            // Perbarui jalur terpendek jika perlu
            if shortestPathSoFar == nil || len(currentPath) < len(shortestPathSoFar) {
                shortestPathSoFar = currentPath
            }
        }
        
        return shortestPathSoFar
    }
    
	//Temukan jalur terpendek
    fmt.Printf("Mencari jalur terpendek untuk %s...\n", targetElement)
    shortestPath := findShortestPath(targetElement, make(map[string]bool))
    
    if shortestPath == nil {
        return nil, nodesVisitedCount, fmt.Errorf("tidak ada jalur valid untuk membuat %s", targetElement)
    }
    
    // PERUBAHAN: Balik urutan jalur sebelum menampilkan & mengembalikan
    reversedPath := make([]Recipe, len(shortestPath))
    for i, recipe := range shortestPath {
        reversedPath[len(shortestPath)-1-i] = recipe
    }
    
    // Debug - tampilkan jalur yang ditemukan dengan urutan terbalik
    fmt.Printf("Jalur DFS (single) (panjang: %d):\n", len(reversedPath))
    for i, recipe := range reversedPath {
        fmt.Printf("  Langkah %d: %s + %s => %s\n", 
                  i+1, recipe.Ingredient1, recipe.Ingredient2, recipe.Result)
    }
    
    return reversedPath, nodesVisitedCount, nil
}


// --- Implementasi DFS Multiple Path ---

// FindMultiplePathsDFS mencari sejumlah 'maxRecipes' jalur resep berbeda ke targetElement menggunakan DFS.
// Menggunakan multithreading untuk memulai pencarian dari resep awal.
// func FindMultiplePathsDFS(targetElement string, maxRecipes int) ([][]Recipe, int, error) {
// 	fmt.Printf("Mencari %d jalur DFS BERBEDA ke: %s (menggunakan multithreading)\n", maxRecipes, targetElement)

// 	// Akses data yang diperlukan
// 	recipeMap := GetRecipeMap() // Map[Hasil] -> []ResepPembuat
// 	if recipeMap == nil {
// 		return nil, 0, errors.New("map resep belum diinisialisasi")
// 	}
// 	if maxRecipes <= 0 {
// 		return nil, 0, errors.New("jumlah resep minimal harus 1")
// 	}
//     if isBaseElementDFS(targetElement) {
//         return [][]Recipe{}, 0, nil // Target adalah elemen dasar
//     }

// 	// Shared variables
// 	var allFoundPaths [][]Recipe
// 	var mu sync.Mutex
// 	var wg sync.WaitGroup
// 	foundCount := 0
// 	nodesVisitedCount := -1 // Sulit dihitung akurat, set -1
//     addedPathIdentifiers := make(map[string]bool) // Cek duplikasi

//     // Channel untuk sinyal berhenti
//     quitChan := make(chan struct{})
//     var quitOnce sync.Once

// 	// Fungsi rekursif internal DFS
// 	var findPathRecursive func(elementToFind string, currentPathRecipes []Recipe, visitedOnBranch map[string]bool)
// 	findPathRecursive = func(elementToFind string, currentPathRecipes []Recipe, visitedOnBranch map[string]bool) {

//         // 1. Cek sinyal berhenti
//         select {
//         case <-quitChan:
//             return
//         default: // Lanjutkan
//         }

// 		// 2. Deteksi Siklus
// 		if visitedOnBranch[elementToFind] {
// 			return
// 		}

// 		// 3. Base Case: Elemen Dasar
// 		if isBaseElementDFS(elementToFind) {
// 			return // Tidak perlu cari resep lagi
// 		}

//         // 4. Tandai kunjungan di cabang ini (buat salinan map)
//         newVisited := make(map[string]bool)
//         for k, v := range visitedOnBranch { newVisited[k] = v }
//         newVisited[elementToFind] = true

// 		// 5. Dapatkan resep pembuat
// 		recipes := recipeMap[elementToFind]
// 		if len(recipes) == 0 { return /* Buntu */ }

// 		// 6. Jelajahi setiap resep
// 		for _, recipe := range recipes {
//             // Tambahkan resep ini ke depan path saat ini
//             pathIncludingCurrent := append([]Recipe{recipe}, currentPathRecipes...)

//             // Cek apakah kedua bahan adalah elemen dasar
//             ing1IsBase := isBaseElementDFS(recipe.Ingredient1)
//             ing2IsBase := isBaseElementDFS(recipe.Ingredient2)

//             if ing1IsBase && ing2IsBase {
//                 // --- Jalur Lengkap Ditemukan ---
//                 pathID := generatePathIdentifierDFS(pathIncludingCurrent) // Buat ID unik
//                 mu.Lock()
//                 if !addedPathIdentifiers[pathID] && foundCount < maxRecipes {
//                     // Buat salinan path sebelum disimpan
//                     completePath := make([]Recipe, len(pathIncludingCurrent))
//                     copy(completePath, pathIncludingCurrent)
//                     allFoundPaths = append(allFoundPaths, completePath)
//                     addedPathIdentifiers[pathID] = true
//                     foundCount++
//                     fmt.Printf("DFS Multiple: Path UNIK ke-%d ditemukan (target: %s)\n", foundCount, targetElement)

//                     if foundCount >= maxRecipes {
//                         quitOnce.Do(func() { close(quitChan) }) // Kirim sinyal berhenti
//                     }
//                 }
//                 mu.Unlock()
//                 // Lanjutkan cek resep lain (jangan return)

//             } else {
//                 // --- Lanjutkan Rekursi ---
//                 // Rekursi untuk bahan 1 (jika bukan base)
//                 if !ing1IsBase {
//                     findPathRecursive(recipe.Ingredient1, pathIncludingCurrent, newVisited)
//                 }

//                 // Cek sinyal berhenti sebelum rekursi kedua
//                 select {
//                 case <-quitChan: return
//                 default: // Lanjutkan
//                     // Rekursi untuk bahan 2 (jika bukan base)
//                     if !ing2IsBase {
//                         findPathRecursive(recipe.Ingredient2, pathIncludingCurrent, newVisited)
//                     }
//                 }
//             }

//              // Cek sinyal berhenti setelah selesai satu resep
//              select {
//              case <-quitChan: return
//              default: // Lanjutkan
//              }
// 		} // Akhir loop resep
// 	} // Akhir findPathRecursive

// 	// --- Mulai Pencarian Paralel ---
// 	initialRecipes := recipeMap[targetElement]
// 	if len(initialRecipes) == 0 {
// 		return nil, nodesVisitedCount, fmt.Errorf("tidak ada resep yang diketahui untuk membuat %s", targetElement)
// 	}

// 	fmt.Printf("Memulai %d pencarian DFS paralel awal untuk %s...\n", len(initialRecipes), targetElement)
// 	for _, recipe := range initialRecipes {
// 		wg.Add(1) // Tambah counter WaitGroup
// 		go func(startRecipe Recipe) {
// 			defer wg.Done() // Pastikan Done dipanggil

//             initialPath := []Recipe{startRecipe} // Path awal hanya resep ini
//             initialVisited := make(map[string]bool) // Visited map awal
//             initialVisited[targetElement] = true // Tandai target agar tidak kembali

//             ing1IsBase := isBaseElementDFS(startRecipe.Ingredient1)
//             ing2IsBase := isBaseElementDFS(startRecipe.Ingredient2)

//             // Jika resep awal sudah lengkap
//             if ing1IsBase && ing2IsBase {
//                  pathID := generatePathIdentifierDFS(initialPath)
//                  mu.Lock()
//                  if !addedPathIdentifiers[pathID] && foundCount < maxRecipes {
//                      finalPath := make([]Recipe, len(initialPath))
//                      copy(finalPath, initialPath)
//                      allFoundPaths = append(allFoundPaths, finalPath)
//                      addedPathIdentifiers[pathID] = true
//                      foundCount++
//                      fmt.Printf("DFS Multiple: Path UNIK ke-%d ditemukan (target: %s) - Initial Recipe\n", foundCount, targetElement)
//                      if foundCount >= maxRecipes {
//                           quitOnce.Do(func() { close(quitChan) })
//                      }
//                  }
//                  mu.Unlock()
//             } else {
//                 // Mulai rekursi untuk bahan non-dasar
//                 if !ing1IsBase {
//                     findPathRecursive(startRecipe.Ingredient1, initialPath, initialVisited)
//                 }
//                 select { // Cek quit signal
//                 case <-quitChan: return
//                 default:
//                      if !ing2IsBase {
//                          findPathRecursive(startRecipe.Ingredient2, initialPath, initialVisited)
//                      }
//                 }
//             }
// 		}(recipe) // Jalankan goroutine dengan salinan resep
// 	}

// 	// Tunggu semua goroutine awal selesai
// 	wg.Wait()
// 	fmt.Println("Semua goroutine DFS awal selesai.")
//     quitOnce.Do(func() { close(quitChan) }) // Pastikan channel ditutup

// 	// Cek hasil akhir
// 	if len(allFoundPaths) == 0 {
//         if !isBaseElementDFS(targetElement) {
// 		    return nil, nodesVisitedCount, fmt.Errorf("jalur ke elemen '%s' tidak ditemukan", targetElement)
//         }
// 	}

// 	// Pastikan tidak mengembalikan lebih dari maxRecipes
//     finalPathsToReturn := allFoundPaths
//     mu.Lock()
//     if len(allFoundPaths) > maxRecipes {
//         finalPathsToReturn = allFoundPaths[:maxRecipes]
//     }
//     mu.Unlock()

// 	return finalPathsToReturn, nodesVisitedCount, nil
// }

// func FindMultiplePathsDFS(targetElement string, maxRecipes int) ([][]Recipe, int, error) {
//     fmt.Printf("Mencari %d jalur DFS BERBEDA ke: %s\n", maxRecipes, targetElement)

//     // Akses data yang diperlukan
//     recipeMap := GetRecipeMap() // Map[Hasil] -> []ResepPembuat
//     if recipeMap == nil {
//         return nil, 0, errors.New("map resep belum diinisialisasi")
//     }
//     if maxRecipes <= 0 {
//         return nil, 0, errors.New("jumlah resep minimal harus 1")
//     }
//     if isBaseElementDFS(targetElement) {
//         return [][]Recipe{}, 0, nil // Target adalah elemen dasar
//     }

//     nodesVisitedCount := -1 // Sulit dihitung akurat, set -1
    
//     // Cache untuk elemen yang bisa dibuat
//     knownCreatableElements := make(map[string]bool)
//     for _, base := range []string{"Air", "Earth", "Fire", "Water"} {
//         knownCreatableElements[base] = true
//     }
//     var knownCreatableMutex sync.RWMutex

//     // Cache untuk jalur terpendek ke setiap elemen
//     shortestPathCache := make(map[string][]Recipe)
//     var shortestPathMutex sync.RWMutex
    
//     // Fungsi untuk memeriksa apakah elemen bisa dibuat
//     var isCreatable func(element string, visited map[string]bool) bool
//     isCreatable = func(element string, visited map[string]bool) bool {
//         // Base case 1: Jika elemen dasar
//         if isBaseElementDFS(element) {
//             return true
//         }
        
//         // Base case 2: Sudah kita ketahui bisa dibuat
//         knownCreatableMutex.RLock()
//         if known, exists := knownCreatableElements[element]; exists {
//             knownCreatableMutex.RUnlock()
//             return known
//         }
//         knownCreatableMutex.RUnlock()
        
//         // Base case 3: Loop deteksi
//         if visited[element] {
//             return false
//         }
        
//         // Tandai sudah dikunjungi untuk cabang ini
//         newVisited := make(map[string]bool)
//         for k, v := range visited {
//             newVisited[k] = v
//         }
//         newVisited[element] = true
        
//         // Cek setiap resep yang menghasilkan elemen ini
//         recipes := recipeMap[element]
//         if len(recipes) == 0 {
//             // Simpan hasil: tidak bisa dibuat
//             knownCreatableMutex.Lock()
//             knownCreatableElements[element] = false
//             knownCreatableMutex.Unlock()
//             return false
//         }
        
//         // Cek apakah ada resep yang valid (kedua bahannya bisa dibuat)
//         for _, recipe := range recipes {
//             ing1Creatable := isCreatable(recipe.Ingredient1, newVisited)
//             ing2Creatable := isCreatable(recipe.Ingredient2, newVisited)
            
//             if ing1Creatable && ing2Creatable {
//                 // Simpan hasil: bisa dibuat
//                 knownCreatableMutex.Lock()
//                 knownCreatableElements[element] = true
//                 knownCreatableMutex.Unlock()
//                 return true
//             }
//         }
        
//         // Simpan hasil: tidak bisa dibuat
//         knownCreatableMutex.Lock()
//         knownCreatableElements[element] = false
//         knownCreatableMutex.Unlock()
//         return false
//     }

//     // Fungsi untuk menemukan jalur terpendek ke suatu elemen (rekursif dengan memoization)
//     var findShortestPath func(target string, visited map[string]bool) []Recipe
//     findShortestPath = func(target string, visited map[string]bool) []Recipe {
//         // Base case 1: Elemen dasar
//         if isBaseElementDFS(target) {
//             return []Recipe{}
//         }
        
//         // Base case 2: Sudah ada di cache
//         shortestPathMutex.RLock()
//         if path, exists := shortestPathCache[target]; exists {
//             shortestPathMutex.RUnlock()
//             return path
//         }
//         shortestPathMutex.RUnlock()
        
//         // Base case 3: Loop deteksi
//         if visited[target] {
//             return nil
//         }
        
//         // Tandai dikunjungi
//         newVisited := make(map[string]bool)
//         for k, v := range visited {
//             newVisited[k] = v
//         }
//         newVisited[target] = true
        
//         var shortestPathSoFar []Recipe
        
//         // Cari resep untuk target
//         recipes := recipeMap[target]
//         for _, recipe := range recipes {
//             // Periksa apakah bahan-bahan bisa dibuat
//             emptyVisited := make(map[string]bool)
//             if !isCreatable(recipe.Ingredient1, emptyVisited) || !isCreatable(recipe.Ingredient2, emptyVisited) {
//                 continue // Lewati resep yang tidak valid
//             }
            
//             // Cari path untuk kedua bahan
//             path1 := []Recipe{}
//             if !isBaseElementDFS(recipe.Ingredient1) {
//                 path1 = findShortestPath(recipe.Ingredient1, newVisited)
//                 if path1 == nil {
//                     continue // Tidak bisa membuat bahan 1
//                 }
//             }
            
//             path2 := []Recipe{}
//             if !isBaseElementDFS(recipe.Ingredient2) {
//                 path2 = findShortestPath(recipe.Ingredient2, newVisited)
//                 if path2 == nil {
//                     continue // Tidak bisa membuat bahan 2
//                 }
//             }
            
//             // Gabungkan path
//             currentPath := []Recipe{recipe}
//             currentPath = append(currentPath, path1...)
//             currentPath = append(currentPath, path2...)
            
//             // Perbarui jalur terpendek jika perlu
//             if shortestPathSoFar == nil || len(currentPath) < len(shortestPathSoFar) {
//                 shortestPathSoFar = currentPath
//             }
//         }
        
//         // Simpan hasil ke cache
//         if shortestPathSoFar != nil {
//             pathCopy := make([]Recipe, len(shortestPathSoFar))
//             copy(pathCopy, shortestPathSoFar)
            
//             shortestPathMutex.Lock()
//             shortestPathCache[target] = pathCopy
//             shortestPathMutex.Unlock()
//         }
        
//         return shortestPathSoFar
//     }

//     // Fungsi untuk menghasilkan jalur unik
//     var generateUniquePaths func(target string, maxPaths int) [][]Recipe
//     generateUniquePaths = func(target string, maxPaths int) [][]Recipe {
//         var results [][]Recipe
//         uniquePathMap := make(map[string]bool)
        
//         // Langkah 1: Cari jalur terpendek
//         shortestPath := findShortestPath(target, make(map[string]bool))
//         if shortestPath != nil {
//             pathID := generatePathIdentifierDFS(shortestPath)
//             uniquePathMap[pathID] = true
//             results = append(results, shortestPath)
//         }
        
//         // Jika cukup dengan 1 jalur, return
//         if maxPaths <= 1 || shortestPath == nil {
//             return results
//         }
        
//         // Langkah 2: Cari jalur alternatif dengan menghindari beberapa resep dari jalur terpendek
//         for i := 0; i < len(shortestPath) && len(results) < maxPaths; i++ {
//             // Coba cari jalur tanpa menggunakan resep ini
//             currentRecipe := shortestPath[i]
//             result := currentRecipe.Result
            
//             // Cari resep alternatif untuk hasil yang sama
//             alternativeRecipes := recipeMap[result]
//             for _, altRecipe := range alternativeRecipes {
//                 if altRecipe.Ingredient1 == currentRecipe.Ingredient1 && 
//                    altRecipe.Ingredient2 == currentRecipe.Ingredient2 {
//                     continue // Sama dengan resep saat ini, lewati
//                 }
                
//                 // Buat jalur baru dengan resep alternatif ini
//                 emptyVisited := make(map[string]bool)
//                 if !isCreatable(altRecipe.Ingredient1, emptyVisited) || 
//                    !isCreatable(altRecipe.Ingredient2, emptyVisited) {
//                     continue // Resep tidak valid
//                 }
                
//                 // Buat jalur lengkap dengan resep alternatif
//                 newPath := make([]Recipe, 0)
                
//                 // Tambahkan resep alternatif
//                 newPath = append(newPath, altRecipe)
                
//                 // Tambahkan jalur untuk kedua bahan dari resep alternatif
//                 if !isBaseElementDFS(altRecipe.Ingredient1) {
//                     ing1Path := findShortestPath(altRecipe.Ingredient1, make(map[string]bool))
//                     if ing1Path == nil {
//                         continue
//                     }
//                     newPath = append(newPath, ing1Path...)
//                 }
                
//                 if !isBaseElementDFS(altRecipe.Ingredient2) {
//                     ing2Path := findShortestPath(altRecipe.Ingredient2, make(map[string]bool))
//                     if ing2Path == nil {
//                         continue
//                     }
//                     newPath = append(newPath, ing2Path...)
//                 }
                
//                 // Periksa apakah jalur ini unik
//                 pathID := generatePathIdentifierDFS(newPath)
//                 if !uniquePathMap[pathID] {
//                     uniquePathMap[pathID] = true
//                     results = append(results, newPath)
                    
//                     if len(results) >= maxPaths {
//                         break
//                     }
//                 }
//             }
//         }
        
//         return results
//     }

//     // Temukan jalur-jalur unik untuk target
//     fmt.Printf("Mencari jalur terpendek dan alternatif untuk %s...\n", targetElement)
//     foundPaths := generateUniquePaths(targetElement, maxRecipes)
    
//     if len(foundPaths) == 0 {
//         return nil, nodesVisitedCount, fmt.Errorf("tidak ada jalur valid untuk membuat %s", targetElement)
//     }
    
//     return foundPaths, nodesVisitedCount, nil
// }

func FindMultiplePathsDFS(targetElement string, maxRecipes int) ([][]Recipe, int, error) {
    fmt.Printf("Mencari %d jalur DFS BERBEDA ke: %s\n", maxRecipes, targetElement)

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

    nodesVisitedCount := -1 // Sulit dihitung akurat, set -1
    
    // Cache untuk elemen yang bisa dibuat
    knownCreatableElements := make(map[string]bool)
    for _, base := range []string{"Air", "Earth", "Fire", "Water"} {
        knownCreatableElements[base] = true
    }
    var knownCreatableMutex sync.RWMutex

    // Cache untuk jalur terpendek ke setiap elemen
    shortestPathCache := make(map[string][]Recipe)
    var shortestPathMutex sync.RWMutex
    
    // Fungsi untuk memeriksa apakah elemen bisa dibuat
    var isCreatable func(element string, visited map[string]bool) bool
    isCreatable = func(element string, visited map[string]bool) bool {
        // Base case 1: Jika elemen dasar
        if isBaseElementDFS(element) {
            return true
        }
        
        // Base case 2: Sudah kita ketahui bisa dibuat
        knownCreatableMutex.RLock()
        if known, exists := knownCreatableElements[element]; exists {
            knownCreatableMutex.RUnlock()
            return known
        }
        knownCreatableMutex.RUnlock()
        
        // Base case 3: Loop deteksi
        if visited[element] {
            return false
        }
        
        // Tandai sudah dikunjungi untuk cabang ini
        newVisited := make(map[string]bool)
        for k, v := range visited {
            newVisited[k] = v
        }
        newVisited[element] = true
        
        // Cek setiap resep yang menghasilkan elemen ini
        recipes := recipeMap[element]
        if len(recipes) == 0 {
            // Simpan hasil: tidak bisa dibuat
            knownCreatableMutex.Lock()
            knownCreatableElements[element] = false
            knownCreatableMutex.Unlock()
            return false
        }
        
        // Cek apakah ada resep yang valid (kedua bahannya bisa dibuat)
        for _, recipe := range recipes {
            ing1Creatable := isCreatable(recipe.Ingredient1, newVisited)
            ing2Creatable := isCreatable(recipe.Ingredient2, newVisited)
            
            if ing1Creatable && ing2Creatable {
                // Simpan hasil: bisa dibuat
                knownCreatableMutex.Lock()
                knownCreatableElements[element] = true
                knownCreatableMutex.Unlock()
                return true
            }
        }
        
        // Simpan hasil: tidak bisa dibuat
        knownCreatableMutex.Lock()
        knownCreatableElements[element] = false
        knownCreatableMutex.Unlock()
        return false
    }

    // Fungsi untuk menemukan jalur terpendek ke suatu elemen (rekursif dengan memoization)
    var findShortestPath func(target string, visited map[string]bool) []Recipe
    findShortestPath = func(target string, visited map[string]bool) []Recipe {
        // Base case 1: Elemen dasar
        if isBaseElementDFS(target) {
            return []Recipe{}
        }
        
        // Base case 2: Sudah ada di cache
        shortestPathMutex.RLock()
        if path, exists := shortestPathCache[target]; exists {
            shortestPathMutex.RUnlock()
            pathCopy := make([]Recipe, len(path))
            copy(pathCopy, path)
            return pathCopy
        }
        shortestPathMutex.RUnlock()
        
        // Base case 3: Loop deteksi
        if visited[target] {
            return nil
        }
        
        // Tandai dikunjungi
        newVisited := make(map[string]bool)
        for k, v := range visited {
            newVisited[k] = v
        }
        newVisited[target] = true
        
        var shortestPathSoFar []Recipe
        
        // Cari resep untuk target
        recipes := recipeMap[target]
        for _, recipe := range recipes {
            // Periksa apakah bahan-bahan bisa dibuat
            emptyVisited := make(map[string]bool)
            if !isCreatable(recipe.Ingredient1, emptyVisited) || !isCreatable(recipe.Ingredient2, emptyVisited) {
                continue // Lewati resep yang tidak valid
            }
            
            // Cari path untuk kedua bahan
            path1 := []Recipe{}
            if !isBaseElementDFS(recipe.Ingredient1) {
                path1 = findShortestPath(recipe.Ingredient1, newVisited)
                if path1 == nil {
                    continue // Tidak bisa membuat bahan 1
                }
            }
            
            path2 := []Recipe{}
            if !isBaseElementDFS(recipe.Ingredient2) {
                path2 = findShortestPath(recipe.Ingredient2, newVisited)
                if path2 == nil {
                    continue // Tidak bisa membuat bahan 2
                }
            }
            
            // Gabungkan path
            currentPath := []Recipe{recipe}
            currentPath = append(currentPath, path1...)
            currentPath = append(currentPath, path2...)
            
            // Perbarui jalur terpendek jika perlu
            if shortestPathSoFar == nil || len(currentPath) < len(shortestPathSoFar) {
                shortestPathSoFar = currentPath
            }
        }
        
        // Simpan hasil ke cache
        if shortestPathSoFar != nil {
            pathCopy := make([]Recipe, len(shortestPathSoFar))
            copy(pathCopy, shortestPathSoFar)
            
            shortestPathMutex.Lock()
            shortestPathCache[target] = pathCopy
            shortestPathMutex.Unlock()
        }
        
        return shortestPathSoFar
    }

    // Fungsi untuk menghasilkan jalur unik
    // Fungsi untuk menghasilkan jalur unik
var generateUniquePaths func(target string, maxPaths int) [][]Recipe
generateUniquePaths = func(target string, maxPaths int) [][]Recipe {
    var results [][]Recipe
    uniquePathMap := make(map[string]bool)
    
    // Langkah 1: Cari jalur terpendek
    shortestPath := findShortestPath(target, make(map[string]bool))
    if shortestPath != nil {
        pathID := generatePathIdentifierDFS(shortestPath)
        uniquePathMap[pathID] = true
        results = append(results, shortestPath)
    }
    
    // Jika cukup dengan 1 jalur, return
    if maxPaths <= 1 || shortestPath == nil {
        return results
    }
    
    // Langkah 2: Cari jalur alternatif untuk target yang sama
    // Dapatkan semua resep yang bisa membuat target
    targetRecipes := recipeMap[target]
    // Coba setiap resep
    for _, recipe := range targetRecipes {
        // Skip jika resep ini sama dengan yang digunakan di jalur terpendek
        if len(shortestPath) > 0 && shortestPath[0].Result == recipe.Result && 
           shortestPath[0].Ingredient1 == recipe.Ingredient1 && 
           shortestPath[0].Ingredient2 == recipe.Ingredient2 {
            continue
        }
        
        // Buat jalur baru dengan resep alternatif ini
        emptyVisited := make(map[string]bool)
        if !isCreatable(recipe.Ingredient1, emptyVisited) || 
           !isCreatable(recipe.Ingredient2, emptyVisited) {
            continue // Resep tidak valid
        }
        
        // Buat jalur lengkap dengan resep alternatif
        newPath := []Recipe{recipe}
        
        // Tambahkan jalur untuk kedua bahan
        if !isBaseElementDFS(recipe.Ingredient1) {
            ing1Path := findShortestPath(recipe.Ingredient1, make(map[string]bool))
            if ing1Path == nil {
                continue
            }
            newPath = append(newPath, ing1Path...)
        }
        
        if !isBaseElementDFS(recipe.Ingredient2) {
            ing2Path := findShortestPath(recipe.Ingredient2, make(map[string]bool))
            if ing2Path == nil {
                continue
            }
            newPath = append(newPath, ing2Path...)
        }
        
        // Periksa apakah jalur ini unik
        pathID := generatePathIdentifierDFS(newPath)
        if !uniquePathMap[pathID] {
            uniquePathMap[pathID] = true
            results = append(results, newPath)
            
            if len(results) >= maxPaths {
                break
            }
        }
    }
    
// Langkah 3: Jika masih perlu jalur tambahan, coba variasi dari jalur yang sudah ada
if len(results) < maxPaths {
    // Buat jalur alternatif dengan mengganti resep-resep lain di jalur pertama
    for _, basePath := range results {
        if len(results) >= maxPaths {
            break
        }
        
        // Coba variasi untuk setiap resep dalam jalur
        for i := 0; i < len(basePath) && len(results) < maxPaths; i++ {
            currentRecipe := basePath[i]
            resultElement := currentRecipe.Result
            
            alternativeRecipes := recipeMap[resultElement]
            for _, altRecipe := range alternativeRecipes {
                if altRecipe.Ingredient1 == currentRecipe.Ingredient1 && 
                   altRecipe.Ingredient2 == currentRecipe.Ingredient2 {
                    continue // Sama dengan resep saat ini
                }
                
                // Verifikasi bahan alternatif dapat dibuat
                emptyVisited := make(map[string]bool)
                if !isCreatable(altRecipe.Ingredient1, emptyVisited) || 
                   !isCreatable(altRecipe.Ingredient2, emptyVisited) {
                    continue // Resep tidak valid
                }
                
                // PERBAIKAN: Bangun jalur lengkap baru dengan resep alternatif
                // mulai dari target
                newPath := []Recipe{}
                if i == 0 { // Jika mengganti resep pertama (resep target)
                    // Tambahkan resep alternatif
                    newPath = append(newPath, altRecipe)
                    
                    // Tambahkan jalur untuk bahan-bahan alternatif
                    if !isBaseElementDFS(altRecipe.Ingredient1) {
                        ing1Path := findShortestPath(altRecipe.Ingredient1, make(map[string]bool))
                        if ing1Path == nil {
                            continue
                        }
                        newPath = append(newPath, ing1Path...)
                    }
                    
                    if !isBaseElementDFS(altRecipe.Ingredient2) {
                        ing2Path := findShortestPath(altRecipe.Ingredient2, make(map[string]bool))
                        if ing2Path == nil {
                            continue
                        }
                        newPath = append(newPath, ing2Path...)
                    }
                } else {
                    // Untuk resep yang bukan pertama, kita perlu:
                    // 1. Temukan resep mana dalam path lama yang membuat elemen untuk resep ini
                    // 2. Bangun ulang jalur dari awal
                    
                    // Pertama, salin resep original dari basePath yang masih valid
                    newPath = append(newPath, basePath[0])
                    
                    // Cari jalur untuk membuat bahan-bahan resep pertama yang masih valid
                    for j := 1; j < len(basePath); j++ {
                        if j == i {
                            // Ganti dengan resep alternatif
                            newPath = append(newPath, altRecipe)
                            
                            // Tambahkan jalur untuk bahan-bahan resep alternatif
                            if !isBaseElementDFS(altRecipe.Ingredient1) {
                                ing1Path := findShortestPath(altRecipe.Ingredient1, make(map[string]bool))
                                if ing1Path != nil {
                                    newPath = append(newPath, ing1Path...)
                                }
                            }
                            
                            if !isBaseElementDFS(altRecipe.Ingredient2) {
                                ing2Path := findShortestPath(altRecipe.Ingredient2, make(map[string]bool))
                                if ing2Path != nil {
                                    newPath = append(newPath, ing2Path...)
                                }
                            }
                        } else {
                            // Tambahkan resep asli
                            newPath = append(newPath, basePath[j])
                        }
                    }
                }
                
                // Periksa apakah jalur ini unik dan valid
                if len(newPath) > 0 {
                    pathID := generatePathIdentifierDFS(newPath)
                    if !uniquePathMap[pathID] {
                        uniquePathMap[pathID] = true
                        results = append(results, newPath)
                        
                        if len(results) >= maxPaths {
                            break
                        }
                    }
                }
            }
        }
    }
}
    
    return results
}

    // Temukan jalur-jalur unik untuk target
    fmt.Printf("Mencari jalur terpendek dan alternatif untuk %s...\n", targetElement)
    foundPaths := generateUniquePaths(targetElement, maxRecipes)
    
    if len(foundPaths) == 0 {
        return nil, nodesVisitedCount, fmt.Errorf("tidak ada jalur valid untuk membuat %s", targetElement)
    }
    
    // PERUBAHAN: Balik urutan setiap jalur sebelum menampilkan & mengembalikan
    reversedPaths := make([][]Recipe, len(foundPaths))
    for i, path := range foundPaths {
        reversedPaths[i] = make([]Recipe, len(path))
        for j, recipe := range path {
            reversedPaths[i][len(path)-1-j] = recipe
        }
        
        // Debug - tampilkan jalur yang ditemukan dengan urutan terbalik
        fmt.Printf("Jalur %d (panjang: %d):\n", i+1, len(reversedPaths[i]))
        for j, recipe := range reversedPaths[i] {
            fmt.Printf("  Langkah %d: %s + %s => %s\n", 
                      j+1, recipe.Ingredient1, recipe.Ingredient2, recipe.Result)
        }
    }
    
    return reversedPaths, nodesVisitedCount, nil
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
// func generatePathIdentifierDFS(path []Recipe) string {
// 	recipesCopy := make([]Recipe, len(path))
// 	copy(recipesCopy, path)
// 	sort.Slice(recipesCopy, func(i, j int) bool {
// 		if recipesCopy[i].Result != recipesCopy[j].Result {
// 			return recipesCopy[i].Result < recipesCopy[j].Result
// 		}
//         ing1i, ing2i := recipesCopy[i].Ingredient1, recipesCopy[i].Ingredient2
//         if ing1i > ing2i { ing1i, ing2i = ing2i, ing1i }
//         ing1j, ing2j := recipesCopy[j].Ingredient1, recipesCopy[j].Ingredient2
//         if ing1j > ing2j { ing1j, ing2j = ing2j, ing1j }
// 		if ing1i != ing1j {
// 			return ing1i < ing1j
// 		}
// 		return ing2i < ing2j
// 	})
// 	var parts []string
// 	for _, r := range recipesCopy {
//         ing1, ing2 := r.Ingredient1, r.Ingredient2
//         if ing1 > ing2 { ing1, ing2 = ing2, ing1 }
// 		parts = append(parts, fmt.Sprintf("%s+%s=>%s", ing1, ing2, r.Result))
// 	}
// 	return strings.Join(parts, "|")
// }
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