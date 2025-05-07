// src/frontend/src/components/SearchResults.jsx
import React from 'react';
import Tree from 'react-d3-tree'; // Import komponen Tree

// Alamat dasar API backend Anda (pastikan ini ada di bagian paling atas file)
const API_BASE_URL = "http://localhost:8080"; // SESUAIKAN jika backend berjalan di port lain

// --- Fungsi Helper untuk Mengecek Elemen Dasar ---
const isBaseElement = (name) => {
    const baseElements = ["Air", "Earth", "Fire", "Water"];
    return baseElements.includes(name);
};


// --- Fungsi Rekursif untuk Membangun Simpul Elemen (dengan Logging Debug Depth) ---
const buildElementNode = (elementName, pathRecipesMap, imageURLs, depth = 0) => {
     // Logging untuk debugging kedalaman rekursi
     console.log(`[buildElementNode] Processing: ${elementName}, Depth: ${depth}`);

     // Kondisi berhenti darurat sementara untuk debugging jika rekursi terlalu dalam
     // Ini MENCEGAH crash browser karena stack overflow, bukan solusi permanen
     if (depth > 500) { // Angka 500 bisa disesuaikan
         console.error(`[buildElementNode] Max depth (${depth}) reached for ${elementName}. Stopping recursion.`);
         return { // Kembalikan simpul sederhana untuk menghentikan rekursi dan visualisasi parsial
              name: `${elementName} (MAX_DEPTH)`,
              attributes: { type: 'Error', originalName: elementName, depth: depth },
              children: [] // Tidak ada anak lagi dari sini
         };
     }

     const node = {
         name: elementName, // Nama elemen yang akan ditampilkan di visualisasi
         attributes: { // Atribut tambahan yang bisa disimpan di simpul
             type: isBaseElement(elementName) ? 'Base Element' : 'Element',
             imageUrl: imageURLs?.[elementName] || '', // URL gambar elemen dari data backend
             depth: depth // Simpan kedalaman di atribut
         },
         children: [] // Disiapkan untuk simpul resep pembuatnya
     };

     // Cari di 'pathRecipesMap' apakah ada resep di jalur ini yang menghasilkan 'elementName'
     const recipeMakingThis = pathRecipesMap[elementName];

     if (recipeMakingThis) {
         // Buat simpul resep untuk resep pembuatnya
         // Panggil rekursif untuk simpul resep, tambahkan depth
         const recipeNode = buildRecipeNode(recipeMakingThis, pathRecipesMap, imageURLs, depth + 1);
         // Tambahkan simpul resep ini sebagai anak dari simpul elemen hasilnya
         node.children.push(recipeNode);
     }

     return node;
};


// --- Fungsi Rekursif untuk Membangun Simpul Resep (dengan Logging Debug Depth) ---
const buildRecipeNode = (recipe, pathRecipesMap, imageURLs, depth = 0) => {
  // Logging untuk debugging kedalaman rekursi
  console.log(`[buildRecipeNode] Processing Recipe: ${recipe.ingredient1} + ${recipe.ingredient2} => ${recipe.result}, Depth: ${depth}`);

   // Kondisi berhenti darurat sementara untuk debugging jika rekursi terlalu dalam
   if (depth > 500) { // Angka 500 bisa disesuaikan
        console.error(`[buildRecipeNode] Max depth (${depth}) reached for recipe ${recipe.ingredient1} + ${recipe.ingredient2} => ${recipe.result}. Stopping recursion.`);
         return { // Kembalikan simpul sederhana untuk menghentikan rekursi dan visualisasi parsial
            name: `(${recipe.ingredient1} + ${recipe.ingredient2}) (MAX_DEPTH)`,
            attributes: { type: 'ErrorRecipe', result: recipe.result, depth: depth },
            children: [] // Tidak ada anak lagi dari sini
        };
   }


  const node = {
      // Nama simpul resep menampilkan kombinasinya
      name: `${recipe.ingredient1} + ${recipe.ingredient2}`,
      attributes: { // Atribut tambahan untuk simpul resep
          type: 'Recipe',
          result: recipe.result,
          ingredient1: recipe.ingredient1,
          ingredient2: recipe.ingredient2,
          depth: depth // Simpan kedalaman
      },
      children: [] // Disiapkan untuk simpul elemen bahan-bahannya
  };

  // Buat simpul anak untuk bahan pertama (ingredient1)
  // Panggil rekursif untuk simpul bahan, tambahkan depth
  const ingredient1Node = buildElementNode(recipe.ingredient1, pathRecipesMap, imageURLs, depth + 1);
  // Buat simpul anak untuk bahan kedua (ingredient2)
  const ingredient2Node = buildElementNode(recipe.ingredient2, pathRecipesMap, imageURLs, depth + 1);

  // Tambahkan simpul bahan sebagai anak dari simpul resep
  node.children.push(ingredient1Node);
  node.children.push(ingredient2Node);

  return node;
};


// --- Fungsi Utama Transformasi Data Pohon ---
const buildTreeData = (path, targetElement, imageURLs) => {
  // Kasus khusus: Jika target adalah elemen dasar atau path kosong (meskipun pathFound true)
  // Kita hanya perlu membuat satu simpul untuk elemen target itu sendiri
  if (!path || path.length === 0 || isBaseElement(targetElement)) {
     const isBase = isBaseElement(targetElement);
     const rootNode = {
         name: targetElement,
         attributes: {
             type: isBase ? 'Base Element' : 'Target Element',
             imageUrl: imageURLs?.[targetElement] || '',
             depth: 0 // Depth 0 untuk elemen dasar/target tanpa path
         },
         children: [] // Children kosong agar komponen Tree tetap bisa render simpul ini
     };
     return [rootNode]; // react-d3-tree butuh array yang berisi simpul akar
  }

  // Jika ada path (bukan elemen dasar dan path tidak kosong):
  // Buat map untuk mencari resep di dalam jalur ini dengan cepat berdasarkan hasilnya
  const pathRecipesMap = {};
  path.forEach(recipe => {
      pathRecipesMap[recipe.result] = recipe;
  });

  // Mulai pembangunan pohon dari elemen target (akar) dengan depth awal 0
  const rootNode = buildElementNode(targetElement, pathRecipesMap, imageURLs, 0); // <-- Mulai dengan depth 0

  return [rootNode];
};


// --- Fungsi untuk Menggambar Simpul Kustom (dengan Teks dan Gambar) ---
const renderNodeWithImage = ({ nodeDatum, toggleNode }) => {
    const isRecipeNode = nodeDatum.attributes?.type === 'Recipe';
    // Gunakan encodeURIComponent untuk parameter URL jika nama elemen mengandung karakter khusus
    const elementNameToFetch = nodeDatum.name || '';
    // HANYA buat URL proxy untuk simpul elemen (bukan simpul resep)
    const imageUrlFromBackendProxy = elementNameToFetch && nodeDatum.attributes?.type !== 'Recipe'
    ? `${API_BASE_URL}/api/image?elementName=${encodeURIComponent(elementNameToFetch)}`
    : ''; // URL kosong jika nama elemen tidak valid atau simpul resep

    return (
      // Group ini bisa diklik untuk toggle collapse/expand
      <g onClick={toggleNode}>

        {/* Tampilan untuk Simpul Elemen (bukan resep) */}
        {!isRecipeNode && (
            <>
                 {/* Lingkaran sebagai latar belakang simpul elemen */}
                <circle r={20} fill="#4682B4" stroke="#000" strokeWidth="1.5" />

                 {/* Gambar Elemen (jika ada URL gambar proxy) */}
                 {/* Hanya tampilkan jika imageUrlFromBackendProxy tidak kosong */}
                {imageUrlFromBackendProxy && (
                   <image
                      x={-15} // Sesuaikan posisi X agar gambar di tengah/kiri lingkaran
                      y={-35} // Sesuaikan posisi Y agar gambar di atas lingkaran
                      width={30} // Ukuran lebar gambar
                      height={30} // Ukuran tinggi gambar
                      href={imageUrlFromBackendProxy} // Gunakan URL dari backend proxy
                      onError={(e) => {
                           console.warn(`Failed to load image for ${nodeDatum.name} from proxy: ${e.target.href}`);
                           e.target.style.display = 'none'; // Sembunyikan elemen <image> jika gagal load
                       }}
                       // Tambahkan title untuk tooltip saat hover
                       title={`Elemen: ${nodeDatum.name}`}
                   />
                )}

                {/* Teks Nama Elemen */}
                <text
                  strokeWidth="0.5"
                  x={0}
                  y={10}
                  textAnchor="middle"
                  alignmentBaseline="middle"
                  style={{
                       // Ukuran font
                       fontSize: '14px',
                       // Warna teks (putih agar terlihat di lingkaran biru)
                       fill: '#fff',
                       fontWeight: 'bold',
                       // Agar klik pada teks tetap mengaktifkan toggleNode pada group <g>
                       pointerEvents: 'none',
                  }}
                >
                  {nodeDatum.name}
                </text>

                 {/* Opsional: Tampilkan Depth untuk debugging */}
                 {/*
                 <text x="0" y="30" textAnchor="middle" fontSize="10" fill="black">
                     Depth: {nodeDatum.attributes?.depth ?? 'N/A'}
                 </text>
                 */}
            </>
        )}

        {/* Tampilan untuk Simpul Resep */}
        {isRecipeNode && (
             <>
                 {/* Kotak sebagai latar belakang simpul resep */}
                 <rect
                    x={-60} // Sesuaikan posisi X agar teks kombinasi berada di tengah
                    y={-15} // Sesuaikan posisi Y
                    width={120} // Lebar kotak (sesuaikan dengan panjang teks resep)
                    height={30} // Tinggi kotak
                    rx={5} // Radius sudut kotak
                    ry={5}
                    fill="#e9ecef" // Warna latar belakang resep (abu-abu muda)
                    stroke="#adb5bd" // Warna border (abu-abu sedang)
                    strokeWidth="1"
                 />
                  {/* Teks Nama Resep / Kombinasi */}
                 <text
                   strokeWidth="0.5"
                   x={0}
                   y={5}
                   textAnchor="middle"
                   alignmentBaseline="middle"
                   style={{
                       // Ukuran font lebih kecil dari nama elemen
                       fontSize: '12px',
                       // Warna teks gelap
                       fill: '#495057',
                       // Agar klik pada teks tetap mengaktifkan toggleNode pada group <g>
                       pointerEvents: 'none',
                   }}
                 >
                   {nodeDatum.name} {/* Ini akan menampilkan "Bahan1 + Bahan2" */}
                 </text>

                 {/* Opsional: Tampilkan Depth untuk debugging */}
                 {/*
                 <text x="0" y="20" textAnchor="middle" fontSize="10" fill="black">
                     Depth: {nodeDatum.attributes?.depth ?? 'N/A'}
                 </text>
                 */}
             </>
        )}

        {/* Anda bisa tambahkan tooltip atau interaksi lain di sini jika diinginkan */}

      </g>
    );
};


function SearchResults({ results, isLoading, error }) {
  // Tampilkan loading indicator
  if (isLoading) {
    return <div style={{ textAlign: 'center', padding: '20px', fontSize: '18px', color: '#333' }}>Loading...</div>;
  }

  // Tampilkan pesan error jika ada
  if (error) {
    const errorMessage = (typeof error === 'object' && error !== null && error.message) ? error.message : String(error);
    return <div style={{ color: '#721c24', border: '1px solid #f5c6cb', background: '#f8d7da', padding: '15px', borderRadius: '5px', marginTop: '20px' }}>Error: {errorMessage}</div>;
  }

  // Tampilkan pesan awal jika belum ada hasil (atau jika results bukan objek)
  // Ini akan tampil saat pertama kali halaman dibuka atau setelah search form direset tanpa hasil
  if (!results || typeof results !== 'object') {
    return <p style={{ textAlign: 'center', marginTop: '20px', color: '#6c757d' }}>Masukkan elemen target dan klik cari.</p>;
  }

   // --- FUNGSI HELPER UNTUK TAMPILAN TEKS ---
   // Fungsi ini digunakan di tampilan teks dan daftar URL Gambar
   const handleImageError = (e) => {
       console.warn(`Image failed to load: ${e.target.src}`);
       e.target.style.display = 'none'; // Sembunyikan tag img
        // Opsional: Tambahkan teks atau placeholder
       const altText = e.target.alt || 'image';
       const errorSpan = document.createElement('span');
       errorSpan.textContent = `(${altText} error)`;
       errorSpan.style.fontSize = '0.8em';
       errorSpan.style.color = '#dc3545';
       // Sisipkan teks error setelah gambar yang gagal, pastikan parentNode ada
       if (e.target.parentNode) {
           e.target.parentNode.insertBefore(errorSpan, e.target.nextSibling);
       }
   };

  // Helper function untuk merender satu langkah resep dalam teks dengan gambar kecil
  // Memperbaiki key dan sintaks URL gambar
  const renderStep = (step, stepIndex, pathIndex = null) => {
    if (!step || typeof step !== 'object') {
        // Gunakan key yang lebih pasti unik untuk data yang tidak valid juga
        // Menggabungkan 'invalid-step', pathIndex, dan stepIndex
        return <li key={`invalid-step-${pathIndex}-${stepIndex}`} style={{color: 'red'}}>Data langkah tidak valid</li>;
    }
    // --- PERBAIKAN KEY ---
    // Key untuk elemen <li> dalam daftar harus unik di dalam list <ol> yang sama.
    // Menggunakan gabungan pathIndex dan stepIndex memastikan keunikan
    // bahkan ketika merender multiple list <ol> di mode multiple.
    // Pastikan pathIndex !== null saat mode multiple.
    // SINTAKS template literal untuk key HARUS: `${pathIndex}-${stepIndex}`
    const key = `${pathIndex}-${stepIndex}`; // <--- PASTIKAN BARIS INI BENAR
    // --- AKHIR PERBAIKAN KEY ---


    // Ambil path URL gambar dari results.imageURLs
    const imageUrlPath1 = results.imageURLs?.[step.ingredient1];
    const imageUrlPath2 = results.imageURLs?.[step.ingredient2];
    const imageUrlPathResult = results.imageURLs?.[step.result];

    // --- PERBAIKAN URL GAMBAR ---
    // Gunakan template literal (backticks `` ` ``) murni untuk interpolasi URL.
    // Format yang benar: `teks biasa ${variabel} teks lain ${variabel lain}`
    // URL ini akan mengarah ke endpoint backend proxy gambar
    // PASTIKAN SINTAKS TEMPLATE LITERAL DI BAWAH BENAR. Tidak ada <span>, \, dll.
    const imageUrl1 = imageUrlPath1 ? `${API_BASE_URL}${imageUrlPath1}` : ''; // <--- PASTIKAN BARIS INI BENAR
    const imageUrl2 = imageUrlPath2 ? `${API_BASE_URL}${imageUrlPath2}` : ''; // <--- PASTIKAN BARIS INI BENAR
    const imageUrlResult = imageUrlPathResult ? `${API_BASE_URL}${imageUrlPathResult}` : ''; // <--- PASTIKAN BARIS INI BENAR
    // --- AKHIR PERBAIKAN URL GAMBAR ---


    const imgStyle = {
        height: '20px',
        width: '20px',
        verticalAlign: 'middle',
        margin: '0 3px',
        border: '1px solid #eee',
        objectFit: 'contain'
    };

    return (
      // Gunakan key yang sudah diperbaiki
      <li key={key} style={{ marginBottom: '8px', lineHeight: '1.5', color: '#212529' }}>
        {/* Gunakan URL LENGKAP pada atribut src, pastikan sintaks template literal benar */}
        {/* Tag img hanya dirender jika imageUrl1 tidak kosong */}
        {/* onError={handleImageError} akan dipanggil jika gambar gagal dimuat */}
        {imageUrl1 ? <img src={imageUrl1} alt={step.ingredient1 || 'ingredient1'} style={imgStyle} onError={handleImageError}/> : null}
        {step.ingredient1 || '?'}
        <span style={{ margin: '0 5px' }}>+</span>
         {/* Tag img hanya dirender jika imageUrl2 tidak kosong */}
        {imageUrl2 ? <img src={imageUrl2} alt={step.ingredient2 || 'ingredient2'} style={imgStyle} onError={handleImageError}/> : null}
        {step.ingredient2 || '?'}
        <span style={{ margin: '0 5px' }}>{' => '}</span>
         {/* Tag img hanya dirender jika imageUrlResult tidak kosong */}
        {imageUrlResult ? <img src={imageUrlResult} alt={step.result || 'result'} style={imgStyle} onError={handleImageError}/> : null}
        <strong style={{ color: '#000' }}>{step.result || '?'}</strong>
      </li>
    );
  };

  // Helper function untuk merender satu jalur resep lengkap dalam teks (daftar ordered list)
  // Memperbaiki key untuk <ol>
  const renderPath = (path, pathIndex = null) => (
    // --- PERBAIKAN KEY ---
    // Key untuk <ol> harus unik jika ada multiple list <ol> (di mode multiple)
    // pathIndex unik untuk setiap jalur di mode multiple.
    // Gunakan pathIndex untuk keunikan, fallback ke key statis 'single-path-list' untuk kasus single path.
    // SINTAKS template literal untuk key HARUS: `path-${pathIndex}` jika pathIndex tidak null
    Array.isArray(path) ? (
        <ol key={pathIndex !== null ? `path-${pathIndex}` : 'single-path-list'} style={{ paddingLeft: '20px', marginTop: '5px' }}>
          {/* Map setiap langkah resep dan render menggunakan renderStep */}
          {/* renderStep sekarang menggunakan key gabungan pathIndex-stepIndex */}
          {path.map((step, stepIndex) => renderStep(step, stepIndex, pathIndex))}
        </ol>
    ) : <p key={pathIndex !== null ? `invalid-path-list-${pathIndex}` : 'invalid-single-path-list'} style={{ color: '#dc3545' }}>Data jalur tidak valid (bukan array).</p>
  );
   // --- AKHIR FUNGSI HELPER UNTUK TAMPILAN TEKS ---


  // Jika hasil ditemukan, siapkan data untuk visualisasi
  // Kita akan menampung data pohon di sini dalam format array of roots ([root1, root2, ...])
  // Untuk mode shortest, array ini hanya berisi 1 root. Untuk mode multiple, array ini berisi root untuk setiap jalur.
  let treeDataForRendering = [];

  // Pastikan results.pathFound adalah true sebelum mencoba memproses data visualisasi
  if (results.pathFound === true) {
      // Kita hanya perlu data tree untuk visualisasi jika target bukan elemen dasar dan ada path
      if (!isBaseElement(results.searchTarget)) {
          if (results.mode === 'shortest' && results.path) {
               // Untuk mode shortest, panggil buildTreeData sekali untuk jalur tunggal
               // buildTreeData mengembalikan array [rootNode]
               treeDataForRendering = buildTreeData(results.path, results.searchTarget, results.imageURLs);
          } else if (results.mode === 'multiple' && results.paths && results.paths.length > 0) {
               // Untuk mode multiple, results.paths adalah array of paths.
               // Kita akan panggil buildTreeData untuk setiap *path* individu dan mengumpulkan rootnya
               treeDataForRendering = results.paths.map(path => buildTreeData(path, results.searchTarget, results.imageURLs)[0]); // Ambil root node dari hasil buildTreeData
          }
      } else {
           // Kasus khusus: target adalah elemen dasar, pathFound true tapi path kosong/tidak perlu diproses sbg jalur
           // Buat simpul target saja untuk visualisasi (pohon 1 node)
           treeDataForRendering = buildTreeData([], results.searchTarget, results.imageURLs);
      }
  }


  return (
    <div style={{ marginTop: '30px', borderTop: '2px solid #0d6efd', paddingTop: '20px', background: '#ffffff', padding: '20px', borderRadius: '8px', boxShadow: '0 2px 4px rgba(0,0,0,0.1)', color: '#212529' }}>

      {/* --- Bagian Judul dan Info Pencarian --- */}
      <h2 style={{ borderBottom: '1px solid #dee2e6', paddingBottom: '10px', marginBottom: '20px', color: '#343a40' }}>
        Hasil Pencarian untuk: <strong style={{ color: '#0d6efd' }}>{results.searchTarget || 'N/A'}</strong>
        <span style={{ fontSize: '0.9em', color: '#6c757d', marginLeft: '10px' }}>
          ({results.algorithm?.toUpperCase() || 'N/A'}/{results.mode || 'N/A'}
          {results.mode === 'multiple' && ` - Max: ${results.maxRecipes || 'N/A'}`})
        </span>
      </h2>

      {/* Tampilkan info nodesVisited dan duration hanya jika pathFound true */}
      {results.pathFound === true && (
         <div style={{ marginBottom: '20px', display: 'flex', justifyContent: 'space-between', flexWrap: 'wrap', gap: '10px', fontSize: '0.95em', color: '#495057' }}>
             <span>Node Dikunjungi: <strong>{results.nodesVisited !== undefined && results.nodesVisited !== -1 ? results.nodesVisited : 'N/A'}</strong></span>
             <span>Durasi: <strong>{results.durationMillis ?? 'N/A'} ms</strong></span>
         </div>
      )}


       {/* AREA UNTUK MENAMPILKAN HASIL TEKS */}
       {results.pathFound === true ? ( // Tampilkan area teks jika pathFound true
           <>
           {/* Tampilan teks untuk mode shortest */}
           {results.mode === 'shortest' && results.path && results.path.length > 0 && (
                <div style={{ background: '#f8f9fa', padding: '15px', borderRadius: '5px', border: '1px solid #dee2e6', marginBottom: '20px' }}>
                   <h3 style={{ marginTop: '0', marginBottom: '10px', color: '#495057' }}>Jalur Resep Terpendek (Teks):</h3>
                   <p style={{ marginBottom: '10px', fontSize: '0.9em', color: '#6c757d' }}>Jumlah Langkah: {results.path.length}</p>
                   {renderPath(results.path)} {/* Panggil renderPath untuk list teks */}
               </div>
           )}

           {/* Tampilan teks untuk mode multiple */}
           {results.mode === 'multiple' && results.paths && results.paths.length > 0 && (
              <div style={{ marginBottom: '20px' }}>
                 <h3 style={{ marginBottom: '15px', color: '#495057' }}>Jalur Resep Ditemukan (Teks):</h3>
                 {/* Loop setiap jalur dan tampilkan dalam bentuk daftar teks */}
                 {results.paths.map((path, index) => (
                     <div key={`text-path-${index}`} style={{ marginBottom: '15px', border: '1px solid #eee', borderRadius: '5px', padding: '15px' }}>
                         <h4 style={{ marginTop: '0', marginBottom: '10px', borderBottom: '1px solid #eee', paddingBottom: '5px', color: '#495057' }}>
                           Jalur {index + 1} (Langkah: {path.length})
                         </h4>
                         {renderPath(path, index)} {/* Panggil renderPath untuk list teks */}
                     </div>
                 ))}
              </div>
           )}

           {/* Pesan jika target adalah elemen dasar dan pathFound true (untuk tampilan teks) */}
           {results.pathFound === true && isBaseElement(results.searchTarget) && (
                // Cek juga jika path/paths benar-benar kosong, karena elemen dasar pathnya 0
               ((results.mode === 'shortest' && (!results.path || results.path.length === 0)) ||
                (results.mode === 'multiple' && (!results.paths || results.paths.length === 0 || (results.paths.length > 0 && results.paths.every(path => path.length === 0)))))
           ) && ( // Pastikan hanya tampil jika memang elemen dasar DAN path kosong
                <div style={{ color: '#6c757d', border: '1px solid #ced4da', background: '#e9ecef', padding: '15px', borderRadius: '5px', marginTop: '20px', marginBottom: '20px' }}>
                     (Target adalah elemen dasar, tidak ada resep pembuat.)
                </div>
           )}
           </>
       ) : ( // Tampilkan pesan jika pathFound false (untuk kedua mode)
            <div style={{ color: '#856404', border: '1px solid #ffeeba', background: '#fff3cd', padding: '15px', borderRadius: '5px', marginTop: '20px' }}>
                 Jalur tidak ditemukan untuk elemen "{results.searchTarget || 'N/A'}". {results.error ? `(${results.error})` : ''}
            </div>
       )}


      {/* AREA UNTUK MENAMPILKAN VISUALISASI POHON */}
       {results.pathFound === true && treeDataForRendering.length > 0 && ( // Tampilkan area visualisasi jika pathFound true dan ada data pohon
          <>
              <h3 style={{ marginBottom: '15px', color: '#495057' }}>Visualisasi Pohon Resep:</h3>

             {/* Jika mode adalah 'shortest' */}
             {results.mode === 'shortest' && ( // Cek mode lagi untuk render Tree
                 <div style={{ background: '#f8f9fa', padding: '15px', borderRadius: '5px', border: '1px solid #dee2e6' }}>
                    <h4 style={{ marginTop: '0', marginBottom: '10px', color: '#495057' }}>Jalur Terpendek:</h4>
                     {/* KONTainer DENGAN UKURAN TETAP untuk komponen Tree. SANGAT PENTING! */}
                    <div id="treeWrapper-shortest" style={{ width: '100%', height: '500px', border: '1px solid #ccc', overflow: 'auto' }}> {/* Gunakan overflow: auto */}
                        {/* Komponen Tree yang akan menggambar pohon */}
                        <Tree
                            data={treeDataForRendering} // Data pohon mode shortest ([rootNode])
                            orientation="vertical" // Arah gambar pohon: 'vertical' atau 'horizontal'
                            translate={{ x: 250, y: 50 }} // Posisi awal simpul akar (root). SESUAIKAN.
                            renderCustomNodeElement={renderNodeWithImage} // Gunakan fungsi custom node kita
                            zoomable={true}
                            draggable={true}
                            // Props kustomisasi lain
                        />
                    </div>
                 </div>
              )}

              {results.mode === 'multiple' && results.paths && results.paths.length > 0 && ( // Cek mode lagi untuk render Trees
                 <div>
                    {/* Loop untuk SETIAP jalur resep yang ditemukan */}
                    {results.paths.map((path, index) => {
                        // Untuk setiap jalur, panggil buildTreeData untuk data pohon *spesifik jalur ini*
                        const singleTreeData = buildTreeData(path, results.searchTarget, results.imageURLs);
                        // Pastikan data pohon tidak kosong untuk jalur ini
                        if (singleTreeData.length === 0) return null; // Jangan render div dan Tree jika data kosong

                        return (
                             // Bungkus setiap visualisasi pohon dalam div tersendiri
                             <div key={`tree-path-${index}`} style={{ marginBottom: '20px', border: '1px solid #dee2e6', borderRadius: '5px', padding: '15px', background: '#ffffff' }}>
                                 <h4 style={{ marginTop: '0', marginBottom: '10px', borderBottom: '1px solid #eee', paddingBottom: '5px', color: '#495057' }}>
                                   Jalur {index + 1}:
                                 </h4>
                                 {/* KONTainer DENGAN UKURAN TETAP untuk komponen Tree. Sangat PENTING! */}
                                 <div id={`treeWrapper-multiple-${index}`} style={{ width: '100%', height: '500px', border: '1px solid #ccc', overflow: 'auto' }}> {/* Gunakan overflow: auto */}
                                     <Tree
                                         data={singleTreeData} // Data pohon untuk 1 jalur ([rootNode])
                                         orientation="vertical"
                                         translate={{ x: 250, y: 50 }} // Sesuaikan
                                         renderCustomNodeElement={renderNodeWithImage} // Gunakan fungsi custom node kita
                                         zoomable={true}
                                         draggable={true}
                                         // Props lain
                                     />
                                 </div>
                             </div>
                        );
                    })}
                 </div>
              )}

          </>
       )}

        {/* Tampilkan pesan jika pathFound true tapi tidak ada visualisasi POHON */}
        {/* Kondisi ini mengecek apakah hasil ditemukan (pathFound=true) tetapi treeDataForRendering kosong */}
        {results.pathFound === true && treeDataForRendering.length === 0 && (
            isBaseElement(results.searchTarget) ?
            // Pesan khusus untuk elemen dasar
               <div style={{ color: '#6c757d', border: '1px solid #ced4da', background: '#e9ecef', padding: '15px', borderRadius: '5px', marginTop: '20px' }}>
                    (Target adalah elemen dasar, tidak ada visualisasi pohon resep.)
               </div>
               :
               // Pesan fallback jika pathFound true tapi treeDataForRendering kosong (bukan elemen dasar)
               <div style={{ color: '#856404', border: '1px solid #ffeeba', background: '#fff3cd', padding: '15px', borderRadius: '5px', marginTop: '20px' }}>
                   Jalur resep ditemukan, tetapi visualisasi pohon tidak tersedia atau jalur kosong. Mohon periksa data resep dari backend.
               </div>
           )}


      {/* AREA UNTUK MENAMPILKAN URL GAMBAR */}
      {results.imageURLs && typeof results.imageURLs === 'object' && Object.keys(results.imageURLs).length > 0 && (
           <details style={{ marginTop: '20px', border: '1px solid #ddd', borderRadius: '5px', background: '#ffffff' }}>
              <summary style={{ padding: '10px', cursor: 'pointer', fontWeight: 'bold', color: '#495057' }}>Tampilkan/Sembunyikan URL Gambar Terkait</summary>
              <ul style={{ listStyle: 'none', padding: '15px', margin: '0', maxHeight: '150px', overflowY: 'auto' }}>
                {Object.entries(results.imageURLs).map(([name, url]) => (
                  <li key={name} style={{ marginBottom: '8px', display: 'flex', alignItems: 'center', color: '#212529' }}>
                    {/* Gunakan URL dari imageURLs yang sudah berisi path ke backend proxy */}
                    {/* url dari imageURLs seharusnya "/api/image?elementName=..." */}
                    {url ? (
                         <img
                           src={`${API_BASE_URL}${url}`} // <-- Gabungkan API_BASE_URL dengan path dari imageURLs
                           alt={name || '?'}
                           style={{ height: '24px', width: '24px', verticalAlign: 'middle', marginRight: '8px', border: '1px solid #eee', objectFit: 'contain' }}
                           onError={(e) => { e.target.style.display = 'none'; }}
                         />
                    ) : null}
                    <span style={{ fontWeight: '500', marginRight: '5px' }}>{name || 'N/A'}:</span>
                    {/* Link juga perlu digabungkan dengan API_BASE_URL */}
                    <a href={`${API_BASE_URL}${url}` || '#'} target="_blank" rel="noopener noreferrer" style={{ fontSize: '0.85em', color: '#0d6efd', wordBreak: 'break-all' }}>{url || 'N/A'}</a> {/* Juga perbaiki link di sini */}
                  </li>
                ))}
              </ul>
           </details>
        )}


    </div>
  );
}

export default SearchResults;