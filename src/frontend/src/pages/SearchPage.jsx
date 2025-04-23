// src/frontend/src/pages/SearchPage.jsx
import React, { useState } from 'react';
import SearchForm from '../components/SearchForm';
import SearchResults from '../components/SearchResults'; // Buat file ini dulu (bisa kosong)
import { findRecipes } from '../api/searchService'; // Impor fungsi API

function SearchPage() {
  const [searchResults, setSearchResults] = useState(null); // Untuk menyimpan hasil dari API
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);

  // Fungsi ini akan dipanggil oleh SearchForm saat disubmit
  const handleSearch = async (searchParams) => {
    console.log('SearchPage: Menerima parameter ->', searchParams);
    setIsLoading(true);
    setError(null);
    setSearchResults(null); // Kosongkan hasil sebelumnya

    try {
      // Panggil fungsi API dari searchService.js
      const data = await findRecipes(searchParams.target, searchParams.algo, searchParams.mode);
      setSearchResults(data); // Simpan hasil ke state
    } catch (err) {
      setError(err.message || 'Terjadi kesalahan saat mencari resep.'); // Simpan pesan error
    } finally {
      setIsLoading(false); // Set loading selesai (baik sukses maupun error)
    }
  };

  return (
    <div>
      <h1>Pencari Resep Little Alchemy 2</h1>
      <SearchForm onSearchSubmit={handleSearch} isLoading={isLoading} />
      <hr />
      {/* Kirim hasil, loading, dan error ke komponen SearchResults */}
      <SearchResults results={searchResults} isLoading={isLoading} error={error} />
    </div>
  );
}

export default SearchPage;