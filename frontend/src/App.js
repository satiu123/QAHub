import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Link, Navigate, useNavigate } from 'react-router-dom';
import axios from 'axios';
import Login from './components/Login';
import Register from './components/Register';
import Profile from './components/Profile';

const API_URL = 'http://localhost:8081/api/v1';

const Layout = () => {
  const [token, setToken] = useState(localStorage.getItem('token'));
  const navigate = useNavigate();

  const handleLogin = (token) => {
    setToken(token);
    localStorage.setItem('token', token);
    navigate('/profile');
  };

  const handleLogout = async () => {
    try {
      await axios.post(`${API_URL}/logout`, null, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
    } catch (error) {
      console.error('Error logging out from server:', error);
    } finally {
      setToken(null);
      localStorage.removeItem('token');
      navigate('/login');
    }
  };

  return (
    <div className="container mt-4">
      <nav className="navbar navbar-expand-lg navbar-light bg-light mb-4">
        <div className="container-fluid">
          <Link className="navbar-brand" to="/">QAHub</Link>
          <div className="navbar-nav">
            {token ? (
              <>
                <Link className="nav-link" to="/profile">Profile</Link>
                <button onClick={handleLogout} className="btn btn-link nav-link">Logout</button>
              </>
            ) : (
              <>
                <Link className="nav-link" to="/login">Login</Link>
                <Link className="nav-link" to="/register">Register</Link>
              </>
            )}
          </div>
        </div>
      </nav>

      <Routes>
        <Route path="/login" element={<Login onLogin={handleLogin} />} />
        <Route path="/register" element={<Register />} />
        <Route 
          path="/profile" 
          element={token ? <Profile token={token} onLogout={handleLogout} /> : <Navigate to="/login" />}
        />
        <Route path="*" element={<Navigate to={token ? "/profile" : "/login"} />} />
      </Routes>
    </div>
  );
}

function App() {
  return (
    <Router>
      <Layout />
    </Router>
  );
}

export default App;