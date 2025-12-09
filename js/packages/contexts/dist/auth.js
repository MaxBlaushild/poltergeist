var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import { jsx as _jsx } from "react/jsx-runtime";
import { createContext, useContext, useState, useEffect } from 'react';
import axios from 'axios';
const tokenKey = 'token';
// Support both Vite and CRA environment variables
const getApiUrl = () => {
    return 'https://api.unclaimedstreets.com';
};
const AuthContext = createContext({
    user: null,
    isWaitingForVerificationCode: false,
    error: null,
    getVerificationCode: () => { },
    logister: () => { },
    isRegister: false,
    logout: () => { },
    loading: false,
});
export const AuthProvider = ({ children, appName, uriPrefix, }) => {
    const token = localStorage.getItem(tokenKey);
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [isRegister, setIsRegister] = useState(false);
    const [isWaitingForVerificationCode, setIsWaitingOnVerificationCode] = useState(false);
    useEffect(() => {
        if (token) {
            setLoading(true);
            const verifyToken = () => __awaiter(void 0, void 0, void 0, function* () {
                try {
                    const response = yield axios.post(`${getApiUrl()}/authenticator/token/verify`, { token });
                    console.log(response.data);
                    setUser(response.data);
                }
                catch (e) {
                    setError(e);
                    setUser(null);
                }
                finally {
                    setLoading(false);
                }
            });
            verifyToken();
        }
        else {
            setLoading(false);
        }
    }, [token]);
    const getVerificationCode = (phoneNumber) => __awaiter(void 0, void 0, void 0, function* () {
        try {
            const { data } = yield axios.post(`${getApiUrl()}/authenticator/text/verification-code`, { phoneNumber, appName });
            setIsWaitingOnVerificationCode(true);
            setIsRegister(!data);
        }
        catch (e) {
            setError(e);
            setIsWaitingOnVerificationCode(false);
        }
    });
    const logister = (phoneNumber, verificationCode) => __awaiter(void 0, void 0, void 0, function* () {
        try {
            const response = yield axios.post(`${getApiUrl()}${uriPrefix}/login`, { phoneNumber, code: verificationCode });
            const { user, token } = response.data;
            localStorage.setItem(tokenKey, token);
            setUser(user);
        }
        catch (e) {
            try {
                const response = yield axios.post(`${getApiUrl()}${uriPrefix}/register`, { phoneNumber, code: verificationCode });
                const { user, token } = response.data;
                localStorage.setItem(tokenKey, token);
                setIsRegister(true);
                setUser(user);
            }
            catch (e) {
                setError(e);
            }
        }
    });
    const logout = () => {
        setUser(null);
        localStorage.removeItem(tokenKey);
    };
    return (_jsx(AuthContext.Provider, Object.assign({ value: {
            user,
            error,
            logister,
            logout,
            isWaitingForVerificationCode,
            getVerificationCode,
            isRegister,
            loading,
        } }, { children: children })));
};
export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};
