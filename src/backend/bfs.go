// src/backend/bfs.go
package main

import (
	"container/list"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// FindPathBFS (fungsi ini tetap sama, untuk mode shortest)
func FindPathBFS(targetElement string) ([]Recipe, int, error) {
	fmt.Printf("Mencari jalur BFS (shortest) ke: %s\n", targetElement)
	graph := GetAlchemyGraph()
	if graph == nil {
		return nil, 0, errors.New("graf alkimia belum diinisialisasi")
	}

	queue := list.New()
	visited := make(map[string]bool)
	parent := make(map[string]Recipe) // Tipe map[string]Recipe untuk BFS standar
	nodesVisitedCount := 0

	baseElements := []string{"Air", "Earth", "Fire", "Water"}

	for _, base := range baseElements {
		if base == targetElement {
			return []Recipe{}, 0, nil
		}
		if _, exists := graph[base]; exists || isBaseElement(base) {
			if !visited[base] {
				queue.PushBack(base)
				visited[base] = true
			}
		}
	}

	for queue.Len() > 0 {
		queueElement := queue.Front()
		currentElement := queueElement.Value.(string)
		queue.Remove(queueElement)
		nodesVisitedCount++

		recipesInvolvingCurrent := graph[currentElement]

		for _, recipe := range recipesInvolvingCurrent {
			var otherIngredient string
			if recipe.Ingredient1 == currentElement {
				otherIngredient = recipe.Ingredient2
			} else {
				otherIngredient = recipe.Ingredient1
			}

			if visited[otherIngredient] {
				result := recipe.Result
				if !visited[result] {
					visited[result] = true
					parent[result] = recipe // Simpan resep pertama
					queue.PushBack(result)

					if result == targetElement {
						fmt.Printf("BFS (shortest): Target '%s' ditemukan!\n", targetElement)
						path := reconstructPathRevised(parent, targetElement)
						return path, nodesVisitedCount, nil
					}
				}
			}
		}
	}

	fmt.Printf("BFS (shortest): Target '%s' tidak dapat ditemukan.\n", targetElement)
	return nil, nodesVisitedCount, fmt.Errorf("jalur ke elemen '%s' tidak ditemukan", targetElement)
}


// --- Implementasi BFS Multiple Path (Logika Berhenti Dilonggarkan) ---

// generatePathIdentifier (fungsi ini tetap sama)
func generatePathIdentifier(path []Recipe) string {
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


// FindMultiplePathsBFS mencari sejumlah 'maxRecipes' jalur resep BERBEDA ke targetElement menggunakan BFS.
// Revisi 3: Melonggarkan kondisi berhenti agar bisa menemukan path sedikit lebih panjang.
func FindMultiplePathsBFS(targetElement string, maxRecipes int) ([][]Recipe, int, error) {
	fmt.Printf("Mencari %d jalur BFS BERBEDA ke: %s (Revisi 3 - Looser Stop)\n", maxRecipes, targetElement)

	graph := GetAlchemyGraph()
	if graph == nil { return nil, 0, errors.New("graf alkimia belum diinisialisasi") }
    if maxRecipes <= 0 { return nil, 0, errors.New("jumlah resep minimal harus 1") }
    if isBaseElement(targetElement) { return [][]Recipe{}, 0, nil }

    // Struktur untuk BFS
	queue := list.New()
	visited := make(map[string]bool) // Cukup bool untuk melacak kunjungan dasar
	parent := make(map[string]Recipe) // Map parent standar untuk rekonstruksi dasar
	nodesVisitedCount := 0

	allFoundPaths := [][]Recipe{} // Hasil akhir
	foundCount := 0               // Jumlah path unik yang ditemukan
    addedPathIdentifiers := make(map[string]bool) // Untuk cek duplikasi path
    var mu sync.Mutex             // Mutex

	// Inisialisasi queue & visited map untuk elemen dasar
	baseElements := []string{"Air", "Earth", "Fire", "Water"}
	for _, base := range baseElements {
		queue.PushBack(base)
		visited[base] = true
	}

	// Loop BFS Utama - Berhenti HANYA jika queue kosong ATAU sudah cukup path
	for queue.Len() > 0 && foundCount < maxRecipes {

        // Dequeue
		queueElement := queue.Front()
		currentElement := queueElement.Value.(string)
		queue.Remove(queueElement)
		nodesVisitedCount++

		// Iterasi resep yang melibatkan currentElement
		recipesInvolvingCurrent := graph[currentElement]
		for _, recipe := range recipesInvolvingCurrent {
             // Cek apakah sudah cukup path ditemukan SEBELUM memproses resep ini
             mu.Lock()
             stopEarly := foundCount >= maxRecipes
             mu.Unlock()
             if stopEarly { break } // Keluar dari loop resep jika sudah cukup

			var otherIngredient string
			if recipe.Ingredient1 == currentElement {
				otherIngredient = recipe.Ingredient2
			} else {
				otherIngredient = recipe.Ingredient1
			}

			// Jika bahan lain sudah dikunjungi
			if visited[otherIngredient] {
				result := recipe.Result

                // Jika hasil BELUM PERNAH dikunjungi, tandai, simpan parent PERTAMA, dan enqueue
				if !visited[result] {
					visited[result] = true
					parent[result] = recipe // Simpan resep pertama
					queue.PushBack(result)
				}
                // ELSE: Jika hasil sudah dikunjungi, kita tidak enqueue lagi,
                // tapi kita TETAP cek apakah itu target, karena bisa jadi ini
                // adalah jalur lain (mungkin lebih panjang) ke target.

				// Cek jika hasil adalah target
				if result == targetElement {
                    // Buat salinan parent map & pastikan resep terakhir benar
                    tempParent := make(map[string]Recipe)
                    for k, v := range parent {
                        tempParent[k] = v
                    }
                    tempParent[targetElement] = recipe // Gunakan resep saat ini

                    currentPath := reconstructPathRevised(tempParent, targetElement)

                    if len(currentPath) > 0 {
                        pathID := generatePathIdentifier(currentPath)
                        mu.Lock()
                        if !addedPathIdentifiers[pathID] && foundCount < maxRecipes {
                            pathToAdd := make([]Recipe, len(currentPath))
                            copy(pathToAdd, currentPath)
                            allFoundPaths = append(allFoundPaths, pathToAdd)
                            addedPathIdentifiers[pathID] = true
                            foundCount++
                            fmt.Printf("BFS Multiple: Path UNIK ke-%d ditemukan (target: %s, Resep Akhir: %v)\n", foundCount, targetElement, recipe)
                            // Jika sudah cukup, break dari loop resep (akan dicek lagi di loop queue)
                            if foundCount >= maxRecipes {
                                mu.Unlock()
                                break
                            }
                        }
                        mu.Unlock()
                    }
				} // end if result == targetElement
			} // end if visited[otherIngredient]
		} // End loop resep
	} // End loop queue

	// Cek hasil akhir
	if len(allFoundPaths) == 0 {
		fmt.Printf("BFS Multiple: Target '%s' tidak dapat ditemukan.\n", targetElement)
        if !isBaseElement(targetElement) {
		    return nil, nodesVisitedCount, fmt.Errorf("jalur ke elemen '%s' tidak ditemukan", targetElement)
        }
	}

	return allFoundPaths, nodesVisitedCount, nil
}


// reconstructPathRevised (fungsi ini tetap sama)
func reconstructPathRevised(parent map[string]Recipe, target string) []Recipe {
    path := list.New()
    queue := list.New()
    queue.PushBack(target)
    processedForPath := make(map[string]bool)
	processedForPath[target]=true

    for queue.Len() > 0 {
        queueEl := queue.Front()
        current := queueEl.Value.(string)
        queue.Remove(queueEl)

        recipe, exists := parent[current]
        if !exists { continue }

        path.PushFront(recipe)

        if _, processed := processedForPath[recipe.Ingredient1]; !processed {
             if _, parentExists := parent[recipe.Ingredient1]; parentExists || isBaseElement(recipe.Ingredient1) {
                queue.PushBack(recipe.Ingredient1)
                processedForPath[recipe.Ingredient1] = true
             }
        }
		if _, processed := processedForPath[recipe.Ingredient2]; !processed {
             if _, parentExists := parent[recipe.Ingredient2]; parentExists || isBaseElement(recipe.Ingredient2) {
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

// isBaseElement (fungsi ini tetap sama)
func isBaseElement(name string) bool {
	baseElements := []string{"Air", "Earth", "Fire", "Water"}
	for _, base := range baseElements {
		if name == base {
			return true
		}
	}
	return false
}
