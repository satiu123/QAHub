import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import axios from 'axios';

const API_URL = 'http://localhost:8080/api/v1';

const Comments = ({ answerId, token }) => {
    const [comments, setComments] = useState([]);

    useEffect(() => {
        const fetchComments = async () => {
            if (!answerId) return;
            try {
                const response = await axios.get(`${API_URL}/answers/${answerId}/comments`, {
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                // 根据实际API响应结构调整
                setComments(response.data?.data || response.data?.comments || []);
            } catch (err) {
                console.error(`Failed to fetch comments for answer ${answerId}`, err);
            }
        };
        fetchComments();
    }, [answerId, token]);

    if (comments.length === 0) {
        return (
            <div className="mt-3">
                <hr className="my-3" />
                <p className="text-muted small mb-0">No comments yet.</p>
            </div>
        );
    }

    return (
        <div className="mt-3">
            <hr className="my-3" />
            <h6 className="text-muted mb-3">Comments:</h6>
            <div className="comments-container">
                {comments.map(comment => (
                    <div key={comment.ID || comment.id} className="mb-3">
                        <div className="card border-0 bg-light">
                            <div className="card-body py-2 px-3">
                                <p className="card-text mb-2 small" style={{ lineHeight: '1.4' }}>
                                    {comment.Content || comment.content}
                                </p>
                                <div className="d-flex flex-column flex-sm-row justify-content-between align-items-start align-items-sm-center">
                                    <small className="text-muted mb-1 mb-sm-0">
                                        User ID: {comment.UserID || comment.user_id}
                                    </small>
                                    <small className="text-muted">
                                        {new Date(comment.CreatedAt || comment.created_at).toLocaleDateString()}
                                    </small>
                                </div>
                            </div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};


function QuestionDetail({ token }) {
    const { questionId } = useParams();
    const [question, setQuestion] = useState(null);
    const [answers, setAnswers] = useState([]);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchDetails = async () => {
            try {
                // Fetch Question
                const qResponse = await axios.get(`${API_URL}/questions/${questionId}`, {
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                setQuestion(qResponse.data);

                // Fetch Answers
                const aResponse = await axios.get(`${API_URL}/questions/${questionId}/answers`, {
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                // 根据您提供的数据结构，答案在 data 字段中
                setAnswers(aResponse.data?.data || aResponse.data?.answers || []);

            } catch (err) {
                setError('Failed to fetch question details.');
                console.error(err);
            }
        };

        if (token && questionId) {
            fetchDetails();
        }
    }, [token, questionId]);

    if (error) {
        return <div className="alert alert-danger">{error}</div>;
    }

    if (!question) {
        return <div>Loading...</div>;
    }

    return (
        <div className="container-fluid">
            <div className="card mb-4">
                <div className="card-body">
                    <h2 className="card-title mb-3">{question.Title || question.title}</h2>
                    <p className="card-text mb-3" style={{ lineHeight: '1.6' }}>{question.Content || question.content}</p>
                    <div className="d-flex flex-column flex-sm-row justify-content-between align-items-start align-items-sm-center mt-3">
                        <small className="text-muted mb-1 mb-sm-0">
                            User ID: {question.UserID || question.user_id}
                        </small>
                        <small className="text-muted">
                            {new Date(question.CreatedAt || question.created_at).toLocaleDateString()}
                        </small>
                    </div>
                </div>
            </div>

            <h4 className="mb-3">Answers ({answers.length})</h4>
            {answers.length > 0 ? (
                answers.map(answer => (
                    <div key={answer.ID || answer.id} className="card mb-4 shadow-sm">
                        <div className="card-body">
                            <p className="card-text mb-3" style={{ lineHeight: '1.6' }}>{answer.Content || answer.content}</p>
                            <div className="row align-items-center mb-3">
                                <div className="col-sm-8 col-12 mb-2 mb-sm-0">
                                    <div className="d-flex flex-column flex-sm-row justify-content-start align-items-start align-items-sm-center">
                                        <small className="text-muted mb-1 mb-sm-0 me-sm-3">
                                            User ID: {answer.UserID || answer.user_id}
                                        </small>
                                        <small className="text-muted">
                                            {new Date(answer.CreatedAt || answer.created_at).toLocaleDateString()}
                                        </small>
                                    </div>
                                </div>
                                <div className="col-sm-4 col-12 text-sm-end text-start">
                                    <span className="badge bg-success fs-6 px-3 py-2">
                                        👍 {answer.UpvoteCount || answer.upvote_count || 0}
                                    </span>
                                </div>
                            </div>
                            <Comments answerId={answer.ID || answer.id} token={token} />
                        </div>
                    </div>
                ))
            ) : (
                <p>No answers yet.</p>
            )}
        </div>
    );
}

export default QuestionDetail;
