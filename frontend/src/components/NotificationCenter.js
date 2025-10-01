import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchNotifications, markNotificationsRead, deleteNotification } from '../services/notificationService';

const normalizeNotification = (raw) => {
    if (!raw) return null;
    const id = raw.id || raw.ID || raw._id;
    const createdAt = raw.created_at || raw.CreatedAt;
    const targetURL = raw.target_url || raw.TargetURL || '';

    if (!id) {
        return null;
    }

    return {
        ...raw,
        id,
        created_at: createdAt,
        target_url: targetURL,
        is_read: raw.is_read ?? raw.IsRead ?? false,
        sender_name: raw.sender_name || raw.SenderName || '',
        type: raw.type || raw.Type || 'notification',
        content: raw.content || raw.Content || '',
    };
};

const sortNotifications = (items) =>
    [...items].sort((a, b) => {
        const dateA = a.created_at ? new Date(a.created_at).getTime() : 0;
        const dateB = b.created_at ? new Date(b.created_at).getTime() : 0;
        return dateB - dateA;
    });

const getNotificationTypeLabel = (type) => {
    const typeMap = {
        new_answer: '新回答',
        new_comment: '新评论',
        new_vote: '新投票',
        system: '系统通知',
    };
    return typeMap[type] || type;
};

const getNotificationTypeColor = (type) => {
    const colorMap = {
        new_answer: 'primary',
        new_comment: 'success',
        new_vote: 'warning',
        system: 'info',
    };
    return colorMap[type] || 'secondary';
};

const formatTime = (dateString) => {
    if (!dateString) return '';

    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return '刚刚';
    if (diffMins < 60) return `${diffMins}分钟前`;
    if (diffHours < 24) return `${diffHours}小时前`;
    if (diffDays < 7) return `${diffDays}天前`;

    return date.toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
    });
};

const NotificationCenter = ({ token }) => {
    const [notifications, setNotifications] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [filter, setFilter] = useState('all'); // all, unread, read
    const navigate = useNavigate();

    const { unreadCount, readCount, filteredNotifications } = useMemo(() => {
        const unread = notifications.filter((n) => !n.is_read);
        const read = notifications.filter((n) => n.is_read);

        let filtered = notifications;
        if (filter === 'unread') {
            filtered = unread;
        } else if (filter === 'read') {
            filtered = read;
        }

        return {
            unreadCount: unread.length,
            readCount: read.length,
            filteredNotifications: filtered,
        };
    }, [notifications, filter]);

    const loadNotifications = useCallback(async () => {
        if (!token) return;

        setLoading(true);
        setError('');
        try {
            const result = await fetchNotifications({ token, limit: 100, offset: 0 });
            const normalized = result
                .map((item) => normalizeNotification(item))
                .filter((item) => item !== null);
            setNotifications(sortNotifications(normalized));
        } catch (err) {
            console.error('Failed to load notifications', err);
            setError('加载通知失败，请稍后重试。');
        } finally {
            setLoading(false);
        }
    }, [token]);

    const markAsReadLocally = useCallback((ids) => {
        if (!ids || ids.length === 0) return;

        setNotifications((prev) =>
            prev.map((notification) =>
                ids.includes(notification.id)
                    ? { ...notification, is_read: true }
                    : notification,
            ),
        );
    }, []);

    const handleNotificationClick = useCallback(
        async (notification) => {
            if (!notification) return;

            const notificationId = notification.id;

            // 标记为已读
            if (!notification.is_read) {
                markAsReadLocally([notificationId]);
                try {
                    await markNotificationsRead({ token, notificationIds: [notificationId] });
                } catch (err) {
                    console.error('Failed to mark notification as read', err);
                    // 如果失败，重新加载通知
                    loadNotifications();
                }
            }

            // 导航到目标页面
            if (notification.target_url) {
                navigate(notification.target_url);
            }
        },
        [token, markAsReadLocally, navigate, loadNotifications],
    );

    const handleMarkAllRead = useCallback(async () => {
        const unreadIds = notifications.filter((item) => !item.is_read).map((item) => item.id);
        if (unreadIds.length === 0) {
            alert('没有未读通知');
            return;
        }

        if (!window.confirm(`确定要将 ${unreadIds.length} 条未读通知标记为已读吗？`)) {
            return;
        }

        markAsReadLocally(unreadIds);
        try {
            await markNotificationsRead({ token, notificationIds: unreadIds });
        } catch (err) {
            console.error('Failed to mark all notifications as read', err);
            setError('标记已读失败，请稍后重试。');
            loadNotifications();
        }
    }, [notifications, markAsReadLocally, token, loadNotifications]);

    const handleDeleteNotification = useCallback(
        async (notificationId, event) => {
            event.stopPropagation(); // 阻止触发点击通知事件

            if (!window.confirm('确定要删除这条通知吗？')) {
                return;
            }

            try {
                await deleteNotification({ token, notificationId });
                setNotifications((prev) => prev.filter((n) => n.id !== notificationId));
            } catch (err) {
                console.error('Failed to delete notification', err);
                setError('删除通知失败，请稍后重试。');
            }
        },
        [token],
    );

    useEffect(() => {
        if (token) {
            loadNotifications();
        }
    }, [loadNotifications, token]);

    if (!token) {
        return (
            <div className="container mt-5">
                <div className="alert alert-warning">请先登录查看通知</div>
            </div>
        );
    }

    return (
        <div className="container mt-4">
            <div className="row">
                <div className="col-lg-8 mx-auto">
                    {/* Header */}
                    <div className="d-flex justify-content-between align-items-center mb-4">
                        <h2 className="mb-0">
                            <span role="img" aria-label="notifications">🔔</span> 通知中心
                        </h2>
                        <button
                            className="btn btn-outline-primary btn-sm"
                            onClick={loadNotifications}
                            disabled={loading}
                        >
                            {loading ? '刷新中...' : '刷新'}
                        </button>
                    </div>

                    {/* Stats and Actions */}
                    <div className="card mb-3">
                        <div className="card-body">
                            <div className="row align-items-center">
                                <div className="col-md-6">
                                    <div className="d-flex gap-3">
                                        <span className="badge bg-primary fs-6">
                                            未读: {unreadCount}
                                        </span>
                                        <span className="badge bg-secondary fs-6">
                                            已读: {readCount}
                                        </span>
                                        <span className="badge bg-light text-dark fs-6">
                                            总计: {notifications.length}
                                        </span>
                                    </div>
                                </div>
                                <div className="col-md-6 text-md-end mt-2 mt-md-0">
                                    <button
                                        className="btn btn-sm btn-success"
                                        onClick={handleMarkAllRead}
                                        disabled={unreadCount === 0}
                                    >
                                        全部标记为已读
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Filter Tabs */}
                    <ul className="nav nav-tabs mb-3">
                        <li className="nav-item">
                            <button
                                className={`nav-link ${filter === 'all' ? 'active' : ''}`}
                                onClick={() => setFilter('all')}
                            >
                                全部 ({notifications.length})
                            </button>
                        </li>
                        <li className="nav-item">
                            <button
                                className={`nav-link ${filter === 'unread' ? 'active' : ''}`}
                                onClick={() => setFilter('unread')}
                            >
                                未读 ({unreadCount})
                            </button>
                        </li>
                        <li className="nav-item">
                            <button
                                className={`nav-link ${filter === 'read' ? 'active' : ''}`}
                                onClick={() => setFilter('read')}
                            >
                                已读 ({readCount})
                            </button>
                        </li>
                    </ul>

                    {/* Error Message */}
                    {error && (
                        <div className="alert alert-danger alert-dismissible fade show" role="alert">
                            {error}
                            <button
                                type="button"
                                className="btn-close"
                                onClick={() => setError('')}
                                aria-label="Close"
                            ></button>
                        </div>
                    )}

                    {/* Loading State */}
                    {loading && (
                        <div className="text-center py-5">
                            <div className="spinner-border text-primary" role="status">
                                <span className="visually-hidden">加载中...</span>
                            </div>
                        </div>
                    )}

                    {/* Notifications List */}
                    {!loading && filteredNotifications.length === 0 && (
                        <div className="text-center py-5 text-muted">
                            <h4>暂无通知</h4>
                            <p>
                                {filter === 'unread' && '没有未读通知'}
                                {filter === 'read' && '没有已读通知'}
                                {filter === 'all' && '暂时还没有任何通知'}
                            </p>
                        </div>
                    )}

                    {!loading && filteredNotifications.length > 0 && (
                        <div className="list-group">
                            {filteredNotifications.map((notification) => (
                                <div
                                    key={notification.id}
                                    className={`list-group-item list-group-item-action ${!notification.is_read ? 'list-group-item-primary' : ''
                                        }`}
                                    style={{ cursor: 'pointer' }}
                                >
                                    <div
                                        onClick={() => handleNotificationClick(notification)}
                                        className="d-flex w-100 justify-content-between"
                                    >
                                        <div className="flex-grow-1">
                                            <div className="d-flex justify-content-between align-items-start mb-2">
                                                <h6 className="mb-1">
                                                    {!notification.is_read && (
                                                        <span className="badge bg-danger me-2">未读</span>
                                                    )}
                                                    <span className={`badge bg-${getNotificationTypeColor(notification.type)} me-2`}>
                                                        {getNotificationTypeLabel(notification.type)}
                                                    </span>
                                                    {notification.sender_name || '系统通知'}
                                                </h6>
                                            </div>
                                            <p className="mb-2">{notification.content || '您有一条新通知'}</p>
                                            <small className="text-muted">
                                                {formatTime(notification.created_at)}
                                            </small>
                                        </div>
                                        <div className="ms-3">
                                            <button
                                                className="btn btn-sm btn-outline-danger"
                                                onClick={(e) => handleDeleteNotification(notification.id, e)}
                                                title="删除通知"
                                            >
                                                <span role="img" aria-label="delete">🗑️</span>
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default NotificationCenter;
