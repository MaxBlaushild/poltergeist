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
import { createContext, useContext, useState } from 'react';
import axios from 'axios';
const tokenKey = 'token';
const AuthContext = createContext({
    user: null,
    isWaitingForVerificationCode: false,
    error: null,
    getVerificationCode: () => { },
    logister: () => { },
    isRegister: false,
    logout: () => { }
});
export const AuthProvider = ({ children, appName, uriPrefix }) => {
    const [user, setUser] = useState(null);
    const [error, setError] = useState(null);
    const [isRegister, setIsRegister] = useState(false);
    const [isWaitingForVerificationCode, setIsWaitingOnVerificationCode] = useState(false);
    const getVerificationCode = (phoneNumber) => __awaiter(void 0, void 0, void 0, function* () {
        try {
            const { data } = yield axios.post(`${process.env.REACT_APP_API_URL}/authenticator/text/verification-code`, { phoneNumber, appName, });
            setIsWaitingOnVerificationCode(true);
            setIsRegister(!data);
        }
        catch (e) {
            setError(e);
            setIsWaitingOnVerificationCode(false);
        }
    });
    const logister = (phoneNumber, verificationCode, name) => __awaiter(void 0, void 0, void 0, function* () {
        try {
            const response = yield axios.post(`${process.env.REACT_APP_API_URL}${uriPrefix}/login`, { phoneNumber, code: verificationCode });
            const { user, token } = response.data;
            localStorage.setItem(tokenKey, token);
            setUser(user);
        }
        catch (e) {
            try {
                const response = yield axios.post(`${process.env.REACT_APP_API_URL}${uriPrefix}/register`, { phoneNumber, code: verificationCode, name });
                const { user, token } = response.data;
                localStorage.setItem(tokenKey, token);
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
        } }, { children: children })));
};
export const useAuth = () => {
    return useContext(AuthContext);
};
