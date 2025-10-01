import axios from 'axios';
import { API_BASE_URL } from '../config/api';

const buildAuthHeaders = (token) => ({
    Authorization: `Bearer ${token}`,
});

export const fetchNotifications = async ({ token, limit = 20, offset = 0 }) => {
    const response = await axios.get(`${API_BASE_URL}/notifications`, {
        headers: buildAuthHeaders(token),
        params: { limit, offset },
    });

    return Array.isArray(response.data) ? response.data : response.data?.data || [];
};

export const markNotificationsRead = async ({ token, notificationIds = [] }) => {
    const payload = Array.isArray(notificationIds) ? notificationIds : [notificationIds];
    const response = await axios.post(
        `${API_BASE_URL}/notifications/read`,
        { notification_ids: payload },
        {
            headers: {
                ...buildAuthHeaders(token),
                'Content-Type': 'application/json',
            },
        },
    );

    return response.data;
};

export const deleteNotification = async ({ token, notificationId }) => {
    return axios.delete(`${API_BASE_URL}/notifications/${notificationId}`, {
        headers: buildAuthHeaders(token),
    });
};
