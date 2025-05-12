import React, { useState, useEffect, useCallback } from 'react';
import Tree from 'react-d3-tree';
import './SearchResults.css'; // Pastikan file CSS ini ada dan sesuai
import notFoundImage from '../assets/notfound_dontol.jpg'; // Pastikan path gambar ini benar

const API_BASE_URL = "http://localhost:8080"; // Sesuaikan jika perlu
const LIVE_UPDATE_DELAY_MS = 800; // Kecepatan animasi live update

// Helper function to check if an element is a base element
const isBaseElement = (name) => {
    const baseElements = ["Air", "Earth", "Fire", "Water"]; // Daftar elemen dasar
    return baseElements.includes(name);
};

// Helper function to build the initial node structure for an element
const buildInitialElementNode = (elementName, imageURLs, depth = 0) => {
    const imageUrlPath = imageURLs && typeof imageURLs === 'object' ? (imageURLs[elementName] || '') : '';
    return {
      name: elementName || "Unknown",
      attributes: {
        type: isBaseElement(elementName) ? 'Base Element' : 'Element',
        imageUrl: imageUrlPath,
        depth: depth,
        originalName: elementName || "Unknown",
      },
      children: [],
    };
};

// Recursively builds the full static tree data from a path map
const buildElementNodeRecursive = (elementName, pathRecipesMap, imageURLs, currentDepth = 0, maxDepth = 20) => {
    // Batasi kedalaman rekursi untuk mencegah infinite loop atau pohon yang terlalu besar
    if (currentDepth > maxDepth) {
        return {
            name: `${elementName || "Unknown"} (Batas Kedalaman)`,
            attributes: { type: 'Error', originalName: elementName || "Unknown", depth: currentDepth, info: 'Max depth reached' },
            children: []
        };
    }

    // Buat node awal untuk elemen saat ini
    const node = buildInitialElementNode(elementName, imageURLs, currentDepth);
    const recipeMakingThis = pathRecipesMap && typeof pathRecipesMap === 'object' ? pathRecipesMap[elementName] : null;

    // Jika ada resep yang menghasilkan elemen ini, tambahkan node resep dan rekursif ke bawah
    if (recipeMakingThis && typeof recipeMakingThis === 'object' && recipeMakingThis.ingredient1 && recipeMakingThis.ingredient2) {
        const recipeNode = {
            name: `${recipeMakingThis.ingredient1} + ${recipeMakingThis.ingredient2}`, // Nama node resep
            attributes: {
                type: 'Recipe', // Tandai sebagai node resep
                result: recipeMakingThis.result,
                ingredient1: recipeMakingThis.ingredient1,
                ingredient2: recipeMakingThis.ingredient2,
                depth: currentDepth + 1,
                originalName: `${recipeMakingThis.ingredient1} + ${recipeMakingThis.ingredient2}`
            },
            children: []
        };
        // Rekursif untuk kedua ingredient
        recipeNode.children.push(buildElementNodeRecursive(recipeMakingThis.ingredient1, pathRecipesMap, imageURLs, currentDepth + 2, maxDepth));
        recipeNode.children.push(buildElementNodeRecursive(recipeMakingThis.ingredient2, pathRecipesMap, imageURLs, currentDepth + 2, maxDepth));
        // Tambahkan node resep sebagai anak dari node elemen
        node.children.push(recipeNode);
    }
    return node;
};

// Builds the complete static tree data structure for a given path
const buildFullTreeData = (path, targetElement, imageURLs) => {
  // Jika elemen dasar atau path tidak valid, hanya tampilkan node elemen dasar
  if (isBaseElement(targetElement) || !Array.isArray(path) || path.length === 0) {
    return [buildInitialElementNode(targetElement, imageURLs)];
  }

  // Buat map resep untuk pencarian cepat
  const pathRecipesMap = {};
  path.forEach(recipe => {
    if (recipe && typeof recipe === 'object' && recipe.result) {
        pathRecipesMap[recipe.result] = recipe;
    }
  });

  // Mulai bangun pohon secara rekursif dari elemen target
  const rootNode = buildElementNodeRecursive(targetElement, pathRecipesMap, imageURLs);
  return rootNode ? [rootNode] : [buildInitialElementNode(targetElement, imageURLs)]; // Pastikan selalu mengembalikan array
};

// Custom node rendering function for react-d3-tree
const renderNodeWithImage = ({ nodeDatum, toggleNode }) => {
    const isRecipeNode = nodeDatum?.attributes?.type === 'Recipe';
    const originalElementName = nodeDatum?.attributes?.originalName || nodeDatum?.name || "N/A";
    const displayName = isRecipeNode ? `+` : originalElementName; // Tampilkan '+' untuk node resep
    const partialImageUrl = nodeDatum?.attributes?.imageUrl;
    const fullImageUrl = partialImageUrl ? `${API_BASE_URL}${partialImageUrl}` : '';
    const imageSize = isRecipeNode ? 20 : (isBaseElement(originalElementName) ? 40 : 35); // Ukuran gambar berbeda
    const textYOffset = isRecipeNode ? 0 : imageSize / 2 + 8; // Posisi teks

    // Kelas CSS berbeda untuk styling
    let nodeClass = "node-element";
    if (isRecipeNode) nodeClass = "node-recipe";
    else if (nodeDatum?.attributes?.type === 'Base Element') nodeClass = "node-base-element";
    else if (originalElementName.includes("Batas Kedalaman") || nodeDatum?.attributes?.type === 'Error') nodeClass = "node-error";

    return (
      <g onClick={toggleNode} className={`tree-node ${nodeClass}`}>
        {/* Lingkaran latar belakang */}
        <circle r={isRecipeNode ? 12 : imageSize / 2 + 3} className="node-circle-bg" />
        {/* Tampilkan gambar jika bukan node resep dan URL valid */}
        {!isRecipeNode && fullImageUrl && (
            <image
              x={-imageSize / 2} y={-imageSize / 2}
              width={imageSize} height={imageSize}
              href={fullImageUrl} className="element-image"
              onError={(e) => { e.target.style.display = 'none'; }} // Sembunyikan jika gambar gagal dimuat
            />
        )}
        {/* Teks nama elemen atau '+' */}
        <text strokeWidth={isRecipeNode ? "0.4" : "0.5"} x={0} y={textYOffset}
          textAnchor="middle" alignmentBaseline={isRecipeNode ? "middle" : "hanging"}
          className="node-text">
          {displayName}
        </text>
      </g>
    );
};

// Builds the *next* state of the tree during live update by applying one recipe step
const buildTreeForNextStep = (currentTreeDataArray, nextRecipeToProcess, allImageURLs) => {
    if (!Array.isArray(currentTreeDataArray) || currentTreeDataArray.length === 0 || !nextRecipeToProcess || typeof nextRecipeToProcess !== 'object') {
        const rootNode = buildInitialElementNode(nextRecipeToProcess.result, allImageURLs, 0);
        const recipeNode = {
            name: `${nextRecipeToProcess.ingredient1} + ${nextRecipeToProcess.ingredient2}`,
            attributes: { type: 'Recipe', result: nextRecipeToProcess.result, ingredient1: nextRecipeToProcess.ingredient1, ingredient2: nextRecipeToProcess.ingredient2, depth: 1, originalName: `${nextRecipeToProcess.ingredient1} + ${nextRecipeToProcess.ingredient2}` },
            children: [
                buildInitialElementNode(nextRecipeToProcess.ingredient1, allImageURLs, 2),
                buildInitialElementNode(nextRecipeToProcess.ingredient2, allImageURLs, 2)
            ]
        };
        rootNode.children = [recipeNode];
        return [rootNode];
    }

    const newTreeData = JSON.parse(JSON.stringify(currentTreeDataArray));
    const rootNode = newTreeData[0];
    let expanded = false;

    const deepCloneNode = (node) => {
        if (!node) return null;
        return JSON.parse(JSON.stringify(node));
    };

    const findNodeByName = (startNode, targetName) => {
        if (!startNode) return null;
        const queue = [startNode];
        while(queue.length > 0) {
            const currentNode = queue.shift();
            if (currentNode.name === targetName && currentNode.attributes?.type !== 'Recipe') {
                return currentNode;
            }
            if (currentNode.children) {
                queue.push(...currentNode.children);
            }
        }
        return null;
    };

    const findAndExpandNode = (node) => {
        if (!node || expanded || !node.attributes) return false;
        
        if (node.name === nextRecipeToProcess.result && node.attributes.type !== 'Recipe') {
            const alreadyHasThisRecipeAsChild = node.children?.some(child =>
                child.attributes?.type === 'Recipe' &&
                child.attributes?.ingredient1 === nextRecipeToProcess.ingredient1 &&
                child.attributes?.ingredient2 === nextRecipeToProcess.ingredient2);
            
            if (!alreadyHasThisRecipeAsChild) {
                const existingChildren = node.children || [];
                const recipeNode = {
                    name: `${nextRecipeToProcess.ingredient1} + ${nextRecipeToProcess.ingredient2}`,
                    attributes: { 
                        type: 'Recipe', 
                        result: nextRecipeToProcess.result, 
                        ingredient1: nextRecipeToProcess.ingredient1, 
                        ingredient2: nextRecipeToProcess.ingredient2, 
                        depth: (node.attributes.depth || 0) + 1, 
                        originalName: `${nextRecipeToProcess.ingredient1} + ${nextRecipeToProcess.ingredient2}`
                    },
                    children: []
                };
                
                const ingredient1FullNode = findNodeByName(rootNode, nextRecipeToProcess.ingredient1);
                const ingredient2FullNode = findNodeByName(rootNode, nextRecipeToProcess.ingredient2);
                
                recipeNode.children.push(
                    ingredient1FullNode && !isBaseElement(ingredient1FullNode.name) 
                    ? deepCloneNode(ingredient1FullNode) 
                    : buildInitialElementNode(nextRecipeToProcess.ingredient1, allImageURLs, (node.attributes.depth || 0) + 2)
                );
                
                recipeNode.children.push(
                    ingredient2FullNode && !isBaseElement(ingredient2FullNode.name) 
                    ? deepCloneNode(ingredient2FullNode) 
                    : buildInitialElementNode(nextRecipeToProcess.ingredient2, allImageURLs, (node.attributes.depth || 0) + 2)
                );
                
                node.children = [recipeNode, ...existingChildren];
                expanded = true;
                return true;
            }
        }
        
        if (node.children && !expanded) {
            for (let child of node.children) {
                const result = findAndExpandNode(child);
                if (result) {
                    return true;
                }
            }
        }
        return false;
    };

    findAndExpandNode(rootNode);
    return newTreeData;
};

// --- Komponen Utama SearchResults ---
function SearchResults({ results, isLoading, error }) {
  const [liveUpdateStates, setLiveUpdateStates] = useState({});
  const [treeDataForStaticView, setTreeDataForStaticView] = useState({});

  // Efek untuk membangun data pohon statis dan RESET LIVE UPDATE saat hasil pencarian berubah
  useEffect(() => {
    // ***** PERBAIKAN: Reset liveUpdateStates setiap kali 'results' baru diterima *****
    setLiveUpdateStates({});
    // *******************************************************************************

    try {
        const newStaticData = {};
        if (results && results.pathFound) {
            const imageURLs = results.imageURLs || {};
            if (results.mode === 'shortest' && Array.isArray(results.path)) {
                const pathKey = `path-block-shortest-0`;
                newStaticData[pathKey] = buildFullTreeData(results.path, results.searchTarget, imageURLs);
            } else if (results.mode === 'multiple' && Array.isArray(results.paths)) {
                results.paths.forEach((p, index) => {
                    if (Array.isArray(p)) {
                        const pathKey = `path-block-multiple-${index}`;
                        newStaticData[pathKey] = buildFullTreeData(p, results.searchTarget, imageURLs);
                    }
                });
            } else if (isBaseElement(results.searchTarget)) {
                const pathKey = `path-block-${results.searchTarget}-base`;
                newStaticData[pathKey] = buildFullTreeData([], results.searchTarget, imageURLs);
            }
        }
        setTreeDataForStaticView(newStaticData);
    } catch (e) {
        console.error("Error saat membangun treeDataForStaticView:", e);
        setTreeDataForStaticView({});
    }
  }, [results]); // Jalankan ulang saat `results` berubah

  const startLiveUpdate = useCallback((pathKey, originalPathData, targetElement, imageURLsInput) => {
    const imageURLs = imageURLsInput || {};
    if (isBaseElement(targetElement) || !Array.isArray(originalPathData) || originalPathData.length === 0) {
      setLiveUpdateStates(prev => ({
          ...prev,
          [pathKey]: {
              isActive: true,
              currentDisplayData: [buildInitialElementNode(targetElement, imageURLs)],
              fullPathRecipes: [],
              currentRecipeStep: 0,
              isBuilding: false,
              isComplete: true,
              pathIdentifier: `live-${pathKey}-${new Date().getTime()}`
          }
      }));
      return;
    }

    const rootNode = buildInitialElementNode(targetElement, imageURLs);
    setLiveUpdateStates(prev => ({
        ...prev,
        [pathKey]: {
            isActive: true,
            currentDisplayData: [rootNode],
            fullPathRecipes: [...originalPathData].reverse(),
            currentRecipeStep: 0,
            isBuilding: true,
            isComplete: false,
            pathIdentifier: `live-${pathKey}-${new Date().getTime()}`
        }
    }));
  }, []);

  const checkIfExpansionNeeded = useCallback((nodes, resultName, recipe) => {
      if (!Array.isArray(nodes) || !recipe) return false;
      const stack = [...nodes];
      while (stack.length > 0) {
          const node = stack.pop();
          if (!node) continue;
          if (node.name === resultName && node.attributes?.type !== 'Recipe') {
               const hasMatchingRecipeChild = node.children?.some(child =>
                   child.attributes?.type === 'Recipe' &&
                   child.attributes?.ingredient1 === recipe.ingredient1 &&
                   child.attributes?.ingredient2 === recipe.ingredient2
               );
               if (!hasMatchingRecipeChild && !isBaseElement(node.name)) {
                   return true;
               }
          }
          if (node.children) {
              for (let i = node.children.length - 1; i >= 0; i--) {
                  stack.push(node.children[i]);
              }
          }
      }
      return false;
  }, []);


  useEffect(() => {
    const activeLiveUpdates = Object.entries(liveUpdateStates).filter(([, state]) => state.isActive && state.isBuilding);
    if (activeLiveUpdates.length === 0) return;

    const timers = activeLiveUpdates.map(([pathKey, state]) => {
      if (state.currentRecipeStep < state.fullPathRecipes.length) {
        const timerId = setTimeout(() => {
          setLiveUpdateStates(prev => {
            const currentPathState = prev[pathKey];
            if (!currentPathState || !currentPathState.isActive || !currentPathState.isBuilding) return prev;

            const nextRecipe = currentPathState.fullPathRecipes[currentPathState.currentRecipeStep];
            const imageURLs = results?.imageURLs || {};
            const partiallyUpdatedTree = buildTreeForNextStep(currentPathState.currentDisplayData, nextRecipe, imageURLs);
            const needsMoreExpansionForThisRecipe = checkIfExpansionNeeded(partiallyUpdatedTree, nextRecipe.result, nextRecipe);
            const nextStepIndex = needsMoreExpansionForThisRecipe
                ? currentPathState.currentRecipeStep
                : currentPathState.currentRecipeStep + 1;
            const isNowComplete = nextStepIndex >= currentPathState.fullPathRecipes.length;

            return {
                ...prev,
                [pathKey]: {
                    ...currentPathState,
                    currentDisplayData: partiallyUpdatedTree,
                    currentRecipeStep: nextStepIndex,
                    isBuilding: !isNowComplete,
                    isComplete: isNowComplete,
                }
            };
          });
        }, LIVE_UPDATE_DELAY_MS);
        return timerId;
      } else {
        setLiveUpdateStates(prev => ({
            ...prev,
            [pathKey]: { ...prev[pathKey], isBuilding: false, isComplete: true }
        }));
        return null;
      }
    });

    return () => timers.forEach(id => { if (id) clearTimeout(id); });
  }, [liveUpdateStates, results?.imageURLs, checkIfExpansionNeeded]);

  const handleImageError = (e) => { e.target.style.display = 'none'; };

  const renderStep = (step, stepIndex, pathIndex = null) => {
    if (!step || typeof step !== 'object' || !step.result) return <li key={`invalid-step-${pathIndex}-${stepIndex}`} className="recipe-step-item invalid">Data langkah tidak valid</li>;
    const key = `${pathIndex !== null ? pathIndex + '-' : ''}${stepIndex}`;
    const imageURLs = results?.imageURLs || {};
    const imageUrl1 = imageURLs[step.ingredient1] ? `${API_BASE_URL}${imageURLs[step.ingredient1]}` : '';
    const imageUrl2 = imageURLs[step.ingredient2] ? `${API_BASE_URL}${imageURLs[step.ingredient2]}` : '';
    const imageUrlResult = imageURLs[step.result] ? `${API_BASE_URL}${imageURLs[step.result]}` : '';
    return (
      <li key={key} className="recipe-step-item">
        {imageUrl1 ? <img src={imageUrl1} alt={step.ingredient1 || ''} className="recipe-step-image" onError={handleImageError}/> : <span className="img-placeholder"></span>}
        {step.ingredient1 || '?'}
        <span className="recipe-step-separator">+</span>
        {imageUrl2 ? <img src={imageUrl2} alt={step.ingredient2 || ''} className="recipe-step-image" onError={handleImageError}/> : <span className="img-placeholder"></span>}
        {step.ingredient2 || '?'}
        <span className="recipe-step-separator">{' => '}</span>
        {imageUrlResult ? <img src={imageUrlResult} alt={step.result} className="recipe-step-image" onError={handleImageError}/> : <span className="img-placeholder"></span>}
        <strong className="recipe-step-result">{step.result}</strong>
      </li>);
  };

  const renderPath = (path, pathIndex = null) => {
    if (!Array.isArray(path)) return <p key={pathIndex !== null ? `invalid-path-data-${pathIndex}` : 'invalid-single-path-data'} className="invalid-path-message">Data jalur tidak valid.</p>;
    if (path.length === 0 && !isBaseElement(results?.searchTarget)) return <p key={pathIndex !== null ? `empty-path-${pathIndex}` : 'empty-single-path'} className="recipe-path-list-empty">(Tidak ada langkah resep)</p>;
    return <ol key={pathIndex !== null ? `path-${pathIndex}` : 'single-path-list'} className="recipe-path-list">{path.map((step, stepIndex) => renderStep(step, stepIndex, pathIndex))}</ol>;
  };

  if (isLoading) return <div className="loading-message">Memuat hasil pencarian...</div>;
  if (error) return <div className="error-message">Error: {(typeof error === 'object' && error.message) ? error.message : String(error)} <img src={notFoundImage} alt="Path Not Found" style={{ marginTop: '20px', maxWidth: '900px', width: '100%', height: 'auto', objectFit: 'contain' }} /></div>;
  if (!results || typeof results !== 'object') return <p className="initial-message">Silakan masukkan elemen yang ingin dicari.</p>;

  return (
    <div className="search-results-container">
      <h2 className="results-title">
        Hasil Pencarian untuk: <strong className="target-element">{results.searchTarget || 'N/A'}</strong>
        <span className="search-info">
          ({results.algorithm?.toUpperCase() || 'N/A'} / {results.mode || 'N/A'}
          {results.mode === 'multiple' && ` - Maks: ${results.maxRecipes || 'N/A'}`})
        </span>
      </h2>
      {results.pathFound && (
         <div className="search-stats">
             <span>Node Diperiksa: <strong>{results.nodesVisited !== undefined && results.nodesVisited !== -1 ? results.nodesVisited.toLocaleString() : 'N/A'}</strong></span>
             <span>Durasi: <strong>{results.durationMillis ?? 'N/A'} ms</strong></span>
         </div> )}

       {results.pathFound ? (
           <>
            {isBaseElement(results.searchTarget) && (
                ((results.mode === 'shortest' && (!Array.isArray(results.path) || results.path.length === 0)) ||
                 (results.mode === 'multiple' && (!Array.isArray(results.paths) || results.paths.length === 0 || (Array.isArray(results.paths) && results.paths.every(p => !Array.isArray(p) || p.length === 0)))))
            ) && ( <div className="base-element-message">"{results.searchTarget}" adalah elemen dasar.</div> )}

           {results.mode === 'shortest' && Array.isArray(results.path) && results.path.length > 0 && (() => {
                const pathKey = `path-block-shortest-0`;
                const currentLiveState = liveUpdateStates[pathKey]; // Akan undefined setelah reset jika pathKey baru
                const staticTreeData = treeDataForStaticView[pathKey];
                const treeToDisplay = currentLiveState?.isActive && currentLiveState.currentDisplayData ? currentLiveState.currentDisplayData : staticTreeData;
                return (
                   <div key={pathKey} className="path-block">
                       <div className="path-details-column">
                           <div className="path-text-section">
                               <h4 className="path-title">Jalur Terpendek (Langkah: {results.path.length})</h4>
                               {renderPath(results.path, `shortest-0`)}
                           </div>
                           {!isBaseElement(results.searchTarget) && (
                               <button
                                   onClick={() => startLiveUpdate(pathKey, results.path, results.searchTarget, results.imageURLs)}
                                   disabled={currentLiveState?.isBuilding}
                                   className="live-update-button">
                                   {currentLiveState?.isBuilding ? 'Memproses...' : (currentLiveState?.isComplete ? 'Putar Ulang Live' : 'Mulai Live Update')}
                               </button>
                           )}
                       </div>
                       {treeToDisplay && treeToDisplay.length > 0 && treeToDisplay[0] && (
                           <div className="path-visualization-column">
                                <h5 className="visualization-title-small">Visualisasi {currentLiveState?.isActive ? "(Live)" : ""}</h5>
                                <div id={`treeWrapper-shortest-0`} className="tree-wrapper">
                                    <Tree data={treeToDisplay} orientation="vertical" translate={{ x: 300, y: 50 }}
                                        renderCustomNodeElement={renderNodeWithImage} zoomable={true} draggable={true}
                                        nodeSize={{ x: 140, y: 120 }} separation={{ siblings: 1.2, nonSiblings: 1.5 }}
                                        pathFunc="straight" depthFactor={150}
                                        key={currentLiveState?.pathIdentifier || `static-tree-${pathKey}-${results.searchTarget}`} // Tambahkan searchTarget ke key statis
                                    />
                                 </div>
                           </div>
                       )}
                   </div>);
            })()}

           {results.mode === 'multiple' && Array.isArray(results.paths) && results.paths.length > 0 &&
            results.paths.map((path, index) => {
                if (!Array.isArray(path) || (path.length === 0 && !isBaseElement(results.searchTarget))) return null;
                if (path.length === 0 && isBaseElement(results.searchTarget)) return null;

                const pathKey = `path-block-multiple-${index}`;
                const currentLiveState = liveUpdateStates[pathKey]; // Akan undefined setelah reset
                const staticTreeData = treeDataForStaticView[pathKey];
                const treeToDisplay = currentLiveState?.isActive && currentLiveState.currentDisplayData ? currentLiveState.currentDisplayData : staticTreeData;
                return (
                   <div key={pathKey} className="path-block">
                        <div className="path-details-column">
                           <div className="path-text-section">
                               <h4 className="path-title">Jalur {index + 1} (Langkah: {path.length})</h4>
                               {renderPath(path, `multiple-${index}`)}
                           </div>
                           {!isBaseElement(results.searchTarget) && path.length > 0 && (
                               <button
                                   onClick={() => startLiveUpdate(pathKey, path, results.searchTarget, results.imageURLs)}
                                   disabled={currentLiveState?.isBuilding}
                                   className="live-update-button">
                                   {currentLiveState?.isBuilding ? 'Memproses...' : (currentLiveState?.isComplete ? 'Putar Ulang Live' : 'Mulai Live Update')}
                               </button>
                           )}
                        </div>
                         {treeToDisplay && treeToDisplay.length > 0 && treeToDisplay[0] && (
                           <div className="path-visualization-column">
                                <h5 className="visualization-title-small">Visualisasi Jalur {index + 1} {currentLiveState?.isActive ? "(Live)" : ""}</h5>
                                <div id={`treeWrapper-multiple-${index}`} className="tree-wrapper">
                                     <Tree data={treeToDisplay} orientation="vertical" translate={{ x: 300, y: 50 }}
                                         renderCustomNodeElement={renderNodeWithImage} zoomable={true} draggable={true}
                                         nodeSize={{ x: 140, y: 120 }} separation={{ siblings: 1.2, nonSiblings: 1.5 }}
                                         pathFunc="straight" depthFactor={150}
                                         key={currentLiveState?.pathIdentifier || `static-tree-${pathKey}-${results.searchTarget}`} // Tambahkan searchTarget ke key statis
                                     />
                                 </div>
                           </div>
                         )}
                   </div>);
            })}
           </>
       ) : (
            results && results.searchTarget &&
            <div className="path-not-found-message">
                 Jalur tidak ditemukan untuk elemen "{results.searchTarget}".
                 {results.error ? ` (Error: ${results.error})` : ''}
            </div>
       )}
    </div>
  );
}

export default SearchResults;
