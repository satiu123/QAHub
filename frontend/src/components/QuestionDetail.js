import React, { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import axios from 'axios';
import { API_BASE_URL } from '../config/api';

const API_URL = API_BASE_URL;

const formatDisplayName = (name, fallbackId) => {
    if (name && name.trim().length > 0) {
        return name;
    }
    if (fallbackId) {
        return `User ${fallbackId}`;
    }
    return 'Anonymous';
};

const Comments = ({ answerId, token, onCommentAdded }) => {
    const [comments, setComments] = useState([]);
    const [newComment, setNewComment] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);

    const fetchComments = useCallback(async () => {
        if (!answerId) return;
        try {
            const response = await axios.get(`${API_URL}/answers/${answerId}/comments`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            setComments(response.data?.comments || []);
        } catch (err) {
            console.error(`Failed to fetch comments for answer ${answerId}`, err);
        }
    }, [answerId, token]);

    useEffect(() => {
        fetchComments();
    }, [fetchComments]);

    const handleSubmitComment = async (e) => {
        e.preventDefault();
        if (!newComment.trim()) return;

        setIsSubmitting(true);
        try {
            await axios.post(
                `${API_URL}/answers/${answerId}/comments`,
                { content: newComment.trim() },
                { headers: { 'Authorization': `Bearer ${token}` } }
            );

            await fetchComments();
            setNewComment('');

            if (onCommentAdded) {
                onCommentAdded();
            }
        } catch (err) {
            console.error('Failed to submit comment', err);
            alert('Failed to submit comment. Please try again.');
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="mt-3">
            <hr className="my-3" />
            <h6 className="text-muted mb-3">Comments ({comments.length}):</h6>
            <div className="comments-container mb-3">
                {comments.length === 0 ? (
                    <p className="text-muted small mb-3">No comments yet.</p>
                ) : (
                    comments.map(comment => {
                        const displayName = formatDisplayName(comment.username, comment.userId);
                        const createdAt = comment.createdAt;

                        return (
                            // MODIFIED: Added id for scrolling
                            <div key={comment.id} id={`comment-${comment.id}`} className="mb-3">
                                <div className="card border-0 bg-light">
                                    <div className="card-body py-2 px-3">
                                        <p className="card-text mb-2 small" style={{ lineHeight: '1.4' }}>
                                            {comment.content}
                                        </p>
                                        <div className="d-flex flex-column flex-sm-row justify-content-between align-items-start align-items-sm-center">
                                            <small className="text-muted mb-1 mb-sm-0">
                                                {`By ${displayName}`}
                                            </small>
                                            <small className="text-muted">
                                                {createdAt ? new Date(createdAt).toLocaleDateString() : ''}
                                            </small>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        );
                    })
                )}
            </div>

            {/* 添加评论表单 */}
            <form onSubmit={handleSubmitComment} className="mt-3">
                <div className="mb-3">
                    <textarea
                        className="form-control"
                        rows="2"
                        placeholder="Write a comment..."
                        value={newComment}
                        onChange={(e) => setNewComment(e.target.value)}
                        disabled={isSubmitting}
                    />
                </div>
                <div className="d-flex justify-content-end">
                    <button
                        type="submit"
                        className="btn btn-primary btn-sm"
                        disabled={isSubmitting || !newComment.trim()}
                    >
                        {isSubmitting ? 'Submitting...' : 'Add Comment'}
                    </button>
                </div>
            </form>
        </div>
    );
};


function QuestionDetail({ token }) {
    const { questionId } = useParams();
    const [question, setQuestion] = useState(null);
    const [answers, setAnswers] = useState([]);
    const [error, setError] = useState('');
    const [newAnswer, setNewAnswer] = useState('');
    const [isSubmittingAnswer, setIsSubmittingAnswer] = useState(false);

    const fetchQuestion = useCallback(async () => {
        if (!token || !questionId) return;
        try {
            const qResponse = await axios.get(`${API_URL}/questions/${questionId}`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            setQuestion(qResponse.data);
            setError('');
        } catch (err) {
            console.error('Failed to fetch question', err);
            setError('Failed to fetch question details.');
        }
    }, [questionId, token]);

    const fetchAnswers = useCallback(async () => {
        if (!token || !questionId) return;
        try {
            const aResponse = await axios.get(`${API_URL}/questions/${questionId}/answers`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });

            const answersWithVoteState = (aResponse.data?.answers || []).map(ans => ({
                ...ans,
                upvoteCount: ans.upvoteCount ?? 0,
                isUpvotedByUser: ans.isUpvotedByUser ?? false,
                isVoting: false,
            }));
            setAnswers(answersWithVoteState);
            setError('');
        } catch (err) {
            console.error('Failed to fetch answers', err);
            setError('Failed to fetch question details.');
        }
    }, [questionId, token]);

    // 在组件加载时获取问题和答案的详细信息
    useEffect(() => {
        if (!token || !questionId) return;
        fetchQuestion();
        fetchAnswers();
    }, [token, questionId, fetchQuestion, fetchAnswers]);

    // NEW: useEffect for scrolling to a specific answer or comment
    useEffect(() => {
        // Run this after answers have been loaded
        if (answers && answers.length > 0) {
            const hash = window.location.hash; // e.g., #answer-45 or #comment-67
            if (hash) {
                try {
                    const element = document.querySelector(hash);
                    if (element) {
                        // Scroll the element into view
                        element.scrollIntoView({ behavior: 'smooth', block: 'center' });

                        // Add a temporary highlight effect for better UX
                        element.style.transition = 'background-color 0.5s ease';
                        element.style.backgroundColor = '#e7f3ff'; // A light blue highlight
                        setTimeout(() => {
                            element.style.backgroundColor = ''; // Remove highlight after 2.5 seconds
                        }, 2500);
                    }
                } catch (e) {
                    // In case of an invalid selector in the hash
                    console.error("Could not scroll to element with hash:", hash, e);
                }
            }
        }
    }, [answers]); // Dependency on `answers` ensures this runs after data is fetched

    // 处理点赞/取消点赞的函数
    const handleVote = async (answerId, isUpvotedByUser) => {
        // 找到当前正在操作的答案
        const targetAnswer = answers.find(a => a.id === answerId);
        if (targetAnswer.isVoting) return; // 如果正在处理中,则不执行任何操作

        const originalAnswers = [...answers]; // 保存原始状态以便在出错时回滚

        // 1. 乐观更新UI
        setAnswers(answers.map(ans => {
            if (ans.id === answerId) {
                return {
                    ...ans,
                    upvoteCount: isUpvotedByUser ? ans.upvoteCount - 1 : ans.upvoteCount + 1,
                    isUpvotedByUser: !isUpvotedByUser,
                    isVoting: true, // 设置为处理中
                };
            }
            return ans;
        }));

        // 2. 调用API
        const endpoint = isUpvotedByUser ? 'downvote' : 'upvote';
        try {
            await axios.post(`${API_URL}/answers/${answerId}/${endpoint}`, null, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
        } catch (err) {
            console.error(`Failed to ${endpoint} answer`, err);
            alert(`Failed to ${endpoint}. Please try again.`);
            // 3. 如果API调用失败，回滚UI状态
            setAnswers(originalAnswers);
        } finally {
            // 4. 无论成功或失败，都结束处理状态
            setAnswers(prevAnswers => prevAnswers.map(ans => {
                if (ans.id === answerId) {
                    return { ...ans, isVoting: false };
                }
                return ans;
            }));
        }
    };

    // 处理提交新答案的函数
    const handleSubmitAnswer = async (e) => {
        e.preventDefault();
        if (!newAnswer.trim()) return;

        setIsSubmittingAnswer(true);
        try {
            const response = await axios.post(
                `${API_URL}/questions/${questionId}/answers`,
                { content: newAnswer.trim() },
                { headers: { 'Authorization': `Bearer ${token}` } }
            );

            if (response.status >= 200 && response.status < 300) {
                await fetchAnswers();
            }
            setNewAnswer('');
        } catch (err) {
            console.error('Failed to submit answer', err);
            alert('Failed to submit answer. Please try again.');
        } finally {
            setIsSubmittingAnswer(false);
        }
    };

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
                    <h2 className="card-title mb-3">{question.title}</h2>
                    <p className="card-text mb-3" style={{ lineHeight: '1.6' }}>{question.content}</p>
                    <div className="d-flex flex-column flex-sm-row justify-content-between align-items-start align-items-sm-center mt-3">
                        <small className="text-muted mb-1 mb-sm-0">
                            {`Asked by ${formatDisplayName(question.authorName, question.userId)}`}
                        </small>
                        <small className="text-muted">
                            {new Date(question.createdAt).toLocaleDateString()}
                        </small>
                    </div>
                </div>
            </div>

            <h4 className="mb-3">Answers ({answers.length})</h4>
            {answers.length > 0 ? (
                answers.map(answer => (
                    // MODIFIED: Added id for scrolling
                    <div key={answer.id} id={`answer-${answer.id}`} className="card mb-4 shadow-sm">
                        <div className="card-body">
                            <p className="card-text mb-3" style={{ lineHeight: '1.6' }}>{answer.content}</p>
                            <div className="row align-items-center mb-3">
                                <div className="col-sm-8 col-12 mb-2 mb-sm-0">
                                    <div className="d-flex flex-column flex-sm-row justify-content-start align-items-start align-items-sm-center">
                                        <small className="text-muted mb-1 mb-sm-0 me-sm-3">
                                            {`Answered by ${formatDisplayName(answer.username, answer.userId)}`}
                                        </small>
                                        <small className="text-muted">
                                            {new Date(answer.createdAt).toLocaleDateString()}
                                        </small>
                                    </div>
                                </div>
                                <div className="col-sm-4 col-12 text-sm-end text-start">
                                    <button
                                        className={`btn ${answer.isUpvotedByUser ? 'btn-success' : 'btn-outline-success'} fs-6 px-3 py-2`}
                                        onClick={() => handleVote(answer.id, answer.isUpvotedByUser)}
                                        disabled={answer.isVoting}
                                    >
                                        👍 {answer.upvoteCount || 0}
                                    </button>
                                </div>
                            </div>
                            <Comments answerId={answer.id} token={token} />
                        </div>
                    </div>
                ))
            ) : (
                <p>No answers yet.</p>
            )}

            {/* 添加答案表单 */}
            <div className="card mt-4">
                <div className="card-header">
                    <h5 className="mb-0">Write Your Answer</h5>
                </div>
                <div className="card-body">
                    <form onSubmit={handleSubmitAnswer}>
                        <div className="mb-3">
                            <textarea
                                className="form-control"
                                rows="4"
                                placeholder="Write your answer here..."
                                value={newAnswer}
                                onChange={(e) => setNewAnswer(e.target.value)}
                                disabled={isSubmittingAnswer}
                                required
                            />
                        </div>
                        <div className="d-flex justify-content-end">
                            <button
                                type="submit"
                                className="btn btn-primary"
                                disabled={isSubmittingAnswer || !newAnswer.trim()}
                            >
                                {isSubmittingAnswer ? 'Submitting...' : 'Submit Answer'}
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    );
}

export default QuestionDetail;
