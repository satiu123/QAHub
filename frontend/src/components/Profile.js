import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { jwtDecode } from 'jwt-decode';

const API_URL = 'http://localhost:8080/api/v1';

function Profile({ token, onLogout }) {
    const [user, setUser] = useState(null);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchUser = async () => {
            try {
                // 1. 解码 Token 获取用户 ID
                const decodedToken = jwtDecode(token);
                const userId = decodedToken.user_id;

                // 2. 使用用户 ID 请求个人资料
                const response = await axios.get(`${API_URL}/users/${userId}`, {
                    headers: {
                        'Authorization': `Bearer ${token}`
                    }
                });
                setUser(response.data);
            } catch (err) {
                setError('Failed to fetch user data.');
                console.error(err); // 在控制台打印详细错误
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
                    <p className="card-text">Bio: {user.bio}</p>
                    <p className="card-text"><small className="text-muted">User ID: {user.id}</small></p>
                </div>
            </div>
            <button onClick={onLogout} className="btn btn-danger mt-3">Logout</button>
        </div>
    );
}

export default Profile;

