// src/backend/bds.go
package main

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
)

// isBaseElement dan generatePathIdentifier diasumsikan ada di bfs.go atau file util bersama

func printParentMap(name string, parentMap map[string]Recipe) {
	fmt.Printf("--- Isi Parent Map: %s ---\n", name)
	if len(parentMap) == 0 {
		fmt.Println("(Kosong)")
		return
	}
	for key, recipe := range parentMap {
		fmt.Printf("  ParentMap['%s']: %s + %s => %s\n", key, recipe.Ingredient1, recipe.Ingredient2, recipe.Result)
	}
	fmt.Println("--- Akhir Parent Map ---")
}

// reconstructSinglePathHelper: parentMap[HASIL_LANGKAH_INI] = RESEP_YANG_MENGHASILKANNYA
// startNode adalah HASIL dari langkah terakhir di segmen ini.
// stopNode adalah HASIL dari langkah pertama di segmen ini (atau "" untuk dasar).
func reconstructSinglePathHelper(parentMap map[string]Recipe, startNode string, stopNode string) []Recipe {
	pathList := list.New()
	curr := startNode

	fmt.Printf("Rekonstruksi Single: Mulai dari '%s'", curr)
	if stopNode != "" { fmt.Printf(" menuju stopNode '%s'", stopNode) } else { fmt.Printf(" menuju elemen dasar")}
	fmt.Println()
	// printParentMap("Digunakan dalam reconstructSinglePathHelper untuk startNode: "+startNode, parentMap) // Optional: uncomment for extreme debugging

	processedInPath := make(map[string]bool)

	for {
		fmt.Printf("Rekonstruksi Single - Iterasi: curr = '%s'\n", curr)

		if curr == stopNode && stopNode != "" {
			fmt.Printf("Rekonstruksi Single: Mencapai stopNode yang ditentukan '%s'.\n", curr)
			break
		}
		if isBaseElement(curr) && stopNode == "" {
			fmt.Printf("Rekonstruksi Single: Mencapai elemen dasar '%s'.\n", curr)
			break
		}
		if processedInPath[curr] { // Mencegah loop dalam rekonstruksi
		    fmt.Printf("Rekonstruksi Single: Loop terdeteksi pada '%s'. Berhenti.\n", curr)
		    return []Recipe{} // Kembalikan path kosong jika loop terdeteksi
		}
		processedInPath[curr] = true

		recipe, exists := parentMap[curr] // recipe adalah {I1, I2, curr}
		if !exists {
			if (stopNode != "" && curr != stopNode) || (stopNode == "" && !isBaseElement(curr)) {
			    fmt.Printf("Rekonstruksi Single: Tidak ada parent untuk '%s' di parentMap sebelum mencapai tujuan. Berhenti.\n", curr)
            } else {
                 fmt.Printf("Rekonstruksi Single: Berhenti normal di '%s' (mungkin stopNode atau elemen dasar tanpa parent eksplisit di map).\n", curr)
            }
			break
		}

		pathList.PushFront(recipe)
		fmt.Printf("Rekonstruksi Single: Menambahkan resep ke path (%s + %s => %s)\n", recipe.Ingredient1, recipe.Ingredient2, recipe.Result)

		nextCand1 := recipe.Ingredient1
		nextCand2 := recipe.Ingredient2
		chosenParent := ""

		fmt.Printf("Rekonstruksi Single: Mencari parent dari '%s' melalui resep %v. Kandidat: '%s', '%s'. StopNode: '%s'\n", curr, recipe, nextCand1, nextCand2, stopNode)

		// Logika pemilihan parent untuk mundur:
		// Pilih bahan yang merupakan stopNode, atau punya parent lagi, atau dasar (jika tidak ada stopNode)
		if stopNode != "" { // Mode mundur menuju meetingNode, atau maju menuju meetingNode (stopNode adalah meetingNode)
			if nextCand1 == stopNode { chosenParent = nextCand1 } else 
			if nextCand2 == stopNode { chosenParent = nextCand2 }
		}

		if chosenParent == "" { // Jika stopNode belum dipilih atau tidak ada stopNode (mode maju ke dasar)
			_, p1Exists := parentMap[nextCand1]
			_, p2Exists := parentMap[nextCand2]

			// Prioritaskan yang ada di parentMap dan bukan loop
			if p1Exists && nextCand1 != curr { chosenParent = nextCand1 } else 
			if p2Exists && nextCand2 != curr { chosenParent = nextCand2 } else 
			// Kemudian elemen dasar (hanya jika tidak ada stopNode spesifik)
			if isBaseElement(nextCand1) && nextCand1 != curr && stopNode == "" { chosenParent = nextCand1 } else 
			if isBaseElement(nextCand2) && nextCand2 != curr && stopNode == "" { chosenParent = nextCand2 } else {
				// Jika tidak ada parent di map dan bukan dasar/stopNode, berhenti
				fmt.Printf("Rekonstruksi Single: Tidak bisa menentukan langkah mundur valid dari '%s' (bahan tidak punya parent/dasar/stopNode).\n", curr)
				break
			}
		}
        
        if chosenParent == "" {
             fmt.Printf("Rekonstruksi Single: Tidak ada chosenParent yang valid untuk '%s'. Berhenti.\n", curr)
             break
        }
        // Mencegah loop A+B=A jika A bukan stopNode (atau bukan elemen dasar jika stopNode="")
        if chosenParent == curr && ( (stopNode != "" && curr != stopNode) || (stopNode == "" && !isBaseElement(curr)) ) {
             fmt.Printf("Rekonstruksi Single: Terdeteksi akan loop di '%s' saat menuju '%s'. Berhenti.\n", curr, stopNode)
             break
        }
		fmt.Printf("Rekonstruksi Single: Mundur ke '%s'.\n", chosenParent)
		curr = chosenParent
	}

	finalPath := make([]Recipe, 0, pathList.Len())
	for e := pathList.Front(); e != nil; e = e.Next() {
		finalPath = append(finalPath, e.Value.(Recipe))
	}
	return finalPath
}

func reconstructBidirectionalPath(parentForward map[string]Recipe, parentBackward map[string]Recipe, meetingNode string, originalTarget string) []Recipe {
	fmt.Printf("Rekonstruksi BDS: Bertemu di '%s'. Target awal: '%s'\n", meetingNode, originalTarget)
    printParentMap("parentForward saat rekonstruksi", parentForward)
    printParentMap("parentBackward saat rekonstruksi", parentBackward)

	// 1. Jalur Maju: Dari meetingNode ke elemen dasar.
	// parentForward[X] = resep yang menghasilkan X.
	pathForward := reconstructSinglePathHelper(parentForward, meetingNode, "")
	fmt.Println("--- Jalur Maju (ke meetingNode) Selesai Direkonstruksi ---")
	for i, r := range pathForward { fmt.Printf("Maju %d: %s + %s => %s\n", i+1, r.Ingredient1, r.Ingredient2, r.Result) }

	// 2. Jalur Mundur: Dari originalTarget ke meetingNode.
	// parentBackward[X] = resep yang menghasilkan X (dari arah mundur).
	pathSegmentMeetingToTarget := []Recipe{}
	if meetingNode != originalTarget {
		// Rekonstruksi dari originalTarget ke meetingNode
		pathBackwardTemp := reconstructSinglePathHelper(parentBackward, originalTarget, meetingNode)
		fmt.Println("--- Jalur Mundur Temp (dari target ke meeting, sebelum dibalik) Selesai Direkonstruksi ---")
		for i, r := range pathBackwardTemp { fmt.Printf("Mundur Temp %d: %s + %s => %s\n", i+1, r.Ingredient1, r.Ingredient2, r.Result) }

		// Balik urutannya untuk mendapatkan dari meetingNode ke originalTarget
		for i := len(pathBackwardTemp) - 1; i >= 0; i-- {
			pathSegmentMeetingToTarget = append(pathSegmentMeetingToTarget, pathBackwardTemp[i])
		}
	}

	fmt.Println("--- Jalur Mundur (dari meetingNode ke target, setelah dibalik) Selesai ---")
	for i, r := range pathSegmentMeetingToTarget { fmt.Printf("Mundur %d: %s + %s => %s\n", i+1, r.Ingredient1, r.Ingredient2, r.Result) }

	fullPath := append([]Recipe{}, pathForward...)
	fullPath = append(fullPath, pathSegmentMeetingToTarget...)
	
	if len(fullPath) == 0 && !isBaseElement(originalTarget) { 
		if meetingNode != originalTarget {
			fmt.Printf("Peringatan Rekonstruksi: Jalur BDS kosong untuk target non-dasar '%s'. Meeting: '%s'\n", originalTarget, meetingNode)
		} else if !isBaseElement(originalTarget) {
			fmt.Printf("Peringatan Rekonstruksi: Jalur BDS kosong, target '%s' mungkin tidak dapat dibuat (meeting node == target).\n", originalTarget)
		}
	} else if len(fullPath) > 0 { 
		if !isBaseElement(originalTarget) { 
			lastRecipe := fullPath[len(fullPath)-1]
			if lastRecipe.Result != originalTarget { 
				fmt.Printf("Peringatan Validasi Rekonstruksi: Resep terakhir (%v) TIDAK menghasilkan target (%s)\n", lastRecipe, originalTarget) 
			} else {
				fmt.Printf("Validasi Rekonstruksi: Resep terakhir (%v) menghasilkan target (%s) - OK\n", lastRecipe, originalTarget)
			} 
		} 
	}
	fmt.Println("\n--- Jalur Akhir Hasil Rekonstruksi BDS ---")
	for i, recipe := range fullPath { fmt.Printf("Langkah %d: %s + %s => %s\n", i+1, recipe.Ingredient1, recipe.Ingredient2, recipe.Result) }

	return fullPath
}

// FindPathBDS dengan pengisian parentBackward yang KONSISTEN
func FindPathBDS(targetElement string) ([]Recipe, int, error) {
	fmt.Printf("Memulai Bidirectional Search (shortest path) ke: %s\n", targetElement)
	recipeMap := GetRecipeMap(); alchemyGraph := GetAlchemyGraph()
	if recipeMap == nil || alchemyGraph == nil { return nil, 0, errors.New("data belum diinisialisasi") }
	if isBaseElement(targetElement) { return []Recipe{}, 0, nil }

	nodesVisitedCount := 0
	queueForward := list.New(); visitedForward := make(map[string]bool); parentForward := make(map[string]Recipe)
	queueBackward := list.New(); visitedBackward := make(map[string]bool); parentBackward := make(map[string]Recipe) // parentBackward[X] = resep yg menghasilkan X dari arah target

	baseElements := []string{"Air", "Earth", "Fire", "Water"}
	for _, base := range baseElements {
		if !visitedForward[base] { queueForward.PushBack(base); visitedForward[base] = true }
	}
	if !visitedBackward[targetElement] { queueBackward.PushBack(targetElement); visitedBackward[targetElement] = true }

	for queueForward.Len() > 0 || queueBackward.Len() > 0 {
		// --- MAJU ---
		if queueForward.Len() > 0 {
			lenF := queueForward.Len()
			for i := 0; i < lenF; i++ {
				if queueForward.Len() == 0 { break }
				currF := queueForward.Remove(queueForward.Front()).(string)
				nodesVisitedCount++
				fmt.Printf("BDS Maju: Memproses '%s'. VisitedBackward['%s']: %t\n", currF, currF, visitedBackward[currF])
				if visitedBackward[currF] {
					fmt.Printf("Pertemuan dari Maju di '%s'\n", currF)
					return reconstructBidirectionalPath(parentForward, parentBackward, currF, targetElement), nodesVisitedCount, nil
				}
				recipesUsingCurrF := alchemyGraph[currF]
				for _, recipe := range recipesUsingCurrF {
					otherIng := ""; if recipe.Ingredient1 == currF { otherIng = recipe.Ingredient2 } else if recipe.Ingredient2 == currF { otherIng = recipe.Ingredient1 } else { continue }
					if visitedForward[otherIng] {
						result := recipe.Result
						if !visitedForward[result] {
							visitedForward[result] = true; parentForward[result] = recipe; queueForward.PushBack(result)
							fmt.Printf("BDS Maju: Parent['%s'] = %v (dari %s + %s)\n", result, recipe, recipe.Ingredient1, recipe.Ingredient2)
							if visitedBackward[result] {
								fmt.Printf("Pertemuan dari Maju (ekspansi) di '%s'\n", result)
								return reconstructBidirectionalPath(parentForward, parentBackward, result, targetElement), nodesVisitedCount, nil
							}
						}
					}
				}
			}
		}

		// --- MUNDUR ---
		if queueBackward.Len() > 0 {
			lenB := queueBackward.Len()
			for i := 0; i < lenB; i++ {
				if queueBackward.Len() == 0 { break }
				// currB_as_RESULT adalah elemen yang kita pop dari queueBackward. Ini adalah HASIL dari suatu resep dari arah mundur.
				currB_as_RESULT := queueBackward.Remove(queueBackward.Front()).(string)
				nodesVisitedCount++
				fmt.Printf("BDS Mundur: Memproses HASIL mundur '%s'. VisitedForward['%s']: %t\n", currB_as_RESULT, currB_as_RESULT, visitedForward[currB_as_RESULT])

				if visitedForward[currB_as_RESULT] {
					fmt.Printf("Pertemuan dari Mundur di '%s'\n", currB_as_RESULT)
					return reconstructBidirectionalPath(parentForward, parentBackward, currB_as_RESULT, targetElement), nodesVisitedCount, nil
				}

				// Cari resep-resep R = {I1, I2, currB_as_RESULT} yang MENGHASILKAN currB_as_RESULT.
				recipes_that_make_currB := recipeMap[currB_as_RESULT]
				for _, recipe_makes_currB := range recipes_that_make_currB { // recipe_makes_currB = {I1, I2, currB_as_RESULT}
					// Bahan-bahan (I1, I2) adalah langkah mundur selanjutnya ("anak-anak" dari currB_as_RESULT dari arah mundur).
					// Kita set parentBackward[currB_as_RESULT] = recipe_makes_currB.
					// Ini berarti resep yang menghasilkan currB_as_RESULT (dari I1 dan I2) disimpan.
					if _, exists := parentBackward[currB_as_RESULT]; !exists { // Simpan resep pertama yang ditemukan (shortest path dari target)
						parentBackward[currB_as_RESULT] = recipe_makes_currB
						fmt.Printf("BDS Mundur: Parent['%s'] = %v (dari %s + %s)\n", currB_as_RESULT, recipe_makes_currB, recipe_makes_currB.Ingredient1, recipe_makes_currB.Ingredient2)
					}

					ingredients_as_prev_results := []string{recipe_makes_currB.Ingredient1, recipe_makes_currB.Ingredient2}
					for _, prev_step_result_ing := range ingredients_as_prev_results {
						if !visitedBackward[prev_step_result_ing] {
							visitedBackward[prev_step_result_ing] = true
							// `parentBackward[prev_step_result_ing]` akan diisi ketika `prev_step_result_ing` menjadi `currB_as_RESULT`
							queueBackward.PushBack(prev_step_result_ing)

							if visitedForward[prev_step_result_ing] {
								fmt.Printf("Pertemuan Mundur (ekspansi di bahan) di '%s'\n", prev_step_result_ing)
								// Saat pertemuan di 'prev_step_result_ing', pastikan parent-nya (currB_as_RESULT) sudah punya entri di parentBackward
								if _, ok := parentBackward[currB_as_RESULT]; !ok {
									parentBackward[currB_as_RESULT] = recipe_makes_currB // Set parent jika belum ada
									fmt.Printf("BDS Mundur (Pertemuan): Parent['%s'] = %v\n", currB_as_RESULT, recipe_makes_currB)
								}
								return reconstructBidirectionalPath(parentForward, parentBackward, prev_step_result_ing, targetElement), nodesVisitedCount, nil
							}
						}
					}
				}
			}
		}
		if queueForward.Len() == 0 && queueBackward.Len() == 0 { break }
	}
	return nil, nodesVisitedCount, fmt.Errorf("jalur BDS (shortest) ke '%s' tidak ditemukan", targetElement)
}


// FindMultiplePathsBDS (tetap sama seperti sebelumnya)
func FindMultiplePathsBDS(targetElement string, maxRecipes int) ([][]Recipe, int, error) {
	fmt.Printf("Memulai Bidirectional Search (multiple paths, wajib BDS & multithreading) ke: %s, max: %d\n", targetElement, maxRecipes)

	if maxRecipes <= 0 {
		return nil, 0, errors.New("jumlah resep minimal harus 1")
	}
	if isBaseElement(targetElement) {
		return [][]Recipe{{}}, 0, nil
	}

	var allFoundPaths [][]Recipe
	addedPathIdentifiers := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup
	finalNodesVisited := -1

	foundCount := 0

	quitChan := make(chan struct{})
	var quitOnce sync.Once
	closeQuitChan := func() {
		quitOnce.Do(func() {
			close(quitChan)
		})
	}
	defer closeQuitChan()

	numGoroutines := maxRecipes
	if numGoroutines < 1 { numGoroutines = 1 }
	if numGoroutines > 10 { numGoroutines = 10 }

	if numGoroutines == 1 {
		fmt.Println("BDS Multiple: Mencari 1 jalur (serial)...")
		path, nodes, err := FindPathBDS(targetElement)
		finalNodesVisited = nodes
		if err == nil && path != nil {
			if len(path) > 0 || (len(path) == 0 && isBaseElement(targetElement)) {
				allFoundPaths = append(allFoundPaths, path)
				foundCount = 1
				fmt.Printf("BDS Multiple (serial): Jalur ditemukan (Panjang: %d), Nodes: %d.\n", len(path), nodes)
			} else if len(path) == 0 && !isBaseElement(targetElement) {
                 if err == nil {
                     err = fmt.Errorf("jalur kosong ditemukan untuk target non-dasar '%s'", targetElement)
                 }
            }
		}
		if err != nil {
			return nil, finalNodesVisited, fmt.Errorf("gagal menemukan jalur BDS tunggal: %v", err)
		}
		if foundCount == 0 && !isBaseElement(targetElement) {
			return nil, finalNodesVisited, fmt.Errorf("tidak ada jalur BDS (serial) yang ditemukan untuk '%s'", targetElement)
		}
		return allFoundPaths, finalNodesVisited, nil
	}

	fmt.Printf("BDS Multiple: Meluncurkan hingga %d goroutine untuk mencari jalur BDS...\n", numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		mu.Lock()
		if foundCount >= maxRecipes {
			mu.Unlock()
			break
		}
		mu.Unlock()

		wg.Add(1)
		go func(goroutineIndex int) {
			defer wg.Done()
			fmt.Printf("Goroutine BDS-%d: Memulai pencarian FindPathBDS.\n", goroutineIndex)
			path, _, err := FindPathBDS(targetElement) // Mengabaikan nodes dari goroutine individual

			mu.Lock()
			defer mu.Unlock()

			select {
			case <-quitChan:
				fmt.Printf("Goroutine BDS-%d: Menerima sinyal quit, hasil tidak ditambahkan.\n", goroutineIndex)
				return
			default:
			}

			if err == nil && path != nil {
				if len(path) > 0 {
					pathID := generatePathIdentifier(path) // Asumsi generatePathIdentifier ada di bfs.go
					if !addedPathIdentifiers[pathID] {
						if foundCount < maxRecipes {
							pathToAppend := make([]Recipe, len(path))
							copy(pathToAppend, path)
							allFoundPaths = append(allFoundPaths, pathToAppend)
							addedPathIdentifiers[pathID] = true
							foundCount++
							fmt.Printf("Goroutine BDS-%d: Jalur UNIK ditemukan (Panjang: %d). Total Ditemukan: %d\n", goroutineIndex, len(pathToAppend), foundCount)
							if foundCount >= maxRecipes {
								closeQuitChan()
							}
						}
					}
				} else if len(path) == 0 && !isBaseElement(targetElement) {
					// Jalur kosong untuk target non-dasar bisa diabaikan atau dicatat sebagai error jika tidak diharapkan
				}
			} else if err != nil {
				// fmt.Printf("Goroutine BDS-%d: Error mencari jalur: %v\n", goroutineIndex, err)
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("BDS Multiple: Semua goroutine pencarian BDS telah selesai.")

	mu.Lock()
	finalPathsToReturn := allFoundPaths
	if len(allFoundPaths) > maxRecipes {
		finalPathsToReturn = allFoundPaths[:maxRecipes]
	}
	currentFoundCount := len(finalPathsToReturn)
	mu.Unlock()

	if currentFoundCount == 0 && !isBaseElement(targetElement) {
		return nil, finalNodesVisited, fmt.Errorf("tidak ada jalur Bidirectional Search (multiple) yang valid ditemukan untuk '%s'", targetElement)
	}
	if isBaseElement(targetElement) && currentFoundCount == 1 && len(finalPathsToReturn[0]) == 0 {
		fmt.Printf("BDS Multiple: Selesai. Target adalah elemen dasar. Ditemukan: %d jalur (kosong).\n", currentFoundCount)
		return finalPathsToReturn, 0, nil
	}

	fmt.Printf("BDS Multiple: Selesai. Total jalur unik ditemukan: %d dari %d yang diminta.\n", currentFoundCount, maxRecipes)
	return finalPathsToReturn, finalNodesVisited, nil
}