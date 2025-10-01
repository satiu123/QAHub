import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Link } from 'react-router-dom';
import { API_BASE_URL } from '../config/api';

function QuestionList({ token }) {
    const [questions, setQuestions] = useState([]);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchQuestions = async () => {
            try {
                const response = await axios.get(`${API_BASE_URL}/questions`, {
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
                    questions.map(q => {
                        const fallbackUser = q.UserID || q.user_id;
                        const displayName = q.AuthorName || q.author_name || (fallbackUser ? `User ${fallbackUser}` : 'Anonymous');
                        const createdAt = q.CreatedAt || q.created_at;
                        const formattedDate = createdAt ? new Date(createdAt).toLocaleDateString() : '';
                        const answerCount = q.AnswerCount ?? q.answer_count;

                        return (
                            <li key={q.ID || q.id} className="list-group-item">
                                <div className="d-flex justify-content-between align-items-start">
                                    <div className="me-3">
                                        <Link to={`/questions/${q.ID || q.id}`} className="fw-semibold text-decoration-none">
                                            {q.Title || q.title}
                                        </Link>
                                        <p className="text-muted small mb-0">
                                            Asked by {displayName}
                                            {formattedDate && ` · ${formattedDate}`}
                                        </p>
                                    </div>
                                    {typeof answerCount === 'number' && (
                                        <span className="badge bg-secondary align-self-center">
                                            {answerCount} answers
                                        </span>
                                    )}
                                </div>
                            </li>
                        );
                    })
                ) : (
                    <p>No questions found.</p>
                )}
            </ul>
        </div>
    );
}

export default QuestionList;
