import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Link } from 'react-router-dom';

const API_URL = 'http://localhost:8080/api/v1';

function QuestionList({ token }) {
    const [questions, setQuestions] = useState([]);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchQuestions = async () => {
            try {
                const response = await axios.get(`${API_URL}/questions`, {
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                // 后端返回的数据结构是 { total: number, data: Question[] }
                setQuestions(response.data.data || []);
            } catch (err) {
                setError('Failed to fetch questions.');
                console.error(err);
            }
        };

        if (token) {
            fetchQuestions();
        }
    }, [token]);

    if (error) {
        return <div className="alert alert-danger">{error}</div>;
    }

    return (
        <div>
            <div className="d-flex justify-content-between align-items-center mb-3">
                <h2>Questions</h2>
                <Link to="/create-question" className="btn btn-primary">Ask Question</Link>
            </div>
            <ul className="list-group">
                {questions.length > 0 ? (
                    questions.map(q => (
                        <li key={q.ID} className="list-group-item">
                            <Link to={`/questions/${q.ID}`}>{q.Title}</Link>
                            <p className="text-muted small mb-0">
                                By User {q.UserID} on {new Date(q.CreatedAt).toLocaleDateString()}
                            </p>
                        </li>
                    ))
                ) : (
                    <p>No questions found.</p>
                )}
            </ul>
        </div>
    );
}

export default QuestionList;
