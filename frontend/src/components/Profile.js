import React, { useState, useEffect } from 'react';
import axios from 'axios';

const API_URL = 'http://localhost:8081/api/v1';

function Profile({ token, onLogout }) {
    const [user, setUser] = useState(null);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchUser = async () => {
            try {
                const response = await axios.get(`${API_URL}/profile`, {
                    headers: {
                        'Authorization': `Bearer ${token}`
                    }
                });
                setUser(response.data);
            } catch (err) {
                setError('Failed to fetch user data.');
            }
        };

        if (token) {
            fetchUser();
        }
    }, [token]);

    if (error) {
        return <div className="alert alert-danger">{error}</div>;
    }

    if (!user) {
        return <div>Loading...</div>;
    }

    return (
        <div>
            <h2>User Profile</h2>
            <div className="card">
                <div className="card-body">
                    <h5 className="card-title">{user.username}</h5>
                    <p className="card-text">Email: {user.email}</p>
                    <p className="card-text">User ID: {user.id}</p>
                </div>
            </div>
            <button onClick={onLogout} className="btn btn-danger mt-3">Logout</button>
        </div>
    );
}

export default Profile;
