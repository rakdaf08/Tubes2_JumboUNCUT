// src/frontend/src/pages/SearchPage.jsx
import React, { useState, useEffect } from 'react'; // Import useEffect jika ingin membersihkan error/hasil saat parameter berubah
import SearchForm from '../components/SearchForm';
import SearchResults from '../components/SearchResults';
import { findRecipes } from '../api/searchService';

function SearchPage() {
  const [searchResults, setSearchResults] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [currentParams, setCurrentParams] = useState(null); // Opsional: simpan parameter pencarian terakhir

  // Fungsi ini akan dipanggil oleh SearchForm saat disubmit
  const handleSearch = async (searchParams) => {
    console.log('SearchPage: Menerima parameter ->', searchParams);
    setCurrentParams(searchParams); // Simpan parameter saat ini
    setIsLoading(true);
    setError(null);
    setSearchResults(null); // Kosongkan hasil sebelumnya

    try {
      // --- PERBAIKAN DI SINI ---
      // Ekstrak semua parameter yang relevan, termasuk 'max' jika ada
      const { target, algo, mode, max } = searchParams;
      // Panggil findRecipes dengan semua argumen yang diperlukan
      const data = await findRecipes(target, algo, mode, max); // Kirim 'max' sebagai argumen ke-4
      // ------------------------
      setSearchResults(data); // Simpan hasil ke state
    } catch (err) {
      // Pastikan kita menangkap dan menampilkan pesan error dari findRecipes
      setError(err.message || 'Terjadi kesalahan saat mencari resep.');
      console.error("SearchPage Error:", err); // Log error lengkap di console frontend
    } finally {
      setIsLoading(false); // Set loading selesai (baik sukses maupun error)
    }
  };

  // Opsional: Reset hasil jika input form berubah (misalnya target dikosongkan)
  // useEffect(() => {
  //   if (!currentParams?.target) { // Jika target kosong di parameter terakhir
  //      setSearchResults(null);
  //      setError(null);
  //   }
  // }, [currentParams]);


  return (
    // Terapkan gaya untuk membuat kontainer ini full width dan tambahkan padding horizontal
    <div style={{ width: '100%', margin: '0 auto', padding: '0 20px', boxSizing: 'border-box' }}> {/* Set width 100%, hapus margin auto vertikal, tambahkan padding horizontal */}
      <h1 style={{ textAlign: 'center', marginBottom: '30px' }}>Pencari Resep Little Alchemy 2</h1> {/* Hapus padding horizontal dari judul */}
      {/* Hapus div pembungkus form karena padding horizontal sudah di container utama */}
      {/* <div style={{ padding: '0 20px' }}> */}
         <SearchForm onSearchSubmit={handleSearch} isLoading={isLoading} />
      {/* </div> */}


      {/* Komponen SearchResults akan menampilkan hasil, loading, atau error */}
      {/* SearchResults sudah diberi width: '100%' dan margin: 'auto' */}
      <SearchResults results={searchResults} isLoading={isLoading} error={error} />
    </div>
  );
}

export default SearchPage;
