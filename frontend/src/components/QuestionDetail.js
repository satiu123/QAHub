import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import axios from 'axios';

const API_URL = 'http://localhost:8080/api/v1';

const Comments = ({ answerId, token, onCommentAdded }) => {
    const [comments, setComments] = useState([]);
    const [newComment, setNewComment] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);

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

    const handleSubmitComment = async (e) => {
        e.preventDefault();
        if (!newComment.trim()) return;

        setIsSubmitting(true);
        try {
            const response = await axios.post(
                `${API_URL}/answers/${answerId}/comments`,
                { content: newComment.trim() },
                { headers: { 'Authorization': `Bearer ${token}` } }
            );

            // 添加新评论到列表
            const addedComment = response.data;
            setComments([...comments, addedComment]);
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

    if (comments.length === 0) {
        return (
            <div className="mt-3">
                <hr className="my-3" />
                <h6 className="text-muted mb-3">Comments (0):</h6>
                <p className="text-muted small mb-3">No comments yet.</p>

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
    }

    return (
        <div className="mt-3">
            <hr className="my-3" />
            <h6 className="text-muted mb-3">Comments ({comments.length}):</h6>
            <div className="comments-container mb-3">
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

    // 在组件加载时获取问题和答案的详细信息
    useEffect(() => {
        const fetchDetails = async () => {
            try {
                // 获取问题
                const qResponse = await axios.get(`${API_URL}/questions/${questionId}`, {
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                setQuestion(qResponse.data);

                // 获取答案
                const aResponse = await axios.get(`${API_URL}/questions/${questionId}/answers`, {
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                
                // 为每个答案添加本地的点赞状态，以便UI可以响应
                // 注意：这个状态只在当前页面有效
                const answersWithVoteState = (aResponse.data?.data || aResponse.data?.answers || []).map(ans => ({
                    ...ans,
                    isUpvoted: false, // 初始状态为未点赞
                    isVoting: false,  // 用于防止重复点击
                }));
                setAnswers(answersWithVoteState);

            } catch (err) {
                setError('Failed to fetch question details.');
                console.error(err);
            }
        };

        if (token && questionId) {
            fetchDetails();
        }
    }, [token, questionId]);

    // 处理点赞/取消点赞的函数
    const handleVote = async (answerId, isUpvoted) => {
        // 找到当前正在操作的答案
        const targetAnswer = answers.find(a => (a.ID || a.id) === answerId);
        if (targetAnswer.isVoting) return; // 如果正在处理中，则不执行任何操作

        const originalAnswers = [...answers]; // 保存原始状态以便在出错时回滚

        // 1. 乐观更新UI
        setAnswers(answers.map(ans => {
            if ((ans.ID || ans.id) === answerId) {
                return {
                    ...ans,
                    UpvoteCount: isUpvoted ? ans.UpvoteCount - 1 : ans.UpvoteCount + 1,
                    isUpvoted: !isUpvoted,
                    isVoting: true, // 设置为处理中
                };
            }
            return ans;
        }));

        // 2. 调用API
        const endpoint = isUpvoted ? 'downvote' : 'upvote';
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
                if ((ans.ID || ans.id) === answerId) {
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

            // 添加新答案到列表，并附加上本地状态
            const addedAnswer = { ...response.data, isUpvoted: false, isVoting: false };
            setAnswers([...answers, addedAnswer]);
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
                                    <button 
                                        className={`btn ${answer.isUpvoted ? 'btn-success' : 'btn-outline-success'} fs-6 px-3 py-2`}
                                        onClick={() => handleVote(answer.ID || answer.id, answer.isUpvoted)}
                                        disabled={answer.isVoting}
                                    >
                                        👍 {answer.UpvoteCount || answer.upvote_count || 0}
                                    </button>
                                </div>
                            </div>
                            <Comments answerId={answer.ID || answer.id} token={token} />
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
