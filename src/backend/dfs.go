// src/backend/dfs.go
package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

func FindPathDFS(targetElement string) ([]Recipe, int, error) {
	fmt.Printf("Mencari jalur DFS (single) ke: %s\n", targetElement)

	// Persiapan
	recipeMap := GetRecipeMap()
	if recipeMap == nil {
		return nil, 0, errors.New("map resep belum diinisialisasi")
	}

	if isBaseElementDFS(targetElement) {
		return []Recipe{}, 0, nil
	}

	nodesVisitedCount := 0

	knownCreatableElements := make(map[string]bool)
	for _, base := range []string{"Air", "Earth", "Fire", "Water"} {
		knownCreatableElements[base] = true
	}

	pathCache := make(map[string][]Recipe)

	var isCreatable func(element string, visited map[string]bool, depth int) bool
	isCreatable = func(element string, visited map[string]bool, depth int) bool {
		nodesVisitedCount++

		if depth > 500 {
			return false
		}

		if isBaseElementDFS(element) {
			return true
		}

		if known, exists := knownCreatableElements[element]; exists {
			return known
		}

		if depth > 30 && visited[element] {
			return false
		}

		newVisited := make(map[string]bool)
		for k, v := range visited {
			newVisited[k] = v
		}
		newVisited[element] = true

		recipes := recipeMap[element]
		if len(recipes) == 0 {
			knownCreatableElements[element] = false
			return false
		}

		for _, recipe := range recipes {
			ing1Creatable := isCreatable(recipe.Ingredient1, newVisited, depth+1)
			ing2Creatable := isCreatable(recipe.Ingredient2, newVisited, depth+1)

			if ing1Creatable && ing2Creatable {
				knownCreatableElements[element] = true
				return true
			}
		}

		knownCreatableElements[element] = false
		return false
	}

	var buildOrderedPath func(target string, availableElements map[string]bool, visited map[string]bool) []Recipe
	buildOrderedPath = func(target string, availableElements map[string]bool, visited map[string]bool) []Recipe {
		nodesVisitedCount++
		if isBaseElementDFS(target) || availableElements[target] {
			return []Recipe{}
		}

		if path, exists := pathCache[target]; exists {
			clonedAvailable := make(map[string]bool)
			for k, v := range availableElements {
				clonedAvailable[k] = v
			}

			valid := true
			for _, recipe := range path {
				if !isBaseElementDFS(recipe.Ingredient1) && !clonedAvailable[recipe.Ingredient1] {
					valid = false
					break
				}
				if !isBaseElementDFS(recipe.Ingredient2) && !clonedAvailable[recipe.Ingredient2] {
					valid = false
					break
				}
				clonedAvailable[recipe.Result] = true
			}

			if valid {
				pathCopy := make([]Recipe, len(path))
				copy(pathCopy, path)
				return pathCopy
			}
		}

		if visited[target] {
			return nil
		}

		newVisited := make(map[string]bool)
		for k, v := range visited {
			newVisited[k] = v
		}
		newVisited[target] = true

		recipes := recipeMap[target]
		if len(recipes) == 0 {
			return nil
		}

		sort.Slice(recipes, func(i, j int) bool {
			iCanMake := (isBaseElementDFS(recipes[i].Ingredient1) || availableElements[recipes[i].Ingredient1]) &&
				(isBaseElementDFS(recipes[i].Ingredient2) || availableElements[recipes[i].Ingredient2])
			jCanMake := (isBaseElementDFS(recipes[j].Ingredient1) || availableElements[recipes[j].Ingredient1]) &&
				(isBaseElementDFS(recipes[j].Ingredient2) || availableElements[recipes[j].Ingredient2])

			if iCanMake && !jCanMake {
				return true
			}
			if !iCanMake && jCanMake {
				return false
			}

			iBaseCount := 0
			jBaseCount := 0

			if isBaseElementDFS(recipes[i].Ingredient1) {
				iBaseCount++
			}
			if isBaseElementDFS(recipes[i].Ingredient2) {
				iBaseCount++
			}
			if isBaseElementDFS(recipes[j].Ingredient1) {
				jBaseCount++
			}
			if isBaseElementDFS(recipes[j].Ingredient2) {
				jBaseCount++
			}

			if iBaseCount != jBaseCount {
				return iBaseCount > jBaseCount
			}

			return recipes[i].Result < recipes[j].Result // Stabil sort
		})

		var bestPath []Recipe

		for _, recipe := range recipes {
			elementsAvailable := make(map[string]bool)
			for k, v := range availableElements {
				elementsAvailable[k] = v
			}

			var path1 []Recipe
			if !isBaseElementDFS(recipe.Ingredient1) && !elementsAvailable[recipe.Ingredient1] {
				path1 = buildOrderedPath(recipe.Ingredient1, elementsAvailable, newVisited)
				if path1 == nil {
					continue
				}

				for _, p := range path1 {
					elementsAvailable[p.Result] = true
				}
			}

			var path2 []Recipe
			if !isBaseElementDFS(recipe.Ingredient2) && !elementsAvailable[recipe.Ingredient2] {
				path2 = buildOrderedPath(recipe.Ingredient2, elementsAvailable, newVisited)
				if path2 == nil {
					continue
				}

				for _, p := range path2 {
					elementsAvailable[p.Result] = true
				}
			}

			if (!isBaseElementDFS(recipe.Ingredient1) && !elementsAvailable[recipe.Ingredient1]) ||
				(!isBaseElementDFS(recipe.Ingredient2) && !elementsAvailable[recipe.Ingredient2]) {
				continue
			}

			completePath := make([]Recipe, 0)

			if path1 != nil {
				completePath = append(completePath, path1...)
			}

			if path2 != nil {
				completePath = append(completePath, path2...)
			}

			completePath = append(completePath, recipe)

			if bestPath == nil || len(completePath) < len(bestPath) {
				bestPath = completePath
			}
		}

		if bestPath != nil {
			pathCopy := make([]Recipe, len(bestPath))
			copy(pathCopy, bestPath)
			pathCache[target] = pathCopy
		}

		return bestPath
	}

	var removeDuplicateRecipes = func(path []Recipe) []Recipe {
		seen := make(map[string]bool)
		unique := make([]Recipe, 0, len(path))

		for _, recipe := range path {
			key := fmt.Sprintf("%s:%s+%s", recipe.Result, recipe.Ingredient1, recipe.Ingredient2)
			if !seen[key] {
				seen[key] = true
				unique = append(unique, recipe)
			}
		}

		return unique
	}

	availableElements := make(map[string]bool)
	for _, base := range []string{"Air", "Earth", "Fire", "Water"} {
		availableElements[base] = true
	}

	fmt.Printf("Mencari jalur optimal untuk %s...\n", targetElement)
	optimalPath := buildOrderedPath(targetElement, availableElements, make(map[string]bool))

	if optimalPath == nil {
		return nil, nodesVisitedCount, fmt.Errorf("tidak ada jalur valid untuk membuat %s", targetElement)
	}

	optimalPath = removeDuplicateRecipes(optimalPath)

	available := make(map[string]bool)
	for _, base := range []string{"Air", "Earth", "Fire", "Water"} {
		available[base] = true
	}

	for i, recipe := range optimalPath {
		if !isBaseElementDFS(recipe.Ingredient1) && !available[recipe.Ingredient1] {
			fmt.Printf("PERINGATAN: Jalur optimal - bahan %s tidak tersedia pada langkah %d\n",
				recipe.Ingredient1, i+1)
		}

		if !isBaseElementDFS(recipe.Ingredient2) && !available[recipe.Ingredient2] {
			fmt.Printf("PERINGATAN: Jalur optimal - bahan %s tidak tersedia pada langkah %d\n",
				recipe.Ingredient2, i+1)
		}

		available[recipe.Result] = true
	}

	fmt.Printf("Jalur DFS (single) (panjang: %d):\n", len(optimalPath))
	for i, recipe := range optimalPath {
		fmt.Printf("  Langkah %d: %s + %s => %s\n",
			i+1, recipe.Ingredient1, recipe.Ingredient2, recipe.Result)
	}

	return optimalPath, nodesVisitedCount, nil
}

func FindMultiplePathsDFS(targetElement string, maxRecipes int) ([][]Recipe, int, error) {
	fmt.Printf("Mencari %d jalur DFS BERBEDA ke: %s dengan multithreading (Super Robust)\n", maxRecipes, targetElement)

	recipeMap := GetRecipeMap()
	if recipeMap == nil {
		return nil, 0, errors.New("map resep belum diinisialisasi")
	}
	if maxRecipes <= 0 {
		return nil, 0, errors.New("jumlah resep minimal harus 1")
	}
	if isBaseElementDFS(targetElement) {
		return [][]Recipe{}, 0, nil
	}

	nodesVisitedCount := 0

	knownCreatableElements := make(map[string]bool)
	for _, base := range []string{"Air", "Earth", "Fire", "Water"} {
		knownCreatableElements[base] = true
	}
	var knownCreatableMutex sync.RWMutex

	pathCache := make(map[string][]Recipe)
	var pathCacheMutex sync.RWMutex

	var isCreatable func(element string, visited map[string]bool, depth int) bool
	isCreatable = func(element string, visited map[string]bool, depth int) bool {
		nodesVisitedCount++

		if depth > 500 {
			return false
		}

		if isBaseElementDFS(element) {
			return true
		}

		knownCreatableMutex.RLock()
		if known, exists := knownCreatableElements[element]; exists {
			knownCreatableMutex.RUnlock()
			return known
		}
		knownCreatableMutex.RUnlock()

		if depth > 30 && visited[element] {
			return false
		}

		newVisited := make(map[string]bool)
		for k, v := range visited {
			newVisited[k] = v
		}
		newVisited[element] = true

		recipes := recipeMap[element]
		if len(recipes) == 0 {
			knownCreatableMutex.Lock()
			knownCreatableElements[element] = false
			knownCreatableMutex.Unlock()
			return false
		}

		for _, recipe := range recipes {
			ing1Creatable := isCreatable(recipe.Ingredient1, newVisited, depth+1)
			ing2Creatable := isCreatable(recipe.Ingredient2, newVisited, depth+1)

			if ing1Creatable && ing2Creatable {
				knownCreatableMutex.Lock()
				knownCreatableElements[element] = true
				knownCreatableMutex.Unlock()
				return true
			}
		}

		knownCreatableMutex.Lock()
		knownCreatableElements[element] = false
		knownCreatableMutex.Unlock()
		return false
	}

	var buildOrderedPath func(target string, availableElements map[string]bool, visited map[string]bool) []Recipe
	buildOrderedPath = func(target string, availableElements map[string]bool, visited map[string]bool) []Recipe {
		nodesVisitedCount++
		if isBaseElementDFS(target) || availableElements[target] {
			return []Recipe{}
		}

		pathCacheMutex.RLock()
		if path, exists := pathCache[target]; exists {
			pathCacheMutex.RUnlock()

			clonedAvailable := make(map[string]bool)
			for k, v := range availableElements {
				clonedAvailable[k] = v
			}

			valid := true
			for _, recipe := range path {
				if !isBaseElementDFS(recipe.Ingredient1) && !clonedAvailable[recipe.Ingredient1] {
					valid = false
					break
				}
				if !isBaseElementDFS(recipe.Ingredient2) && !clonedAvailable[recipe.Ingredient2] {
					valid = false
					break
				}
				clonedAvailable[recipe.Result] = true
			}

			if valid {
				pathCopy := make([]Recipe, len(path))
				copy(pathCopy, path)
				return pathCopy
			}
		} else {
			pathCacheMutex.RUnlock()
		}

		if visited[target] {
			return nil
		}

		newVisited := make(map[string]bool)
		for k, v := range visited {
			newVisited[k] = v
		}
		newVisited[target] = true

		recipes := recipeMap[target]
		if len(recipes) == 0 {
			return nil
		}

		sort.Slice(recipes, func(i, j int) bool {
			iCanMake := (isBaseElementDFS(recipes[i].Ingredient1) || availableElements[recipes[i].Ingredient1]) &&
				(isBaseElementDFS(recipes[i].Ingredient2) || availableElements[recipes[i].Ingredient2])
			jCanMake := (isBaseElementDFS(recipes[j].Ingredient1) || availableElements[recipes[j].Ingredient1]) &&
				(isBaseElementDFS(recipes[j].Ingredient2) || availableElements[recipes[j].Ingredient2])

			if iCanMake && !jCanMake {
				return true
			}
			if !iCanMake && jCanMake {
				return false
			}

			iBaseCount := 0
			jBaseCount := 0

			if isBaseElementDFS(recipes[i].Ingredient1) {
				iBaseCount++
			}
			if isBaseElementDFS(recipes[i].Ingredient2) {
				iBaseCount++
			}
			if isBaseElementDFS(recipes[j].Ingredient1) {
				jBaseCount++
			}
			if isBaseElementDFS(recipes[j].Ingredient2) {
				jBaseCount++
			}

			if iBaseCount != jBaseCount {
				return iBaseCount > jBaseCount
			}

			return recipes[i].Result < recipes[j].Result
		})

		var bestPath []Recipe

		for _, recipe := range recipes {
			elementsAvailable := make(map[string]bool)
			for k, v := range availableElements {
				elementsAvailable[k] = v
			}
			var path1 []Recipe
			if !isBaseElementDFS(recipe.Ingredient1) && !elementsAvailable[recipe.Ingredient1] {
				path1 = buildOrderedPath(recipe.Ingredient1, elementsAvailable, newVisited)
				if path1 == nil {
					continue
				}

				for _, p := range path1 {
					elementsAvailable[p.Result] = true
				}
			}

			var path2 []Recipe
			if !isBaseElementDFS(recipe.Ingredient2) && !elementsAvailable[recipe.Ingredient2] {
				path2 = buildOrderedPath(recipe.Ingredient2, elementsAvailable, newVisited)
				if path2 == nil {
					continue
				}

				for _, p := range path2 {
					elementsAvailable[p.Result] = true
				}
			}

			if (!isBaseElementDFS(recipe.Ingredient1) && !elementsAvailable[recipe.Ingredient1]) ||
				(!isBaseElementDFS(recipe.Ingredient2) && !elementsAvailable[recipe.Ingredient2]) {
				continue
			}

			completePath := make([]Recipe, 0)

			if path1 != nil {
				completePath = append(completePath, path1...)
			}

			if path2 != nil {
				completePath = append(completePath, path2...)
			}

			completePath = append(completePath, recipe)

			if bestPath == nil || len(completePath) < len(bestPath) {
				bestPath = completePath
			}
		}

		if bestPath != nil {
			pathCopy := make([]Recipe, len(bestPath))
			copy(pathCopy, bestPath)

			pathCacheMutex.Lock()
			pathCache[target] = pathCopy
			pathCacheMutex.Unlock()
		}

		return bestPath
	}

	var removeDuplicateRecipes = func(path []Recipe) []Recipe {
		seen := make(map[string]bool)
		unique := make([]Recipe, 0, len(path))

		for _, recipe := range path {
			key := fmt.Sprintf("%s:%s+%s", recipe.Result, recipe.Ingredient1, recipe.Ingredient2)
			if !seen[key] {
				seen[key] = true
				unique = append(unique, recipe)
			}
		}

		return unique
	}

	var findAlternativePaths = func(target string, existingPath []Recipe, maxPaths int) [][]Recipe {
		results := [][]Recipe{existingPath}
		uniquePathMap := make(map[string]bool)

		existingPathID := generatePathIdentifierDFS(existingPath)
		uniquePathMap[existingPathID] = true

		var wg sync.WaitGroup
		semaphore := make(chan struct{}, 8)
		var resultsMutex sync.Mutex

		recipesForTarget := recipeMap[target]

		for _, recipe := range recipesForTarget {
			if len(existingPath) > 0 {
				lastRecipe := existingPath[len(existingPath)-1]
				if lastRecipe.Result == recipe.Result &&
					lastRecipe.Ingredient1 == recipe.Ingredient1 &&
					lastRecipe.Ingredient2 == recipe.Ingredient2 {
					continue
				}
			}

			wg.Add(1)
			go func(r Recipe) {
				semaphore <- struct{}{}
				defer func() {
					<-semaphore
					wg.Done()
				}()

				availableElements := make(map[string]bool)
				for _, base := range []string{"Air", "Earth", "Fire", "Water"} {
					availableElements[base] = true
				}

				if !isCreatable(r.Ingredient1, make(map[string]bool), 0) ||
					!isCreatable(r.Ingredient2, make(map[string]bool), 0) {
					return
				}

				var completePath []Recipe

				if !isBaseElementDFS(r.Ingredient1) {
					ing1Path := buildOrderedPath(r.Ingredient1, availableElements, make(map[string]bool))
					if ing1Path == nil {
						return
					}

					completePath = append(completePath, ing1Path...)

					for _, p := range ing1Path {
						availableElements[p.Result] = true
					}
				}

				if !isBaseElementDFS(r.Ingredient2) && !availableElements[r.Ingredient2] {
					ing2Path := buildOrderedPath(r.Ingredient2, availableElements, make(map[string]bool))
					if ing2Path == nil {
						return
					}

					completePath = append(completePath, ing2Path...)

					for _, p := range ing2Path {
						availableElements[p.Result] = true
					}
				}

				completePath = append(completePath, r)
				finalPath := removeDuplicateRecipes(completePath)

				available := make(map[string]bool)
				for _, base := range []string{"Air", "Earth", "Fire", "Water"} {
					available[base] = true
				}

				valid := true
				for _, recipe := range finalPath {
					if !isBaseElementDFS(recipe.Ingredient1) && !available[recipe.Ingredient1] {
						valid = false
						break
					}

					if !isBaseElementDFS(recipe.Ingredient2) && !available[recipe.Ingredient2] {
						valid = false
						break
					}

					available[recipe.Result] = true
				}

				if !valid {
					return
				}

				pathID := generatePathIdentifierDFS(finalPath)

				resultsMutex.Lock()
				defer resultsMutex.Unlock()

				if !uniquePathMap[pathID] && len(results) < maxPaths {
					uniquePathMap[pathID] = true
					results = append(results, finalPath)
				}
			}(recipe)
		}

		wg.Wait()

		return results
	}

	availableElements := make(map[string]bool)
	for _, base := range []string{"Air", "Earth", "Fire", "Water"} {
		availableElements[base] = true
	}

	fmt.Printf("Mencari jalur optimal untuk %s...\n", targetElement)
	optimalPath := buildOrderedPath(targetElement, availableElements, make(map[string]bool))

	if optimalPath == nil {
		return nil, nodesVisitedCount, fmt.Errorf("tidak ada jalur valid untuk membuat %s", targetElement)
	}

	optimalPath = removeDuplicateRecipes(optimalPath)

	available := make(map[string]bool)
	for _, base := range []string{"Air", "Earth", "Fire", "Water"} {
		available[base] = true
	}

	for i, recipe := range optimalPath {
		if !isBaseElementDFS(recipe.Ingredient1) && !available[recipe.Ingredient1] {
			fmt.Printf("PERINGATAN: Jalur optimal - bahan %s tidak tersedia pada langkah %d\n",
				recipe.Ingredient1, i+1)
		}

		if !isBaseElementDFS(recipe.Ingredient2) && !available[recipe.Ingredient2] {
			fmt.Printf("PERINGATAN: Jalur optimal - bahan %s tidak tersedia pada langkah %d\n",
				recipe.Ingredient2, i+1)
		}

		available[recipe.Result] = true
	}

	fmt.Printf("Jalur optimal (panjang: %d):\n", len(optimalPath))
	for i, recipe := range optimalPath {
		fmt.Printf("  Langkah %d: %s + %s => %s\n",
			i+1, recipe.Ingredient1, recipe.Ingredient2, recipe.Result)
	}

	if maxRecipes <= 1 {
		return [][]Recipe{optimalPath}, nodesVisitedCount, nil
	}

	fmt.Printf("Mencari %d jalur alternatif...\n", maxRecipes-1)
	allPaths := findAlternativePaths(targetElement, optimalPath, maxRecipes)

	sort.Slice(allPaths, func(i, j int) bool {
		return len(allPaths[i]) < len(allPaths[j])
	})

	for i, path := range allPaths {
		fmt.Printf("Jalur %d (panjang: %d):\n", i+1, len(path))
		for j, recipe := range path {
			fmt.Printf("  Langkah %d: %s + %s => %s\n",
				j+1, recipe.Ingredient1, recipe.Ingredient2, recipe.Result)
		}
	}

	return allPaths, nodesVisitedCount, nil
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

func generatePathIdentifierDFS(path []Recipe) string {
	recipesCopy := make([]Recipe, len(path))
	copy(recipesCopy, path)
	sort.Slice(recipesCopy, func(i, j int) bool {
		if recipesCopy[i].Result != recipesCopy[j].Result {
			return recipesCopy[i].Result < recipesCopy[j].Result
		}
		ing1i, ing2i := recipesCopy[i].Ingredient1, recipesCopy[i].Ingredient2
		if ing1i > ing2i {
			ing1i, ing2i = ing2i, ing1i
		}
		ing1j, ing2j := recipesCopy[j].Ingredient1, recipesCopy[j].Ingredient2
		if ing1j > ing2j {
			ing1j, ing2j = ing2j, ing1j
		}
		if ing1i != ing1j {
			return ing1i < ing1j
		}
		return ing2i < ing2j
	})
	var parts []string
	for _, r := range recipesCopy {
		ing1, ing2 := r.Ingredient1, r.Ingredient2
		if ing1 > ing2 {
			ing1, ing2 = ing2, ing1
		}
		parts = append(parts, fmt.Sprintf("%s+%s=>%s", ing1, ing2, r.Result))
	}
	return strings.Join(parts, "|")
}
