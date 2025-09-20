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
                setComments(response.data.comments || []);
            } catch (err) {
                console.error(`Failed to fetch comments for answer ${answerId}`, err);
            }
        };
        fetchComments();
    }, [answerId, token]);

    if (comments.length === 0) {
        return <p className="text-muted small mt-2 ms-4">No comments yet.</p>;
    }

    return (
        <div className="mt-2 ms-4">
            <h6 className="small">Comments:</h6>
            {comments.map(comment => (
                <div key={comment.id} className="card bg-light mt-2">
                    <div className="card-body py-2 px-3">
                        <p className="card-text mb-1">{comment.content}</p>
                        <footer className="blockquote-footer small">{comment.author_username}</footer>
                    </div>
                </div>
            ))}
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
                setAnswers(aResponse.data.answers || []);

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
        <div>
            <div className="card mb-4">
                <div className="card-body">
                    <h2 className="card-title">{question.title}</h2>
                    <p className="card-text">{question.content}</p>
                    <footer className="blockquote-footer">
                        Asked by {question.author_username} on {new Date(question.created_at).toLocaleDateString()}
                    </footer>
                </div>
            </div>

            <h4>Answers</h4>
            {answers.length > 0 ? (
                answers.map(answer => (
                    <div key={answer.id} className="card mb-3">
                        <div className="card-body">
                            <p className="card-text">{answer.content}</p>
                            <footer className="blockquote-footer">
                                Answered by {answer.author_username}
                            </footer>
                            <hr />
                            <Comments answerId={answer.id} token={token} />
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
