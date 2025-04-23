// src/frontend/src/api/searchService.js

// Alamat dasar API backend Anda
const API_BASE_URL = "http://localhost:8080"; // Sesuaikan jika perlu

/**
 * Fungsi untuk memanggil endpoint /api/search
 * @param {string} target Nama elemen target
 * @param {string} algo Algoritma ('bfs' atau 'dfs')
 * @param {string} mode Mode ('shortest' atau 'multiple')
 * @param {number} [maxRecipes] Jumlah maksimal resep (hanya untuk mode 'multiple')
 * @returns {Promise<object>} Promise yang resolve dengan data JSON dari API
 */
async function findRecipes(target, algo, mode, maxRecipes) { // Tambahkan parameter maxRecipes
  // Buat query string dari parameter dasar
  const params = new URLSearchParams({ target, algo, mode });

  // Tambahkan parameter 'max' HANYA jika mode 'multiple' dan maxRecipes valid
  if (mode === 'multiple' && maxRecipes && maxRecipes > 0) {
    params.append('max', maxRecipes.toString()); // Tambahkan parameter 'max'
  }

  const url = `${API_BASE_URL}/api/search?${params.toString()}`;

  console.log(`Frontend: Mengirim request ke: ${url}`); // Untuk debugging

  try {
    const response = await fetch(url);

    if (!response.ok) {
      // Tangani error HTTP
      const errorData = await response.json().catch(() => ({ message: response.statusText }));
      // Coba ambil pesan error spesifik dari backend jika ada
      const backendErrorMessage = errorData.error || 'Unknown API error';
      throw new Error(`API Error (${response.status}): ${backendErrorMessage}`);
    }

    const data = await response.json(); // Parse response body sebagai JSON
    console.log("Frontend: Menerima data:", data); // Debug: lihat data yang diterima
    return data; // Kembalikan data hasil (objek MultiSearchResponse dari backend)

  } catch (error) {
    console.error("Frontend: Gagal mengambil resep dari API:", error);
    // Lempar ulang error agar bisa ditangani oleh komponen React pemanggil
    throw error;
  }
}

// Ekspor fungsi agar bisa digunakan di komponen React lain
export { findRecipes };
