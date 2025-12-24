import { config } from "../config";
import { atom } from "nanostores";
export const JWT_TOKEN_KEY = 'sortedchat.jwt' as const;


interface JWT {
    sub: string;      // user id
    email: string;
    roles: string[];
    iss: string;      // issuer
    exp: number;      // expiration timestamp
    iat: number;      // issued at timestamp
    name: string;
}


export function decodeJWTPayload(token: string): JWT | null {
    try {
        const parts = token.split('.');
        if (parts.length !== 3) {
            return null;
        }

        const payload = parts[1];
        const decoded = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')));
        console.log('decoded', decoded);
        return decoded as JWT;
    } catch (error) {
        console.error('Failed to decode JWT payload:', error);
        return null;
    }
}

interface UserLoggedIn {
    isLoggedIn: boolean;
    user: JWT | null;
}
export const $LoggedInUser = atom<UserLoggedIn>({ isLoggedIn: false, user: null });

export function isLoggedIn(): boolean {
    const token = localStorage.getItem(JWT_TOKEN_KEY);
    if (!token) {
        return false;
    }
    const payload = decodeJWTPayload(token);
    if (!payload) {
        return false;
    }
    
    $LoggedInUser.set({ isLoggedIn: true, user: payload });
    return true;
}

export function getJWTToken(): string | null {
    return localStorage.getItem(JWT_TOKEN_KEY);
}

export function Logout() {
    localStorage.removeItem(JWT_TOKEN_KEY);
    document.cookie = `${JWT_TOKEN_KEY}=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT`;
    window.location.href = config.appUrl;
    $LoggedInUser.set({ isLoggedIn: false, user: null });
}
