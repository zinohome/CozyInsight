import { create } from 'zustand';

interface UserState {
    token: string | null;
    userInfo: any | null;
    setToken: (token: string) => void;
    setUserInfo: (info: any) => void;
    logout: () => void;
}

export const useUserStore = create<UserState>((set) => ({
    token: localStorage.getItem('token'),
    userInfo: null,
    setToken: (token) => {
        localStorage.setItem('token', token);
        set({ token });
    },
    setUserInfo: (info) => set({ userInfo: info }),
    logout: () => {
        localStorage.removeItem('token');
        set({ token: null, userInfo: null });
    },
}));
