import axios from 'axios';
import { API_BASE_URL } from '../config/api';

const buildAuthHeaders = (token) => ({
    Authorization: `Bearer ${token}`,
});

/**
 * 搜索问题
 * @param {Object} params - 搜索参数
 * @param {string} params.token - 认证 token
 * @param {string} params.query - 搜索关键词
 * @param {number} params.limit - 返回结果的最大数量 (默认: 20)
 * @param {number} params.offset - 分页偏移量 (默认: 0)
 * @returns {Promise<Array>} 搜索结果中的问题列表
 */
export const searchQuestions = async ({ token, query, limit = 20, offset = 0 }) => {
    const response = await axios.get(`${API_BASE_URL}/search/questions`, {
        headers: buildAuthHeaders(token),
        params: {
            query,
            limit,
            offset,
        },
    });

    // 根据 gRPC 响应结构返回 questions 数组
    return response.data?.questions || [];
};
