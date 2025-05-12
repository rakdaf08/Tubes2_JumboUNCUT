import React, { useState, useEffect, useCallback } from 'react';
import Tree from 'react-d3-tree';
import './SearchResults.css';
import notFoundImage from '../assets/notfound_dontol.jpg';

// API_BASE_URL HANYA untuk panggilan API ke backend, BUKAN untuk path gambar lokal
// const API_BASE_URL = "http://localhost:8080"; // Ini untuk /api/search, dll.
// Untuk gambar, kita akan menggunakan path relatif langsung seperti "/image/Fire.png"

const LIVE_UPDATE_DELAY_MS = 800;

const isBaseElement = (name) => {
    const baseElements = ["Air", "Earth", "Fire", "Water"];
    return baseElements.includes(name);
};

const buildInitialElementNode = (elementName, imageURLs, depth = 0) => {
    // DEBUG: Lihat imageURLs dan elementName yang diterima
    // console.log(`buildInitialElementNode - Element: ${elementName}, ImageURLs Diterima:`, imageURLs);

    // Ambil path gambar dari map imageURLs
    const imageUrlPath = imageURLs && typeof imageURLs === 'object' ? (imageURLs[elementName] || '') : '';

    // DEBUG: Lihat path gambar yang diambil
    // console.log(`buildInitialElementNode - Element: ${elementName}, Path Gambar Digunakan: ${imageUrlPath}`);

    return {
      name: elementName || "Unknown",
      attributes: {
        type: isBaseElement(elementName) ? 'Base Element' : 'Element',
        imageUrl: imageUrlPath, // Ini adalah path relatif seperti "/image/Fire.png"
        depth: depth,
        originalName: elementName || "Unknown",
      },
      children: [],
    };
};

// buildElementNodeRecursive dan buildFullTreeData tetap sama, mereka akan menggunakan buildInitialElementNode
// ... (buildElementNodeRecursive dan buildFullTreeData seperti yang Anda berikan sebelumnya) ...
const buildElementNodeRecursive = (elementName, pathRecipesMap, imageURLs, currentDepth = 0, maxDepth = 20) => {
    if (currentDepth > maxDepth) {
        return {
            name: `${elementName || "Unknown"} (Batas Kedalaman)`,
            attributes: { type: 'Error', originalName: elementName || "Unknown", depth: currentDepth, info: 'Max depth reached' },
            children: []
        };
    }
    const node = buildInitialElementNode(elementName, imageURLs, currentDepth);
    const recipeMakingThis = pathRecipesMap && typeof pathRecipesMap === 'object' ? pathRecipesMap[elementName] : null;
    if (recipeMakingThis && typeof recipeMakingThis === 'object' && recipeMakingThis.ingredient1 && recipeMakingThis.ingredient2) {
        const recipeNode = {
            name: `${recipeMakingThis.ingredient1} + ${recipeMakingThis.ingredient2}`,
            attributes: {
                type: 'Recipe',
                result: recipeMakingThis.result,
                ingredient1: recipeMakingThis.ingredient1,
                ingredient2: recipeMakingThis.ingredient2,
                depth: currentDepth + 1,
                originalName: `${recipeMakingThis.ingredient1} + ${recipeMakingThis.ingredient2}`
            },
            children: []
        };
        recipeNode.children.push(buildElementNodeRecursive(recipeMakingThis.ingredient1, pathRecipesMap, imageURLs, currentDepth + 2, maxDepth));
        recipeNode.children.push(buildElementNodeRecursive(recipeMakingThis.ingredient2, pathRecipesMap, imageURLs, currentDepth + 2, maxDepth));
        node.children.push(recipeNode);
    }
    return node;
};

const buildFullTreeData = (path, targetElement, imageURLs) => {
  if (isBaseElement(targetElement) || !Array.isArray(path) || path.length === 0) {
    return [buildInitialElementNode(targetElement, imageURLs)];
  }
  const pathRecipesMap = {};
  path.forEach(recipe => {
    if (recipe && typeof recipe === 'object' && recipe.result) {
        pathRecipesMap[recipe.result] = recipe;
    }
  });
  const rootNode = buildElementNodeRecursive(targetElement, pathRecipesMap, imageURLs);
  return rootNode ? [rootNode] : [buildInitialElementNode(targetElement, imageURLs)];
};


const renderNodeWithImage = ({ nodeDatum, toggleNode }) => {
    const isRecipeNode = nodeDatum?.attributes?.type === 'Recipe';
    const originalElementName = nodeDatum?.attributes?.originalName || nodeDatum?.name || "N/A";
    const displayName = isRecipeNode ? `+` : originalElementName;
    
    // partialImageUrl sekarang adalah path relatif seperti "/image/Fire.png"
    const partialImageUrl = nodeDatum?.attributes?.imageUrl;
    
    // TIDAK PERLU API_BASE_URL untuk gambar lokal yang disajikan frontend
    // const fullImageUrl = partialImageUrl ? `${API_BASE_URL}${partialImageUrl}` : ''; // <--- HAPUS API_BASE_URL
    const fullImageUrl = partialImageUrl; // <--- GUNAKAN LANGSUNG

    // DEBUG: Log path gambar yang akan dirender
    if (!isRecipeNode && partialImageUrl) {
        // console.log(`renderNodeWithImage - Merender gambar untuk: ${originalElementName}, Path: ${fullImageUrl}`);
    }


    const imageSize = isRecipeNode ? 20 : (isBaseElement(originalElementName) ? 40 : 35);
    const textYOffset = isRecipeNode ? 0 : imageSize / 2 + 8;

    let nodeClass = "node-element";
    if (isRecipeNode) nodeClass = "node-recipe";
    else if (nodeDatum?.attributes?.type === 'Base Element') nodeClass = "node-base-element";
    else if (originalElementName.includes("Batas Kedalaman") || nodeDatum?.attributes?.type === 'Error') nodeClass = "node-error";

    return (
      <g onClick={toggleNode} className={`tree-node ${nodeClass}`}>
        <circle r={isRecipeNode ? 12 : imageSize / 2 + 3} className="node-circle-bg" />
        {!isRecipeNode && fullImageUrl && (
            <image
              x={-imageSize / 2} y={-imageSize / 2}
              width={imageSize} height={imageSize}
              href={fullImageUrl} // href sekarang menggunakan path relatif langsung
              className="element-image"
              onError={(e) => {
                console.error(`renderNodeWithImage - GAGAL MEMUAT GAMBAR: ${fullImageUrl} untuk elemen ${originalElementName}`);
                e.target.style.display = 'none';
              }}
            />
        )}
        <text strokeWidth={isRecipeNode ? "0.4" : "0.5"} x={0} y={textYOffset}
          textAnchor="middle" alignmentBaseline={isRecipeNode ? "middle" : "hanging"}
          className="node-text">
          {displayName}
        </text>
      </g>
    );
};


// buildTreeForNextStep perlu menggunakan imageURLs dengan benar
const buildTreeForNextStep = (currentTreeDataArray, nextRecipeToProcess, allImageURLs) => {
    // ... (bagian awal fungsi ini mirip) ...
    if (!Array.isArray(currentTreeDataArray) || currentTreeDataArray.length === 0 || !nextRecipeToProcess || typeof nextRecipeToProcess !== 'object') {
        const rootNode = buildInitialElementNode(nextRecipeToProcess.result, allImageURLs, 0); // Gunakan allImageURLs
        const recipeNode = {
            name: `${nextRecipeToProcess.ingredient1} + ${nextRecipeToProcess.ingredient2}`,
            attributes: { type: 'Recipe', result: nextRecipeToProcess.result, ingredient1: nextRecipeToProcess.ingredient1, ingredient2: nextRecipeToProcess.ingredient2, depth: 1, originalName: `${nextRecipeToProcess.ingredient1} + ${nextRecipeToProcess.ingredient2}` },
            children: [
                buildInitialElementNode(nextRecipeToProcess.ingredient1, allImageURLs, 2), // Gunakan allImageURLs
                buildInitialElementNode(nextRecipeToProcess.ingredient2, allImageURLs, 2)  // Gunakan allImageURLs
            ]
        };
        rootNode.children = [recipeNode];
        return [rootNode];
    }
    // ... (sisa fungsi findAndExpandNode dan lainnya, pastikan buildInitialElementNode dipanggil dengan allImageURLs) ...
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


function SearchResults({ results, isLoading, error }) {
  const [liveUpdateStates, setLiveUpdateStates] = useState({});
  const [treeDataForStaticView, setTreeDataForStaticView] = useState({});

  useEffect(() => {
    setLiveUpdateStates({});
    console.log("SearchResults DEBUG: Menerima `results` baru:", JSON.stringify(results, null, 2)); // Log `results` yang diterima

    try {
        const newStaticData = {};
        if (results && results.pathFound) {
            const imageURLs = results.imageURLs || {};
            console.log("SearchResults DEBUG: `imageURLs` yang akan digunakan untuk pohon statis:", imageURLs); // Log imageURLs
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
        } else if (results) {
            console.log("SearchResults DEBUG: results.pathFound adalah false atau results tidak ada. Target:", results.searchTarget);
        }
        setTreeDataForStaticView(newStaticData);
    } catch (e) {
        console.error("Error saat membangun treeDataForStaticView:", e);
        setTreeDataForStaticView({});
    }
  }, [results]);

  // ... (startLiveUpdate dan useEffect untuk live update tetap sama, pastikan mereka menggunakan imageURLs dengan benar) ...
  const startLiveUpdate = useCallback((pathKey, originalPathData, targetElement, imageURLsInput) => {
    const imageURLs = imageURLsInput || {}; // Gunakan imageURLs yang diterima
    console.log(`startLiveUpdate DEBUG: PathKey: ${pathKey}, Target: ${targetElement}, ImageURLs:`, imageURLs);
    // ... sisa fungsi ...
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
            const imageURLs = results?.imageURLs || {}; // Ambil dari results prop
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


  const handleImageError = (e, elementName, imageUrlAttempted) => {
    console.error(`SearchResults DEBUG: Gagal memuat gambar untuk elemen '${elementName}' dari path '${imageUrlAttempted}'`);
    e.target.style.display = 'none'; // Sembunyikan gambar yang gagal
    // Anda bisa menggantinya dengan placeholder jika mau: e.target.src = '/path/to/placeholder.png';
  };

  const renderStep = (step, stepIndex, pathIndex = null) => {
    if (!step || typeof step !== 'object' || !step.result) return <li key={`invalid-step-${pathIndex}-${stepIndex}`} className="recipe-step-item invalid">Data langkah tidak valid</li>;
    const key = `${pathIndex !== null ? pathIndex + '-' : ''}${stepIndex}`;
    const imageURLs = results?.imageURLs || {}; // Ambil dari results prop

    // DEBUG: Log path gambar untuk setiap langkah resep
    const imgPath1 = imageURLs[step.ingredient1] || '';
    const imgPath2 = imageURLs[step.ingredient2] || '';
    const imgPathResult = imageURLs[step.result] || '';

    // console.log(`renderStep DEBUG: Step ${stepIndex}, Ing1: ${step.ingredient1}, Path1: ${imgPath1}`);
    // console.log(`renderStep DEBUG: Step ${stepIndex}, Ing2: ${step.ingredient2}, Path2: ${imgPath2}`);
    // console.log(`renderStep DEBUG: Step ${stepIndex}, Result: ${step.result}, PathRes: ${imgPathResult}`);

    // TIDAK PERLU API_BASE_URL untuk path relatif
    const imageUrl1 = imgPath1;
    const imageUrl2 = imgPath2;
    const imageUrlResult = imgPathResult;

    return (
      <li key={key} className="recipe-step-item">
        {imageUrl1 ? <img src={imageUrl1} alt={step.ingredient1 || ''} className="recipe-step-image" onError={(e) => handleImageError(e, step.ingredient1, imageUrl1)}/> : <span className="img-placeholder">?</span>}
        {step.ingredient1 || '?'}
        <span className="recipe-step-separator">+</span>
        {imageUrl2 ? <img src={imageUrl2} alt={step.ingredient2 || ''} className="recipe-step-image" onError={(e) => handleImageError(e, step.ingredient2, imageUrl2)}/> : <span className="img-placeholder">?</span>}
        {step.ingredient2 || '?'}
        <span className="recipe-step-separator">{' => '}</span>
        {imageUrlResult ? <img src={imageUrlResult} alt={step.result} className="recipe-step-image" onError={(e) => handleImageError(e, step.result, imageUrlResult)}/> : <span className="img-placeholder">?</span>}
        <strong className="recipe-step-result">{step.result}</strong>
      </li>);
  };

  // ... (renderPath, dan bagian return utama tetap sama) ...
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
                const currentLiveState = liveUpdateStates[pathKey]; 
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
                                        key={currentLiveState?.pathIdentifier || `static-tree-${pathKey}-${results.searchTarget}`} 
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
                const currentLiveState = liveUpdateStates[pathKey]; 
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
                                         key={currentLiveState?.pathIdentifier || `static-tree-${pathKey}-${results.searchTarget}`} 
                                     />
                                 </div>
                           </div>
                         )}
                   </div>);
            })}
           </>
       ) : (
            results && results.searchTarget && // Hanya tampilkan pesan ini jika results ada dan punya searchTarget
            <div className="path-not-found-message">
                 Jalur tidak ditemukan untuk elemen "{results.searchTarget}".
                 {results.error ? ` (Error: ${results.error})` : ''}
            </div>
       )}
    </div>
  );
}

export default SearchResults;