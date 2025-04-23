// Contoh menggunakan fetch (lebih detail implementasinya nanti saat coding React)

// Alamat dasar API backend Anda
const API_BASE_URL = "http://localhost:8080"; // Atau alamat backend Anda nanti

/**
 * Fungsi untuk memanggil endpoint /api/search
 * @param {string} target Nama elemen target
 * @param {string} algo Algoritma ('bfs' atau 'dfs')
 * @param {string} mode Mode ('shortest' atau 'multiple')
 * @returns {Promise<object>} Promise yang resolve dengan data JSON dari API
 */
async function findRecipes(target, algo, mode) {
  // Buat query string dari parameter
  const params = new URLSearchParams({ target, algo, mode });
  const url = `${API_BASE_URL}/api/search?${params.toString()}`;

  console.log(`Workspaceing: ${url}`); // Untuk debugging

  try {
    const response = await fetch(url);

    if (!response.ok) {
      // Jika status response bukan 2xx (misal 400, 404, 500)
      const errorData = await response.json().catch(() => ({ message: response.statusText })); // Coba parse error JSON, fallback ke status text
      throw new Error(`API Error (${response.status}): ${errorData.error || 'Unknown error'}`);
    }

    const data = await response.json(); // Parse response body sebagai JSON
    return data; // Kembalikan data hasil (objek SearchResponse dari backend)

  } catch (error) {
    console.error("Gagal mengambil resep dari API:", error);
    // Lempar ulang error agar bisa ditangani oleh komponen React pemanggil
    throw error;
  }
}

// Ekspor fungsi agar bisa digunakan di komponen React lain
export { findRecipes };