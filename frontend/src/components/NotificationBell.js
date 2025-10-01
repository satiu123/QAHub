import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchNotifications, markNotificationsRead } from '../services/notificationService';
import { WS_NOTIFICATIONS_URL } from '../config/api';

const MAX_NOTIFICATIONS = 50;

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

const wsStatusStyles = {
    connected: 'text-success',
    connecting: 'text-primary',
    error: 'text-danger',
    disconnected: 'text-warning',
    idle: 'text-muted',
};

const wsStatusLabels = {
    connected: '已连接',
    connecting: '连接中',
    error: '连接异常',
    disconnected: '已断开',
    idle: '未连接',
};

const NotificationBell = ({ token }) => {
    const [notifications, setNotifications] = useState([]);
    const [isOpen, setIsOpen] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [wsStatus, setWsStatus] = useState('idle');
    const wsRef = useRef(null);
    const reconnectTimerRef = useRef(null);
    const containerRef = useRef(null);
    const navigate = useNavigate();

    const unreadCount = useMemo(
        () => notifications.filter((notification) => !notification.is_read).length,
        [notifications],
    );

    const upsertNotification = useCallback((incoming) => {
        const normalized = normalizeNotification(incoming);
        if (!normalized) {
            return;
        }

        setNotifications((prev) => {
            const withoutCurrent = prev.filter((item) => item.id !== normalized.id);
            const next = sortNotifications([normalized, ...withoutCurrent]).slice(0, MAX_NOTIFICATIONS);
            return next;
        });
    }, []);

    const handleFetchNotifications = useCallback(async () => {
        if (!token) {
            return;
        }
        setLoading(true);
        setError('');
        try {
            const result = await fetchNotifications({ token });
            const normalized = result
                .map((item) => normalizeNotification(item))
                .filter((item) => item !== null);
            setNotifications(sortNotifications(normalized).slice(0, MAX_NOTIFICATIONS));
        } catch (err) {
            console.error('Failed to load notifications', err);
            setError('加载通知失败，请稍后重试。');
        } finally {
            setLoading(false);
        }
    }, [token]);

    const closeDropdown = useCallback(() => setIsOpen(false), []);

    const markAsReadLocally = useCallback((ids) => {
        if (!ids || ids.length === 0) {
            return;
        }
        setNotifications((prev) =>
            prev.map((notification) =>
                ids.includes(notification.id)
                    ? { ...notification, is_read: true }
                    : notification,
            ),
        );
    }, []);

    const handleNotificationClick = useCallback(
        async (notification, event) => {
            if (!notification) return;

            // 阻止事件冒泡
            if (event) {
                event.stopPropagation();
            }

            const notificationId = notification.id;

            // 先标记为已读
            if (!notification.is_read) {
                markAsReadLocally([notificationId]);
                try {
                    await markNotificationsRead({ token, notificationIds: [notificationId] });
                } catch (err) {
                    console.error('Failed to mark notification as read', err);
                    // 如果失败，重新加载通知列表
                    handleFetchNotifications();
                }
            }

            closeDropdown();

            // 导航到目标页面
            if (notification.target_url) {
                navigate(notification.target_url);
            }
        },
        [markAsReadLocally, navigate, closeDropdown, token, handleFetchNotifications],
    );

    const handleMarkAllRead = useCallback(async (event) => {
        if (event) {
            event.stopPropagation();
        }

        const unreadIds = notifications.filter((item) => !item.is_read).map((item) => item.id);
        if (unreadIds.length === 0) {
            return;
        }

        markAsReadLocally(unreadIds);
        try {
            await markNotificationsRead({ token, notificationIds: unreadIds });
        } catch (err) {
            console.error('Failed to mark all notifications as read', err);
            // 如果失败，重新加载通知列表
            handleFetchNotifications();
        }
    }, [notifications, markAsReadLocally, token, handleFetchNotifications]);

    const toggleDropdown = useCallback(() => {
        setIsOpen((prev) => !prev);
    }, []);

    // Fetch initial notifications when token changes
    useEffect(() => {
        if (!token) {
            setNotifications([]);
            setError('');
            setWsStatus('idle');
            return;
        }

        handleFetchNotifications();
    }, [handleFetchNotifications, token]);

    // Handle WebSocket connection lifecycle
    useEffect(() => {
        if (!token) {
            if (wsRef.current) {
                wsRef.current.close();
                wsRef.current = null;
            }
            clearTimeout(reconnectTimerRef.current);
            setWsStatus('idle');
            return;
        }

        let shouldReconnect = true;
        let reconnectDelay = 2000;

        const connect = () => {
            setWsStatus('connecting');
            try {
                const url = new URL(WS_NOTIFICATIONS_URL);
                url.searchParams.set('token', token);
                const ws = new WebSocket(url.toString());
                wsRef.current = ws;

                ws.onopen = () => {
                    reconnectDelay = 2000;
                    setWsStatus('connected');
                };

                ws.onmessage = (event) => {
                    try {
                        const payload = JSON.parse(event.data);
                        upsertNotification(payload);
                    } catch (err) {
                        console.error('Failed to parse notification payload', err);
                    }
                };

                ws.onerror = (event) => {
                    console.error('Notification WebSocket encountered an error', event);
                    setWsStatus('error');
                };

                ws.onclose = () => {
                    if (shouldReconnect) {
                        setWsStatus('disconnected');
                        clearTimeout(reconnectTimerRef.current);
                        reconnectTimerRef.current = setTimeout(() => {
                            reconnectDelay = Math.min(reconnectDelay * 1.5, 15000);
                            connect();
                        }, reconnectDelay);
                    }
                };
            } catch (err) {
                console.error('Failed to initiate WebSocket connection', err);
                setWsStatus('error');
            }
        };

        connect();

        return () => {
            shouldReconnect = false;
            clearTimeout(reconnectTimerRef.current);
            if (wsRef.current) {
                wsRef.current.close();
                wsRef.current = null;
            }
        };
    }, [token, upsertNotification]);

    // Close dropdown on outside click
    useEffect(() => {
        if (!isOpen) return;

        const handleOutsideClick = (event) => {
            if (containerRef.current && !containerRef.current.contains(event.target)) {
                setIsOpen(false);
            }
        };

        const handleEscape = (event) => {
            if (event.key === 'Escape') {
                setIsOpen(false);
            }
        };

        document.addEventListener('mousedown', handleOutsideClick);
        window.addEventListener('keydown', handleEscape);

        return () => {
            document.removeEventListener('mousedown', handleOutsideClick);
            window.removeEventListener('keydown', handleEscape);
        };
    }, [isOpen]);

    // 移除自动标记已读功能，只在用户主动点击时标记

    if (!token) {
        return null;
    }

    const statusStyle = wsStatusStyles[wsStatus] || wsStatusStyles.idle;
    const statusLabel = wsStatusLabels[wsStatus] || wsStatusLabels.idle;

    return (
        <div className="position-relative" ref={containerRef}>
            <button
                type="button"
                className="btn btn-outline-secondary position-relative"
                onClick={toggleDropdown}
                title="通知"
            >
                <span role="img" aria-label="notifications">🔔</span>
                {unreadCount > 0 && (
                    <span className="position-absolute top-0 start-100 translate-middle badge rounded-pill bg-danger">
                        {unreadCount > 99 ? '99+' : unreadCount}
                        <span className="visually-hidden">未读通知</span>
                    </span>
                )}
            </button>

            {isOpen && (
                <div className="dropdown-menu dropdown-menu-end show p-0 shadow notification-dropdown">
                    <div className="d-flex justify-content-between align-items-center px-3 py-2 border-bottom">
                        <span className="fw-semibold">通知</span>
                        <div className="d-flex align-items-center gap-2">
                            <span className={`small d-flex align-items-center gap-1 ${statusStyle}`}>
                                <span className="badge rounded-pill bg-light border">状态</span>
                                {statusLabel}
                            </span>
                            {unreadCount > 0 && (
                                <button
                                    type="button"
                                    className="btn btn-sm btn-link text-decoration-none p-0"
                                    onClick={handleMarkAllRead}
                                    title="全部标记为已读"
                                >
                                    全部已读
                                </button>
                            )}
                        </div>
                    </div>

                    <div className="notification-scroll">
                        {loading && (
                            <div className="px-3 py-3 small text-muted">加载中...</div>
                        )}
                        {!loading && error && (
                            <div className="px-3 py-3 text-danger small">{error}</div>
                        )}
                        {!loading && !error && notifications.length === 0 && (
                            <div className="px-3 py-4 text-center text-muted small">暂无通知</div>
                        )}
                        {!loading && !error && notifications.length > 0 && (
                            <ul className="list-unstyled mb-0">
                                {notifications.map((notification) => {
                                    const createdLabel = notification.created_at
                                        ? new Date(notification.created_at).toLocaleString()
                                        : '';
                                    const isUnread = !notification.is_read;
                                    return (
                                        <li
                                            key={notification.id}
                                            className={`notification-entry ${isUnread ? 'notification-entry-unread' : ''}`}
                                        >
                                            <button
                                                type="button"
                                                className="w-100 text-start border-0 bg-transparent px-3 py-2"
                                                onClick={(e) => handleNotificationClick(notification, e)}
                                            >
                                                <div className="d-flex justify-content-between align-items-start mb-1">
                                                    <span className="fw-semibold text-truncate me-2">
                                                        {notification.sender_name || '系统通知'}
                                                    </span>
                                                    <span className="badge bg-secondary-subtle text-secondary-emphasis text-uppercase">
                                                        {notification.type}
                                                    </span>
                                                </div>
                                                <p className={`mb-1 small ${isUnread ? 'fw-semibold' : ''}`}>
                                                    {notification.content || '您有一条新通知'}
                                                </p>
                                                {createdLabel && (
                                                    <small className="text-muted">{createdLabel}</small>
                                                )}
                                            </button>
                                        </li>
                                    );
                                })}
                            </ul>
                        )}
                    </div>

                    <div className="px-3 py-2 border-top bg-body-tertiary d-flex justify-content-between align-items-center">
                        <button
                            type="button"
                            className="btn btn-link btn-sm text-decoration-none"
                            onClick={handleFetchNotifications}
                        >
                            刷新
                        </button>
                        <div className="d-flex align-items-center gap-2">
                            <span className="small text-muted">共 {notifications.length} 条</span>
                            <button
                                type="button"
                                className="btn btn-link btn-sm text-decoration-none"
                                onClick={() => {
                                    closeDropdown();
                                    navigate('/notifications');
                                }}
                            >
                                查看全部
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default NotificationBell;
