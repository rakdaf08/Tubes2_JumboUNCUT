package main

import (
	"container/list"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

func reconstructSingleSegmentPath(parentMap map[string]Recipe, startNode string, stopCondition func(string) bool) []Recipe {
	pathList := list.New()
	processed := make(map[string]bool)
	curr := startNode
	for curr != "" && !stopCondition(curr) {
		recipe, exists := parentMap[curr]
		if !exists {
			break
		}
		recipeKey := getUniqueRecipeKey(recipe)
		if processed[recipeKey] {
			break
		}
		pathList.PushFront(recipe)
		processed[recipeKey] = true
		_, p1Exists := parentMap[recipe.Ingredient1]
		_, p2Exists := parentMap[recipe.Ingredient2]
		chosenParent := ""
		if p1Exists && p2Exists {
			if recipe.Ingredient1 < recipe.Ingredient2 {
				chosenParent = recipe.Ingredient1
			} else {
				chosenParent = recipe.Ingredient2
			}
		} else if p1Exists {
			chosenParent = recipe.Ingredient1
		} else if p2Exists {
			chosenParent = recipe.Ingredient2
		} else {
			if isBaseElement(recipe.Ingredient1) && stopCondition(recipe.Ingredient1) {
				chosenParent = recipe.Ingredient1
			} else if isBaseElement(recipe.Ingredient2) && stopCondition(recipe.Ingredient2) {
				chosenParent = recipe.Ingredient2
			} else if isBaseElement(recipe.Ingredient1) && !stopCondition(recipe.Ingredient1) {
				chosenParent = recipe.Ingredient1
			} else if isBaseElement(recipe.Ingredient2) && !stopCondition(recipe.Ingredient2) {
				chosenParent = recipe.Ingredient2
			} else {
				chosenParent = ""
			}
		}
		curr = chosenParent
	}
	finalPath := make([]Recipe, 0, pathList.Len())
	for e := pathList.Front(); e != nil; e = e.Next() {
		finalPath = append(finalPath, e.Value.(Recipe))
	}
	return finalPath
}

func buildSortedPathFromRecipes(recipes map[string]Recipe, targetElement string) []Recipe {
	fmt.Println("  Mengurutkan resep gabungan berdasarkan dependensi...")
	if len(recipes) == 0 {
		return []Recipe{}
	}

	elementsInvolved := make(map[string]bool)
	for _, r := range recipes {
		elementsInvolved[r.Ingredient1] = true
		elementsInvolved[r.Ingredient2] = true
		elementsInvolved[r.Result] = true
	}

	available := make(map[string]bool)
	for _, base := range baseElements {
		if elementsInvolved[base] {
			available[base] = true
		}
	}

	remainingRecipes := make(map[string]Recipe)
	for k, v := range recipes {
		remainingRecipes[k] = v
	}

	sortedPath := make([]Recipe, 0, len(recipes))
	iterations := 0
	maxIterations := len(recipes)*2 + 10

	for !available[targetElement] && iterations < maxIterations {
		addedRecipeInIteration := false
		candidates := make([]Recipe, 0)
		candidateKeys := make([]string, 0)

		for key, recipe := range remainingRecipes {
			if available[recipe.Ingredient1] && available[recipe.Ingredient2] {
				candidates = append(candidates, recipe)
				candidateKeys = append(candidateKeys, key)
			}
		}

		if len(candidates) == 0 {
			fmt.Printf("  ERROR (Sort): Tidak ada kandidat resep yang bisa dibuat, target '%s' belum tersedia. Elemen tersedia: %v\n", targetElement, available)
			return sortedPath
		}

		sort.SliceStable(candidates, func(i, j int) bool {
			return candidates[i].Result < candidates[j].Result
		})

		for i, recipe := range candidates {
			key := candidateKeys[i]
			sortedPath = append(sortedPath, recipe)
			available[recipe.Result] = true
			delete(remainingRecipes, key)
			addedRecipeInIteration = true
		}

		if !addedRecipeInIteration && !available[targetElement] {
			fmt.Printf("  ERROR (Sort): Tidak ada resep baru ditambahkan di iterasi %d, target '%s' belum tersedia.\n", iterations+1, targetElement)
			return sortedPath
		}
		iterations++
	}

	if iterations >= maxIterations {
		fmt.Printf("  ERROR (Sort): Melebihi batas iterasi maksimum (%d), target '%s' mungkin tidak tercapai atau ada loop dependensi.\n", maxIterations, targetElement)
	} else if !available[targetElement] {
		fmt.Printf("  PERINGATAN (Sort): Loop selesai, tapi target '%s' tidak tersedia di akhir.\n", targetElement)
	} else {
		fmt.Printf("  Pengurutan resep selesai. Total langkah terurut: %d\n", len(sortedPath))
	}

	return sortedPath
}

func FindPathBDS(targetElement string) ([]Recipe, int, error) {
	fmt.Printf("Hybrid BDS+BFS: Mencari jalur ke: %s\n", targetElement)
	recipeMap := GetRecipeMap()
	alchemyGraph := GetAlchemyGraph()
	if recipeMap == nil || alchemyGraph == nil {
		return nil, 0, errors.New("data resep/graf belum diinisialisasi")
	}
	if isBaseElement(targetElement) {
		return []Recipe{}, 0, nil
	}

	nodesVisitedCount := 0
	queueForward := list.New()
	visitedForward := make(map[string]int)
	parentForward := make(map[string]Recipe)
	queueBackward := list.New()
	visitedBackward := make(map[string]int)
	parentBackward := make(map[string]Recipe)

	// Inisialisasi
	for _, base := range baseElements {
		if visitedForward[base] == 0 {
			queueForward.PushBack(base)
			visitedForward[base] = 1
		}
	}
	if visitedBackward[targetElement] == 0 {
		queueBackward.PushBack(targetElement)
		visitedBackward[targetElement] = 1
	}

	currentLevelForward := 1
	currentLevelBackward := 1
	var meetingNode string = ""

	for queueForward.Len() > 0 && queueBackward.Len() > 0 && meetingNode == "" {

		lenF := queueForward.Len()
		for i := 0; i < lenF && meetingNode == ""; i++ {
			if queueForward.Len() == 0 {
				break
			}
			currF := queueForward.Remove(queueForward.Front()).(string)
			nodesVisitedCount++

			if visitedBackward[currF] > 0 {
				meetingNode = currF
			}

			recipesUsingCurrF := alchemyGraph[currF]
			for _, recipe := range recipesUsingCurrF {
				otherIng := ""
				if recipe.Ingredient1 == currF {
					otherIng = recipe.Ingredient2
				} else if recipe.Ingredient2 == currF {
					otherIng = recipe.Ingredient1
				} else {
					continue
				}

				if visitedForward[otherIng] > 0 && visitedForward[otherIng] <= currentLevelForward {
					result := recipe.Result
					if visitedForward[result] == 0 {
						visitedForward[result] = currentLevelForward + 1
						parentForward[result] = recipe
						queueForward.PushBack(result)

						if visitedBackward[result] > 0 && meetingNode == "" {
							meetingNode = result
						}
					}
				}
			}
		}
		if meetingNode != "" {
			break
		}
		currentLevelForward++

		lenB := queueBackward.Len()
		for i := 0; i < lenB && meetingNode == ""; i++ {
			if queueBackward.Len() == 0 {
				break
			}
			currB := queueBackward.Remove(queueBackward.Front()).(string)
			nodesVisitedCount++

			if visitedForward[currB] > 0 {
				meetingNode = currB
			}

			recipesMakingCurrB := recipeMap[currB]
			if _, exists := parentBackward[currB]; !exists && len(recipesMakingCurrB) > 0 {
				parentBackward[currB] = recipesMakingCurrB[0]
			}

			for _, recipe := range recipesMakingCurrB {
				ingredients := []string{recipe.Ingredient1, recipe.Ingredient2}
				for _, ing := range ingredients {
					if visitedBackward[ing] == 0 {
						visitedBackward[ing] = currentLevelBackward + 1
						queueBackward.PushBack(ing)

						if visitedForward[ing] > 0 && meetingNode == "" {
							meetingNode = ing
						}
					}
				}
			}
		}
		if meetingNode != "" {
			break
		}
		currentLevelBackward++

		if queueForward.Len() == 0 || queueBackward.Len() == 0 {
			break
		}
	}

	if meetingNode == "" {
		fmt.Printf("Hybrid BDS+BFS: Tidak ada pertemuan ditemukan untuk '%s'.\n", targetElement)
		return nil, nodesVisitedCount, fmt.Errorf("jalur (BDS meeting) ke '%s' tidak ditemukan", targetElement)
	}

	fmt.Printf("Hybrid BDS+BFS: Pertemuan di '%s'. Memulai rekonstruksi dan pencarian BFS tambahan...\n", meetingNode)

	finalRecipe, finalRecipeExists := parentBackward[targetElement]
	if !finalRecipeExists {
		recipesForTarget := recipeMap[targetElement]
		if len(recipesForTarget) > 0 {
			finalRecipe = recipesForTarget[0]
			finalRecipeExists = true
		} else {
			fmt.Printf("  ERROR: Tidak dapat menemukan resep final untuk '%s'.\n", targetElement)
			return nil, nodesVisitedCount, fmt.Errorf("resep final untuk '%s' tidak ditemukan", targetElement)
		}
	}

	ing1 := finalRecipe.Ingredient1
	ing2 := finalRecipe.Ingredient2

	combinedRecipes := make(map[string]Recipe)

	if meetingNode == ing1 || meetingNode == ing2 {
		var ingredientToSearchBFS string
		var pathForMeetingNodeSegment []Recipe

		if meetingNode == ing1 {
			ingredientToSearchBFS = ing2
		} else {
			ingredientToSearchBFS = ing1
		}

		fmt.Printf("  Merekonstruksi jalur FWD untuk meeting node '%s'...\n", meetingNode)
		stopAtBase := func(node string) bool { return isBaseElement(node) }
		pathForMeetingNodeSegment = reconstructSingleSegmentPath(parentForward, meetingNode, stopAtBase)
		fmt.Printf("  Jalur FWD untuk '%s' ditemukan (panjang: %d)\n", meetingNode, len(pathForMeetingNodeSegment))

		fmt.Printf("  Mencari jalur BFS untuk bahan '%s'\n", ingredientToSearchBFS)
		pathOtherIngredient, bfsNodes, errBFS := FindPathBFS(ingredientToSearchBFS)
		if errBFS != nil {
			fmt.Printf("  ERROR: Gagal mencari jalur BFS untuk '%s': %v\n", ingredientToSearchBFS, errBFS)
			return nil, nodesVisitedCount + bfsNodes, fmt.Errorf("gagal mencari jalur BFS untuk bahan '%s': %v", ingredientToSearchBFS, errBFS)
		}
		nodesVisitedCount += bfsNodes
		fmt.Printf("  Jalur BFS untuk '%s' ditemukan (panjang: %d)\n", ingredientToSearchBFS, len(pathOtherIngredient))

		for _, r := range pathForMeetingNodeSegment {
			combinedRecipes[getUniqueRecipeKey(r)] = r
		}
		for _, r := range pathOtherIngredient {
			combinedRecipes[getUniqueRecipeKey(r)] = r
		}

	} else {
		fmt.Printf("  PERINGATAN: Meeting node '%s' bukan bahan langsung. Mencari BFS untuk KEDUA bahan '%s' dan '%s'.\n", meetingNode, ing1, ing2)

		fmt.Printf("  Mencari jalur BFS untuk bahan 1: '%s'\n", ing1)
		pathIng1, bfsNodes1, err1 := FindPathBFS(ing1)
		if err1 != nil {
			fmt.Printf("  ERROR: Gagal mencari jalur BFS untuk '%s': %v\n", ing1, err1)
			return nil, nodesVisitedCount + bfsNodes1, fmt.Errorf("gagal mencari jalur BFS untuk bahan '%s': %v", ing1, err1)
		}
		nodesVisitedCount += bfsNodes1
		fmt.Printf("  Jalur BFS untuk '%s' ditemukan (panjang: %d)\n", ing1, len(pathIng1))
		for _, r := range pathIng1 {
			combinedRecipes[getUniqueRecipeKey(r)] = r
		}

		fmt.Printf("  Mencari jalur BFS untuk bahan 2: '%s'\n", ing2)
		pathIng2, bfsNodes2, err2 := FindPathBFS(ing2)
		if err2 != nil {
			fmt.Printf("  ERROR: Gagal mencari jalur BFS untuk '%s': %v\n", ing2, err2)
			return nil, nodesVisitedCount + bfsNodes2, fmt.Errorf("gagal mencari jalur BFS untuk bahan '%s': %v", ing2, err2)
		}
		nodesVisitedCount += bfsNodes2
		fmt.Printf("  Jalur BFS untuk '%s' ditemukan (panjang: %d)\n", ing2, len(pathIng2))
		for _, r := range pathIng2 {
			combinedRecipes[getUniqueRecipeKey(r)] = r
		}

		fmt.Printf("  Merekonstruksi jalur FWD untuk meeting node '%s' (kasus 2)...\n", meetingNode)
		stopAtBase := func(node string) bool { return isBaseElement(node) }
		pathMeetingToBase := reconstructSingleSegmentPath(parentForward, meetingNode, stopAtBase)
		fmt.Printf("  Jalur FWD untuk '%s' ditemukan (panjang: %d)\n", meetingNode, len(pathMeetingToBase))
		for _, r := range pathMeetingToBase {
			combinedRecipes[getUniqueRecipeKey(r)] = r
		}
	}

	combinedRecipes[getUniqueRecipeKey(finalRecipe)] = finalRecipe

	finalPathSorted := buildSortedPathFromRecipes(combinedRecipes, targetElement)

	if len(finalPathSorted) == 0 && !isBaseElement(targetElement) {
		fmt.Printf("  PERINGATAN AKHIR: Jalur terurut kosong untuk target non-dasar '%s'.\n", targetElement)
	} else if len(finalPathSorted) > 0 && finalPathSorted[len(finalPathSorted)-1].Result != targetElement {
		fmt.Printf("  PERINGATAN AKHIR: Jalur terurut TIDAK menghasilkan target '%s'. Resep terakhir: %v\n", targetElement, finalPathSorted[len(finalPathSorted)-1].Result)
	}

	fmt.Printf("Hybrid BDS+BFS: Penggabungan dan pengurutan selesai. Total resep unik terurut: %d\n", len(finalPathSorted))
	return finalPathSorted, nodesVisitedCount, nil
}

func FindMultiplePathsBDS(targetElement string, maxRecipes int) ([][]Recipe, int, error) {
	fmt.Printf("BDS Multiple (Hybrid): Mencari %d jalur ke: %s (Multithreaded)\n", maxRecipes, targetElement)

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
	nodesVisitedTotal := atomic.Int32{}
	foundCount := atomic.Int32{}
	quitChan := make(chan struct{})
	var quitOnce sync.Once
	closeQuitChan := func() {
		quitOnce.Do(func() { close(quitChan) })
	}
	defer closeQuitChan()

	numGoroutines := maxRecipes
	if numGoroutines < 1 {
		numGoroutines = 1
	}
	maxGo := 10
	if numGoroutines > maxGo {
		numGoroutines = maxGo
	}

	fmt.Printf("BDS Multiple (Hybrid): Meluncurkan %d goroutine...\n", numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		if foundCount.Load() >= int32(maxRecipes) {
			break
		}
		wg.Add(1)
		go func(goroutineIndex int) {
			defer wg.Done()

			path, nodesVisited, err := FindPathBDS(targetElement)
			nodesVisitedTotal.Add(int32(nodesVisited))
			mu.Lock()
			defer mu.Unlock()
			select {
			case <-quitChan:
				return
			default:
			}
			if err == nil && path != nil {
				if len(path) > 0 {
					pathID := generatePathIdentifier(path)
					if !addedPathIdentifiers[pathID] {
						if currentFound := foundCount.Load(); currentFound < int32(maxRecipes) {
							pathToAppend := make([]Recipe, len(path))
							copy(pathToAppend, path)
							allFoundPaths = append(allFoundPaths, pathToAppend)
							addedPathIdentifiers[pathID] = true
							newCount := foundCount.Add(1)
							fmt.Printf("Goroutine Hybrid-%d: Jalur UNIK ditemukan (Panjang: %d). Total Ditemukan: %d/%d\n", goroutineIndex, len(pathToAppend), newCount, maxRecipes)
							if newCount >= int32(maxRecipes) {
								closeQuitChan()
							}
						}
					}
				}
			}
		}(i)
	}

	wg.Wait()
	mu.Lock()
	finalPathsToReturn := make([][]Recipe, len(allFoundPaths))
	copy(finalPathsToReturn, allFoundPaths)
	currentFoundCount := len(finalPathsToReturn)
	mu.Unlock()

	if currentFoundCount == 0 && !isBaseElement(targetElement) {
		return nil, int(nodesVisitedTotal.Load()), fmt.Errorf("tidak ada jalur Hybrid BDS+BFS (multiple) yang valid ditemukan untuk '%s'", targetElement)
	}

	sort.SliceStable(finalPathsToReturn, func(i, j int) bool {
		return len(finalPathsToReturn[i]) < len(finalPathsToReturn[j])
	})

	fmt.Printf("BDS Multiple (Hybrid): Selesai. Total jalur unik ditemukan: %d (diminta: %d). Total nodes visited (approx): %d\n", currentFoundCount, maxRecipes, nodesVisitedTotal.Load())
	return finalPathsToReturn, int(nodesVisitedTotal.Load()), nil
}
