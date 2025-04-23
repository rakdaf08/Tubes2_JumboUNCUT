// src/frontend/src/components/SearchResults.jsx
import React from 'react';

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
  if (!results || typeof results !== 'object') {
    return <p style={{ textAlign: 'center', marginTop: '20px', color: '#6c757d' }}>Masukkan elemen target dan klik cari.</p>;
  }

  // --- PINDAHKAN DEFINISI handleImageError KE SINI ---
  // Fallback jika gambar gagal load
  const handleImageError = (e) => {
      console.warn(`Image failed to load: ${e.target.src}`); // Tambahkan log untuk debug
      e.target.style.display = 'none'; // Sembunyikan gambar yang error
      // Opsional: Tampilkan placeholder atau teks alternatif
      const altText = e.target.alt || 'image';
      const errorSpan = document.createElement('span');
      errorSpan.textContent = `(${altText} error)`;
      errorSpan.style.fontSize = '0.8em';
      errorSpan.style.color = '#dc3545';
      e.target.parentNode?.insertBefore(errorSpan, e.target.nextSibling); // Tambahkan span setelah img jika parent ada
  };
  // ----------------------------------------------------

  // Helper function untuk merender satu langkah resep dengan gambar
  const renderStep = (step, stepIndex, pathIndex = null) => {
    if (!step || typeof step !== 'object') {
        return <li key={`${pathIndex}-${stepIndex}`} style={{color: 'red'}}>Data langkah tidak valid</li>;
    }
    const key = `${pathIndex}-${stepIndex}`;
    const imageUrl1 = results.imageURLs?.[step.ingredient1];
    const imageUrl2 = results.imageURLs?.[step.ingredient2];
    const imageUrlResult = results.imageURLs?.[step.result];

    const imgStyle = {
        height: '20px',
        width: '20px',
        verticalAlign: 'middle',
        margin: '0 3px',
        border: '1px solid #eee',
        objectFit: 'contain'
    };

    // handleImageError sekarang sudah didefinisikan di scope luar

    return (
      <li key={key} style={{ marginBottom: '8px', lineHeight: '1.5', color: '#212529' }}>
        {imageUrl1 ? <img src={imageUrl1} alt={step.ingredient1 || 'ingredient1'} style={imgStyle} onError={handleImageError}/> : null}
        {step.ingredient1 || '?'}
        <span style={{ margin: '0 5px' }}>+</span>
        {imageUrl2 ? <img src={imageUrl2} alt={step.ingredient2 || 'ingredient2'} style={imgStyle} onError={handleImageError}/> : null}
        {step.ingredient2 || '?'}
        <span style={{ margin: '0 5px' }}>{' => '}</span>
        {imageUrlResult ? <img src={imageUrlResult} alt={step.result || 'result'} style={imgStyle} onError={handleImageError}/> : null}
        <strong style={{ color: '#000' }}>{step.result || '?'}</strong>
      </li>
    );
  };

  // Helper function untuk merender satu jalur resep lengkap
  const renderPath = (path, pathIndex = null) => (
    Array.isArray(path) ? (
        <ol key={pathIndex} style={{ paddingLeft: '20px', marginTop: '5px' }}>
          {path.map((step, index) => renderStep(step, index, pathIndex))}
        </ol>
    ) : <p style={{ color: '#dc3545' }}>Data jalur tidak valid (bukan array).</p>
  );

  // Gunakan optional chaining (?.) saat mengakses properti results
  return (
    <div style={{ marginTop: '30px', borderTop: '2px solid #0d6efd', paddingTop: '20px', background: '#ffffff', padding: '20px', borderRadius: '8px', boxShadow: '0 2px 4px rgba(0,0,0,0.1)', color: '#212529' }}>
      <h2 style={{ borderBottom: '1px solid #dee2e6', paddingBottom: '10px', marginBottom: '20px', color: '#343a40' }}>
        Hasil Pencarian untuk: <strong style={{ color: '#0d6efd' }}>{results.searchTarget || 'N/A'}</strong>
        <span style={{ fontSize: '0.9em', color: '#6c757d', marginLeft: '10px' }}>
          ({results.algorithm?.toUpperCase() || 'N/A'}/{results.mode || 'N/A'}
          {results.mode === 'multiple' && ` - Max: ${results.maxRecipes || 'N/A'}`})
        </span>
      </h2>

      {results.pathFound === true ? (
        <>
          <div style={{ marginBottom: '20px', display: 'flex', justifyContent: 'space-between', flexWrap: 'wrap', gap: '10px', fontSize: '0.95em', color: '#495057' }}>
            <span>Node Dikunjungi: <strong>{results.nodesVisited !== undefined && results.nodesVisited !== -1 ? results.nodesVisited : 'N/A'}</strong></span>
            <span>Durasi: <strong>{results.durationMillis ?? 'N/A'} ms</strong></span>
          </div>

          {/* Tampilkan hasil berdasarkan mode */}
          {results.mode === 'shortest' ? (
            // --- Tampilan Mode Shortest ---
            <div style={{ background: '#f8f9fa', padding: '15px', borderRadius: '5px', border: '1px solid #dee2e6' }}>
              <h3 style={{ marginTop: '0', marginBottom: '10px', color: '#495057' }}>Jalur Resep Terpendek:</h3>
              {Array.isArray(results.path) && results.path.length > 0 ? (
                <>
                  <p style={{ marginBottom: '10px', fontSize: '0.9em', color: '#6c757d' }}>Jumlah Langkah: {results.path.length}</p>
                  {renderPath(results.path)}
                </>
              ) : (
                 results.searchTarget && ['Air', 'Earth', 'Fire', 'Water'].includes(results.searchTarget)
                 ? <p style={{ color: '#6c757d' }}>(Target adalah elemen dasar, tidak ada resep.)</p>
                 : <p style={{ color: '#6c757d' }}>(Jalur resep tidak ditemukan atau kosong untuk mode shortest.)</p>
              )}
            </div>
          ) : results.mode === 'multiple' ? (
            // --- Tampilan Mode Multiple ---
            <div>
              <h3 style={{ marginBottom: '15px', color: '#495057' }}>
                Jalur Resep Ditemukan ({Array.isArray(results.paths) ? results.paths.length : 0} dari max {results.maxRecipes || 'N/A'}):
              </h3>
              {Array.isArray(results.paths) && results.paths.length > 0 ? (
                results.paths.map((path, index) => (
                  <div key={index} style={{ marginBottom: '20px', border: '1px solid #dee2e6', borderRadius: '5px', padding: '15px', background: '#ffffff' }}>
                    <h4 style={{ marginTop: '0', marginBottom: '10px', borderBottom: '1px solid #eee', paddingBottom: '5px', color: '#495057' }}>
                      Jalur {index + 1} (Langkah: {Array.isArray(path) ? path.length : 0})
                    </h4>
                    {renderPath(path, index)}
                  </div>
                ))
              ) : (
                 <p style={{ color: '#6c757d' }}>(Tidak ada jalur ditemukan untuk mode multiple dengan kriteria ini)</p>
              )}
            </div>
          ) : null }

          {/* Bagian URL Gambar */}
          {results.imageURLs && typeof results.imageURLs === 'object' && Object.keys(results.imageURLs).length > 0 && (
             <details style={{ marginTop: '20px', border: '1px solid #ddd', borderRadius: '5px', background: '#ffffff' }}>
                <summary style={{ padding: '10px', cursor: 'pointer', fontWeight: 'bold', color: '#495057' }}>Tampilkan/Sembunyikan URL Gambar Terkait</summary>
                <ul style={{ listStyle: 'none', padding: '15px', margin: '0', maxHeight: '150px', overflowY: 'auto' }}>
                  {/* Gunakan Object.entries untuk iterasi map */}
                  {Object.entries(results.imageURLs).map(([name, url]) => (
                    <li key={name} style={{ marginBottom: '8px', display: 'flex', alignItems: 'center', color: '#212529' }}>
                      <img
                        src={url || ''}
                        alt={name || '?'}
                        style={{ height: '24px', width: '24px', verticalAlign: 'middle', marginRight: '8px', border: '1px solid #eee', objectFit: 'contain' }}
                        // Panggil handleImageError yang sudah didefinisikan di scope luar
                        onError={handleImageError}
                      />
                      <span style={{ fontWeight: '500', marginRight: '5px' }}>{name || 'N/A'}:</span>
                      <a href={url || '#'} target="_blank" rel="noopener noreferrer" style={{ fontSize: '0.85em', color: '#0d6efd', wordBreak: 'break-all' }}>{url || 'N/A'}</a>
                    </li>
                  ))}
                </ul>
             </details>
          )}

        </>
      ) : (
        // --- Tampilan Jika Jalur Tidak Ditemukan (pathFound === false) ---
        <div style={{ color: '#856404', border: '1px solid #ffeeba', background: '#fff3cd', padding: '15px', borderRadius: '5px', marginTop: '20px' }}>
            Jalur tidak ditemukan untuk elemen "{results.searchTarget || 'N/A'}". {results.error ? `(${results.error})` : ''}
        </div>
      )}
      {/* ... (kode debug JSON seperti sebelumnya) ... */}
    </div>
  );
}

export default SearchResults;
