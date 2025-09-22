import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Link, Navigate, useNavigate } from 'react-router-dom';
import axios from 'axios';
import Login from './components/Login';
import Register from './components/Register';
import Profile from './components/Profile';
import QuestionList from './components/QuestionList';
import QuestionDetail from './components/QuestionDetail';
import CreateQuestion from './components/CreateQuestion';
import SearchBox from './components/SearchBox';
import SearchResult from './components/SearchResult';


const API_URL = 'http://localhost:8080/api/v1';

const Layout = () => {
  const [token, setToken] = useState(localStorage.getItem('token'));
  const navigate = useNavigate();

  const handleLogin = (token) => {
    setToken(token);
    localStorage.setItem('token', token);
    navigate('/questions');
  };

  const handleLogout = async () => {
    try {
      await axios.post(`${API_URL}/users/logout`, null, {
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
          <div className="navbar-nav me-auto">
            {token ? (
              <>
                <Link className="nav-link" to="/questions">Questions</Link>
                <Link className="nav-link" to="/profile">Profile</Link>
              </>
            ) : (
              <>
                <Link className="nav-link" to="/login">Login</Link>
                <Link className="nav-link" to="/register">Register</Link>
              </>
            )}
          </div>
          {token && (
            <div className="d-flex align-items-center">
              <SearchBox />
              <button onClick={handleLogout} className="btn btn-link nav-link ms-2">Logout</button>
            </div>
          )}
        </div>
      </nav>

      <Routes>
        <Route path="/login" element={<Login onLogin={handleLogin} />} />
        <Route path="/register" element={<Register />} />
        <Route
          path="/profile"
          element={token ? <Profile token={token} onLogout={handleLogout} /> : <Navigate to="/login" />}
        />
        <Route
          path="/questions"
          element={token ? <QuestionList token={token} /> : <Navigate to="/login" />}
        />
        <Route
          path="/questions/:questionId"
          element={token ? <QuestionDetail token={token} /> : <Navigate to="/login" />}
        />
        <Route
          path="/create-question"
          element={token ? <CreateQuestion token={token} /> : <Navigate to="/login" />}
        />
        <Route 
          path="/search"
          element={token ? <SearchResult token={token} /> : <Navigate to="/login" />}
        />
        <Route path="*" element={<Navigate to={token ? "/questions" : "/login"} />} />
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