const DEFAULT_API_BASE_URL = 'http://localhost:8080/api/v1';

const buildDefaultWsUrl = (apiBaseUrl) => {
    try {
        const url = new URL(apiBaseUrl);
        const normalizedPath = url.pathname.endsWith('/')
            ? `${url.pathname.slice(0, -1)}/ws`
            : `${url.pathname}/ws`;

        url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
        url.pathname = normalizedPath;
        url.search = '';
        url.hash = '';
        return url.toString();
    } catch (error) {
        console.warn('[config/api] Failed to derive WebSocket URL from API base, falling back to default.', error);
        return 'ws://localhost:8080/api/v1/ws';
    }
};

export const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || DEFAULT_API_BASE_URL;
export const WS_NOTIFICATIONS_URL = process.env.REACT_APP_WS_NOTIFICATIONS_URL || buildDefaultWsUrl(API_BASE_URL);
