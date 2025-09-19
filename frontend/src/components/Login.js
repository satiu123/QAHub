import React, { useState } from 'react';
import axios from 'axios';

const API_URL = 'http://localhost:8081/api/v1';

function Login() {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [message, setMessage] = useState('');
    const [error, setError] = useState('');
    const [token, setToken] = useState('');

    const handleSubmit = async (e) => {
        e.preventDefault();
        setMessage('');
        setError('');
        setToken('');

        try {
            const response = await axios.post(`${API_URL}/login`, {
                username,
                password
            });
            setToken(response.data.token);
            setMessage('Login successful!');
        } catch (err) {
            setError(err.response?.data?.error || 'An error occurred during login.');
        }
    };

    return (
        <div>
            <h2>Login</h2>
            <form onSubmit={handleSubmit}>
                <div className="mb-3">
                    <label className="form-label">Username</label>
                    <input
                        type="text"
                        className="form-control"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        required
                    />
                </div>
                <div className="mb-3">
                    <label className="form-label">Password</label>
                    <input
                        type="password"
                        className="form-control"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        required
                    />
                </div>
                <button type="submit" className="btn btn-primary">Login</button>
            </form>
            {message && <div className="alert alert-success mt-3">{message}</div>}
            {error && <div className="alert alert-danger mt-3">{error}</div>}
            {token && (
                <div className="alert alert-info mt-3">
                    <strong>JWT Token:</strong>
                    <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>{token}</pre>
                </div>
            )}
        </div>
    );
}

export default Login;
