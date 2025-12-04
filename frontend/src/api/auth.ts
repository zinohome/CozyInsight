import request from './request';

interface LoginResponse {
    token: string;
    user: {
        id: string;
        username: string;
        email: string;
        role: string;
    };
}

interface RegisterRequest {
    username: string;
    email: string;
    password: string;
}

export const authAPI = {
    // 用户登录
    login: (username: string, password: string) => {
        return request.post<LoginResponse>('/auth/login', { username, password });
    },

    // 用户注册
    register: (data: RegisterRequest) => {
        return request.post('/auth/register', data);
    },

    // 获取当前用户信息
    me: () => {
        return request.get('/auth/me');
    },

    // 退出登录
    logout: () => {
        return request.post('/auth/logout');
    },
};

export default authAPI;
