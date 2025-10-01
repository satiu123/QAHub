import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { jwtDecode } from 'jwt-decode';
import { Link } from 'react-router-dom';
import { API_BASE_URL } from '../config/api';

const API_URL = API_BASE_URL;

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

    // 用户信息编辑相关状态
    const [isEditingProfile, setIsEditingProfile] = useState(false);
    const [editingProfile, setEditingProfile] = useState({
        username: '',
        email: '',
        bio: ''
    });
    const [isUpdatingProfile, setIsUpdatingProfile] = useState(false);
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

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

    // 开始编辑用户资料
    const startEditProfile = () => {
        setEditingProfile({
            username: user.username,
            email: user.email,
            bio: user.bio
        });
        setIsEditingProfile(true);
    };

    // 取消编辑用户资料
    const cancelEditProfile = () => {
        setIsEditingProfile(false);
        setEditingProfile({
            username: '',
            email: '',
            bio: ''
        });
    };

    // 更新用户资料
    const handleUpdateProfile = async (e) => {
        e.preventDefault();
        if (!editingProfile.username.trim() || !editingProfile.email.trim()) {
            alert('用户名和邮箱不能为空');
            return;
        }

        setIsUpdatingProfile(true);
        try {
            const decodedToken = jwtDecode(token);
            const userId = decodedToken.user_id;

            // 构建只包含修改字段的更新数据
            const updateData = {};
            if (editingProfile.username !== user.username) {
                updateData.username = editingProfile.username;
            }
            if (editingProfile.email !== user.email) {
                updateData.email = editingProfile.email;
            }
            if (editingProfile.bio !== user.bio) {
                updateData.bio = editingProfile.bio;
            }

            // 如果没有任何变化，直接取消编辑
            if (Object.keys(updateData).length === 0) {
                cancelEditProfile();
                return;
            }

            await axios.put(
                `${API_URL}/users/${userId}`,
                updateData,
                {
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json'
                    }
                }
            );

            // 更新本地用户数据
            setUser({ ...user, ...editingProfile });
            setIsEditingProfile(false);
            alert('用户资料更新成功！');
        } catch (err) {
            console.error('Failed to update profile:', err);
            if (err.response?.data?.error) {
                alert(`更新失败: ${err.response.data.error}`);
            } else {
                alert('更新用户资料失败，请重试。');
            }
        } finally {
            setIsUpdatingProfile(false);
        }
    };

    // 注销账号
    const handleDeleteAccount = async () => {
        try {
            const decodedToken = jwtDecode(token);
            const userId = decodedToken.user_id;

            await axios.delete(`${API_URL}/users/${userId}`, {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });

            alert('账号已成功注销');
            onLogout(); // 登出用户
        } catch (err) {
            console.error('Failed to delete account:', err);
            if (err.response?.data?.error) {
                alert(`注销失败: ${err.response.data.error}`);
            } else {
                alert('注销账号失败，请重试。');
            }
        } finally {
            setShowDeleteConfirm(false);
        }
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
                <li className="nav-item">
                    <button
                        className={`nav-link ${activeTab === 'settings' ? 'active' : ''}`}
                        onClick={() => handleTabChange('settings')}
                    >
                        账号设置
                    </button>
                </li>
            </ul>

            {/* Tab Content */}
            {activeTab === 'profile' && (
                <div>
                    {!isEditingProfile ? (
                        <div className="card">
                            <div className="card-body">
                                <div className="d-flex justify-content-between align-items-start">
                                    <div>
                                        <h5 className="card-title">{user.username}</h5>
                                        <p className="card-text">Email: {user.email}</p>
                                        <p className="card-text">Bio: {user.bio || '暂无个人简介'}</p>
                                        <p className="card-text"><small className="text-muted">User ID: {user.id}</small></p>
                                    </div>
                                    <button
                                        className="btn btn-outline-primary"
                                        onClick={startEditProfile}
                                    >
                                        编辑资料
                                    </button>
                                </div>
                            </div>
                        </div>
                    ) : (
                        <div className="card">
                            <div className="card-body">
                                <h5 className="card-title mb-3">编辑个人资料</h5>
                                <form onSubmit={handleUpdateProfile}>
                                    <div className="mb-3">
                                        <label htmlFor="username" className="form-label">用户名</label>
                                        <input
                                            type="text"
                                            className="form-control"
                                            id="username"
                                            value={editingProfile.username}
                                            onChange={(e) => setEditingProfile({
                                                ...editingProfile,
                                                username: e.target.value
                                            })}
                                            required
                                            disabled={isUpdatingProfile}
                                        />
                                    </div>
                                    <div className="mb-3">
                                        <label htmlFor="email" className="form-label">邮箱</label>
                                        <input
                                            type="email"
                                            className="form-control"
                                            id="email"
                                            value={editingProfile.email}
                                            onChange={(e) => setEditingProfile({
                                                ...editingProfile,
                                                email: e.target.value
                                            })}
                                            required
                                            disabled={isUpdatingProfile}
                                        />
                                    </div>
                                    <div className="mb-3">
                                        <label htmlFor="bio" className="form-label">个人简介</label>
                                        <textarea
                                            className="form-control"
                                            id="bio"
                                            rows="3"
                                            value={editingProfile.bio}
                                            onChange={(e) => setEditingProfile({
                                                ...editingProfile,
                                                bio: e.target.value
                                            })}
                                            disabled={isUpdatingProfile}
                                            placeholder="介绍一下自己..."
                                        ></textarea>
                                    </div>
                                    <div className="d-flex gap-2">
                                        <button
                                            type="submit"
                                            className="btn btn-primary"
                                            disabled={isUpdatingProfile}
                                        >
                                            {isUpdatingProfile ? '保存中...' : '保存更改'}
                                        </button>
                                        <button
                                            type="button"
                                            className="btn btn-secondary"
                                            onClick={cancelEditProfile}
                                            disabled={isUpdatingProfile}
                                        >
                                            取消
                                        </button>
                                    </div>
                                </form>
                            </div>
                        </div>
                    )}
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

            {activeTab === 'settings' && (
                <div>
                    <div className="card">
                        <div className="card-body">
                            <h5 className="card-title">账号设置</h5>

                            <div className="mb-4">
                                <h6>安全操作</h6>
                                <div className="d-flex gap-2">
                                    <button
                                        onClick={onLogout}
                                        className="btn btn-outline-secondary"
                                    >
                                        退出登录
                                    </button>
                                    <button
                                        onClick={() => setShowDeleteConfirm(true)}
                                        className="btn btn-danger"
                                    >
                                        注销账号
                                    </button>
                                </div>
                            </div>

                            <div className="alert alert-warning">
                                <strong>注意：</strong>注销账号将永久删除您的所有数据，包括问题、回答和个人信息。此操作不可撤销。
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {/* 注销账号确认对话框 */}
            {showDeleteConfirm && (
                <div className="modal fade show d-block" style={{ backgroundColor: 'rgba(0,0,0,0.5)' }}>
                    <div className="modal-dialog">
                        <div className="modal-content">
                            <div className="modal-header">
                                <h5 className="modal-title text-danger">确认注销账号</h5>
                                <button
                                    type="button"
                                    className="btn-close"
                                    onClick={() => setShowDeleteConfirm(false)}
                                ></button>
                            </div>
                            <div className="modal-body">
                                <div className="alert alert-danger">
                                    <i className="fas fa-exclamation-triangle me-2"></i>
                                    <strong>警告：此操作不可撤销！</strong>
                                </div>
                                <p>您确定要注销账号吗？这将：</p>
                                <ul>
                                    <li>永久删除您的个人资料</li>
                                    <li>删除您发布的所有问题</li>
                                    <li>删除您的所有回答和评论</li>
                                    <li>清除所有相关数据</li>
                                </ul>
                                <p className="text-muted">如果您只是想暂时停用账号，建议选择"退出登录"。</p>
                            </div>
                            <div className="modal-footer">
                                <button
                                    type="button"
                                    className="btn btn-secondary"
                                    onClick={() => setShowDeleteConfirm(false)}
                                >
                                    取消
                                </button>
                                <button
                                    type="button"
                                    className="btn btn-danger"
                                    onClick={handleDeleteAccount}
                                >
                                    确认注销账号
                                </button>
                            </div>
                        </div>
                    </div>
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

