// src/frontend/src/App.jsx
import React, { useState, useEffect } from 'react';
import './App.css'; // CSS untuk App, splash, logo, sans, dialog box
import SearchPage from './pages/SearchPage'; // Akan ditampilkan nanti

// Impor gambar Anda dari direktori assets
import jumboKocokLogo from './assets/Jumbo_Kocok.png';
import clickToStartImage from './assets/ClickToStart.png';
import sansImage from './assets/sans.jpg'; // Pastikan nama file ini benar

function App() {
  const [appState, setAppState] = useState('splash');
  const [logoPositionClass, setLogoPositionClass] = useState('center');

  const handleInitialClick = () => {
    if (appState === 'splash') {
      setLogoPositionClass('topLeft');
      setAppState('logoMoving');
    }
  };

  useEffect(() => {
    let timer;
    if (appState === 'logoMoving') {
      timer = setTimeout(() => {
        setAppState('sansAppearing');
      }, 1000); // Durasi animasi logo
    } else if (appState === 'sansAppearing') {
      timer = setTimeout(() => {
        setAppState('sansBobbing'); // State di mana Sans sudah muncul dan mulai bobbing
                                    // Kotak dialog juga akan terlihat di state ini
      }, 500); // Durasi fade-in Sans dan kotak
    }
    return () => clearTimeout(timer);
  }, [appState]);

  return (
    <div className={`App-container current-state-${appState}`}>
      <img
        src={jumboKocokLogo}
        alt="Jumbo Kocok Logo"
        className={`main-logo ${logoPositionClass}`}
      />

      {appState === 'splash' && (
        <div className="splash-initial-content" onClick={handleInitialClick}>
          <img
            src={clickToStartImage}
            alt="Click To Start"
            className="splash-start-text"
          />
        </div>
      )}

      {/* Gambar Sans dan Kotak Dialog akan muncul bersamaan atau berdekatan */}
      {(appState === 'sansAppearing' || appState === 'sansBobbing') && (
        <> {/* Menggunakan Fragment agar bisa merender dua elemen sibling */}
          <img
            src={sansImage}
            alt="Sans"
            className={`sans-image ${appState === 'sansBobbing' ? 'bobbing' : ''}`}
          />
          <div className={`dialog-box ${appState === 'sansBobbing' ? 'visible' : ''}`}>
            {/* Teks akan dimasukkan di sini nanti */}
            <p>* Please enter the recipe you are looking for</p>
          </div>
        </>
      )}

      {appState === 'searchPage' && (
        <div className="search-page-wrapper">
          <SearchPage />
        </div>
      )}
    </div>
  );
}

export default App;