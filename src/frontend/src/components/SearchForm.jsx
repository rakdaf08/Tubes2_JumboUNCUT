// src/frontend/src/components/SearchForm.jsx
import React, { useState } from 'react';

// Menerima prop 'onSearchSubmit' dari parent (SearchPage)
function SearchForm({ onSearchSubmit, isLoading }) {
  const [target, setTarget] = useState('');
  const [algo, setAlgo] = useState('bfs'); // Default BFS
  const [mode, setMode] = useState('shortest'); // Default shortest
  const [maxRecipes, setMaxRecipes] = useState(1); // State baru untuk max recipes, default 1

  const handleSubmit = (event) => {
    event.preventDefault(); // Mencegah refresh halaman standar form HTML
    if (!target) {
      alert('Masukkan elemen target!');
      return;
    }
    // Jika mode multiple, pastikan maxRecipes valid
    if (mode === 'multiple' && (!maxRecipes || maxRecipes <= 0)) {
        alert('Masukkan jumlah resep minimal 1 untuk mode multiple!');
        return;
    }

    // Panggil fungsi yang di-pass dari parent dengan data dari state
    // Sertakan maxRecipes HANYA jika mode adalah 'multiple'
    const searchParams = { target, algo, mode };
    if (mode === 'multiple') {
        searchParams.max = maxRecipes; // Gunakan key 'max' sesuai backend handler
    }
    onSearchSubmit(searchParams);
  };

  // Definisikan style untuk label agar konsisten
  const labelStyle = {
    cursor: 'pointer',
    color: '#212529', // Warna teks label diubah menjadi gelap
    display: 'inline-flex', // Agar radio button dan teks sejajar
    alignItems: 'center', // Sejajarkan item di tengah secara vertikal
  };

  const radioInputStyle = {
      marginRight: '5px', // Jarak antara radio button dan teks
  };

  return (
    // Styling dasar agar lebih rapi (bisa dipindah ke CSS)
    // Hapus maxWidth dan sesuaikan margin agar form mengambil lebar penuh
    <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '15px', width: '1200px', margin: '20px auto', padding: '20px', border: '1px solid #ccc', borderRadius: '8px', background: '#f9f9f9', boxSizing: 'border-box' }}> {/* Hapus maxWidth, set width 100%, tambahkan boxSizing */}
      <div>
        <label htmlFor="targetElement" style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold', color: '#343a40' /* Warna label utama */ }}>Elemen Target:</label>
        <input
          type="text"
          id="targetElement"
          value={target}
          onChange={(e) => setTarget(e.target.value)}
          placeholder="Contoh: Mud, Human, ..."
          required
          style={{ width: '100%', padding: '10px', boxSizing: 'border-box', borderRadius: '4px', border: '1px solid #ccc' }}
        />
      </div>

      <div style={{ border: '1px solid #ddd', padding: '15px', borderRadius: '5px', background: '#fff' }}>
        <p style={{ marginTop: '0', marginBottom: '10px', fontWeight: 'bold', color: '#343a40' /* Warna judul bagian */ }}>Algoritma:</p>
        <div style={{ display: 'flex', gap: '20px' }}>
            {/* Terapkan style ke label */}
            <label style={labelStyle}>
              <input
                type="radio"
                value="bfs"
                checked={algo === 'bfs'}
                onChange={(e) => setAlgo(e.target.value)}
                style={radioInputStyle}
              /> BFS (Shortest Path)
            </label>
            {/* Terapkan style ke label */}
            <label style={labelStyle}>
              <input
                type="radio"
                value="dfs"
                checked={algo === 'dfs'}
                onChange={(e) => setAlgo(e.target.value)}
                 style={radioInputStyle}
              /> DFS (A Path / Multiple)
            </label>
        </div>
      </div>

      <div style={{ border: '1px solid #ddd', padding: '15px', borderRadius: '5px', background: '#fff' }}>
        <p style={{ marginTop: '0', marginBottom: '10px', fontWeight: 'bold', color: '#343a40' /* Warna judul bagian */ }}>Mode:</p>
        <div style={{ display: 'flex', gap: '20px', alignItems: 'center', flexWrap: 'wrap' }}>
            {/* Terapkan style ke label */}
            <label style={labelStyle}>
              <input
                type="radio"
                value="shortest"
                checked={mode === 'shortest'}
                onChange={(e) => setMode(e.target.value)}
                 style={radioInputStyle}
              /> Shortest
            </label>
             {/* Terapkan style ke label */}
            <label style={labelStyle}>
              <input
                type="radio"
                value="multiple"
                checked={mode === 'multiple'}
                onChange={(e) => setMode(e.target.value)}
                 style={radioInputStyle}
              /> Multiple
            </label>
        </div>
         {/* Tampilkan input jumlah HANYA jika mode 'multiple' dipilih */}
         {mode === 'multiple' && (
            <div style={{ marginTop: '15px' }}>
                 {/* Warna label input jumlah */}
                 <label htmlFor="maxRecipes" style={{ marginRight: '8px', fontWeight: '500', color: '#343a40' }}>Jumlah Resep:</label>
                 <input
                    type="number"
                    id="maxRecipes"
                    value={maxRecipes}
                    onChange={(e) => setMaxRecipes(parseInt(e.target.value, 10) || 1)} // Pastikan integer, min 1
                    min="1"
                    style={{ width: '70px', padding: '8px', borderRadius: '4px', border: '1px solid #ccc' }}
                 />
            </div>
         )}
      </div>

      <button
        type="submit"
        disabled={isLoading}
        style={{
            padding: '12px 20px',
            cursor: isLoading ? 'not-allowed' : 'pointer',
            background: isLoading ? '#ccc' : '#0d6efd', // Warna biru sedikit berbeda
            color: 'white',
            border: 'none',
            borderRadius: '5px',
            fontSize: '16px',
            fontWeight: 'bold',
            transition: 'background-color 0.2s ease'
         }}
         onMouseOver={(e) => { if (!isLoading) e.currentTarget.style.background = '#0b5ed7'; }} // Warna hover
         onMouseOut={(e) => { if (!isLoading) e.currentTarget.style.background = '#0d6efd'; }}
      >
        {isLoading ? 'Mencari...' : 'Cari Resep'}
      </button>
    </form>
  );
}

export default SearchForm;
