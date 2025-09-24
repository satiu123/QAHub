import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { jwtDecode } from 'jwt-decode';
import { Link } from 'react-router-dom';

const API_URL = 'http://localhost:8080/api/v1';

function Profile({ token, onLogout }) {
    const [user, setUser] = useState(null);
    const [myQuestions, setMyQuestions] = useState([]);
    const [activeTab, setActiveTab] = useState('profile');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(true);
    const [questionsLoading, setQuestionsLoading] = useState(false);
    const [editingQuestion, setEditingQuestion] = useState(null);
    const [editTitle, setEditTitle] = useState('');
    const [editContent, setEditContent] = useState('');
    const [isUpdating, setIsUpdating] = useState(false);

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
            } finally {
                setLoading(false);
            }
        };

        if (token) {
            fetchUser();
        }
    }, [token]);

    const fetchMyQuestions = async () => {
        if (!user) return;

        setQuestionsLoading(true);
        try {
            const response = await axios.get(`${API_URL}/questions?author=me`, {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
            setMyQuestions(response.data?.data || []);
        } catch (err) {
            console.error('Failed to fetch user questions:', err);
            setError('Failed to fetch your questions.');
        } finally {
            setQuestionsLoading(false);
        }
    };

    const handleDeleteQuestion = async (questionId) => {
        if (!window.confirm('确定要删除这个问题吗？')) {
            return;
        }

        try {
            await axios.delete(`${API_URL}/questions/${questionId}`, {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
            // 从列表中移除已删除的问题
            setMyQuestions(myQuestions.filter(q => q.ID !== questionId));
        } catch (err) {
            console.error('Failed to delete question:', err);
            alert('删除问题失败，请重试。');
        }
    };

    const handleEditQuestion = (question) => {
        setEditingQuestion(question);
        setEditTitle(question.Title);
        setEditContent(question.Content);
    };

    const handleUpdateQuestion = async (e) => {
        e.preventDefault();
        if (!editTitle.trim() || !editContent.trim()) {
            alert('标题和内容不能为空');
            return;
        }

        setIsUpdating(true);
        try {
            const response = await axios.put(
                `${API_URL}/questions/${editingQuestion.ID}`,
                {
                    title: editTitle,
                    content: editContent
                },
                {
                    headers: {
                        'Authorization': `Bearer ${token}`
                    }
                }
            );

            // 更新本地列表中的问题
            setMyQuestions(myQuestions.map(q =>
                q.ID === editingQuestion.ID ? response.data : q
            ));

            // 关闭编辑模态框
            setEditingQuestion(null);
            setEditTitle('');
            setEditContent('');
        } catch (err) {
            console.error('Failed to update question:', err);
            alert('更新问题失败，请重试。');
        } finally {
            setIsUpdating(false);
        }
    };

    const cancelEdit = () => {
        setEditingQuestion(null);
        setEditTitle('');
        setEditContent('');
    };

    // Tab change handler
    const handleTabChange = (tab) => {
        setActiveTab(tab);
        if (tab === 'questions' && myQuestions.length === 0) {
            fetchMyQuestions();
        }
    };

    if (error) {
        return <div className="alert alert-danger">{error}</div>;
    }

    if (loading) {
        return <div>Loading...</div>;
    }

    return (
        <div>
            <h2>User Profile</h2>

            {/* Tab Navigation */}
            <ul className="nav nav-tabs mb-4">
                <li className="nav-item">
                    <button
                        className={`nav-link ${activeTab === 'profile' ? 'active' : ''}`}
                        onClick={() => handleTabChange('profile')}
                    >
                        个人信息
                    </button>
                </li>
                <li className="nav-item">
                    <button
                        className={`nav-link ${activeTab === 'questions' ? 'active' : ''}`}
                        onClick={() => handleTabChange('questions')}
                    >
                        我的问题
                    </button>
                </li>
            </ul>

            {/* Tab Content */}
            {activeTab === 'profile' && (
                <div>
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
            )}

            {activeTab === 'questions' && (
                <div>
                    <div className="d-flex justify-content-between align-items-center mb-3">
                        <h4>我的问题</h4>
                        <Link to="/create-question" className="btn btn-primary">
                            提问
                        </Link>
                    </div>

                    {questionsLoading ? (
                        <div>Loading questions...</div>
                    ) : (
                        <div>
                            {myQuestions.length > 0 ? (
                                <div className="list-group">
                                    {myQuestions.map(question => (
                                        <div key={question.ID} className="list-group-item">
                                            <div className="d-flex justify-content-between align-items-start">
                                                <div className="flex-grow-1">
                                                    <h6 className="mb-1">
                                                        <Link to={`/questions/${question.ID}`}>
                                                            {question.Title}
                                                        </Link>
                                                    </h6>
                                                    <p className="mb-1 text-muted small">
                                                        {question.Content?.substring(0, 100)}
                                                        {question.Content?.length > 100 ? '...' : ''}
                                                    </p>
                                                    <small className="text-muted">
                                                        创建于: {new Date(question.CreatedAt).toLocaleDateString()}
                                                    </small>
                                                </div>
                                                <div className="ms-3">
                                                    <button
                                                        className="btn btn-sm btn-outline-primary me-2"
                                                        onClick={() => handleEditQuestion(question)}
                                                    >
                                                        编辑
                                                    </button>
                                                    <button
                                                        className="btn btn-sm btn-outline-danger"
                                                        onClick={() => handleDeleteQuestion(question.ID)}
                                                    >
                                                        删除
                                                    </button>
                                                </div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            ) : (
                                <div className="text-center py-4">
                                    <p className="text-muted">您还没有提过问题</p>
                                    <Link to="/create-question" className="btn btn-primary">
                                        提第一个问题
                                    </Link>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            )}

            {/* Edit Question Modal */}
            {editingQuestion && (
                <div className="modal fade show d-block" style={{ backgroundColor: 'rgba(0,0,0,0.5)' }}>
                    <div className="modal-dialog modal-lg">
                        <div className="modal-content">
                            <div className="modal-header">
                                <h5 className="modal-title">编辑问题</h5>
                                <button
                                    type="button"
                                    className="btn-close"
                                    onClick={cancelEdit}
                                    disabled={isUpdating}
                                ></button>
                            </div>
                            <form onSubmit={handleUpdateQuestion}>
                                <div className="modal-body">
                                    <div className="mb-3">
                                        <label htmlFor="editTitle" className="form-label">标题</label>
                                        <input
                                            type="text"
                                            className="form-control"
                                            id="editTitle"
                                            value={editTitle}
                                            onChange={(e) => setEditTitle(e.target.value)}
                                            required
                                            disabled={isUpdating}
                                        />
                                    </div>
                                    <div className="mb-3">
                                        <label htmlFor="editContent" className="form-label">内容</label>
                                        <textarea
                                            className="form-control"
                                            id="editContent"
                                            rows="8"
                                            value={editContent}
                                            onChange={(e) => setEditContent(e.target.value)}
                                            required
                                            disabled={isUpdating}
                                        ></textarea>
                                    </div>
                                </div>
                                <div className="modal-footer">
                                    <button
                                        type="button"
                                        className="btn btn-secondary"
                                        onClick={cancelEdit}
                                        disabled={isUpdating}
                                    >
                                        取消
                                    </button>
                                    <button
                                        type="submit"
                                        className="btn btn-primary"
                                        disabled={isUpdating}
                                    >
                                        {isUpdating ? '更新中...' : '保存更改'}
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}

export default Profile;

