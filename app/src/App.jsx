import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Login from './Login';
import Monitor from './Monitor';
import Testing from './Testing';
import Sidebar from './components/Sidebar';

function App() {
  return (
    <Router>
      <div style={{ display: 'flex' }}>
        <Sidebar />
        <div style={{ marginLeft: '220px', padding: '20px' }}>
          <Routes>
            <Route path="/" element={<Login />} />
            <Route path="/login" element={<Login />} />
            <Route path="/monitor" element={<Monitor />} />
            <Route path="/testing" element={<Testing />} />
          </Routes>
        </div>
      </div>
    </Router>
  );
}

export default App;
