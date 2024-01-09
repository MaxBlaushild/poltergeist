import React, { createContext, useContext, useState, ReactNode } from 'react';
import { User } from '@poltergeist/types';
import axios from 'axios';

const tokenKey = 'token';

type AuthContextType = {
    user: User | null;
    isWaitingForVerificationCode: boolean;
    error: unknown;
    getVerificationCode: (phoneNumber: string) => void;
    logister: (phoneNumber: string, verificationCode: string) => void;
    logout: () => void;
};

const AuthContext = createContext<AuthContextType>({
    user: null,
    isWaitingForVerificationCode: false,
    error: null,
    getVerificationCode: () => {},
    logister: () => {},
    logout: () => {}
});

type AuthProviderProps = {
    children: ReactNode;
    appName: string;
    uriPrefix: string;
};

export const AuthProvider = ({ children, appName, uriPrefix }: AuthProviderProps) => {
    const [user, setUser] = useState<User | null>(null);
    const [error, setError] = useState<unknown>(null);
    const [isWaitingForVerificationCode, setIsWaitingOnVerificationCode] = useState<boolean>(false);

    const getVerificationCode = async (phoneNumber: string) => {
        try {
            await axios.post(
              `${process.env.REACT_APP_API_URL}/authenticator/text/verification-code`,
              { phoneNumber, appName, }
            );
            setIsWaitingOnVerificationCode(true);
          } catch (e) {
            setError(e);
            setIsWaitingOnVerificationCode(false);
          }
    };

    const logister = async (phoneNumber: string, verificationCode: string) => {
      try {
          const response = await axios.post(
            `${process.env.REACT_APP_API_URL}${uriPrefix}/login`,
            { phoneNumber, code: verificationCode }
          );
          const { user, token } = response.data;
    
          localStorage.setItem(tokenKey, token);
    
          setUser(user);
        } catch (e) {
          try {
            const response = await axios.post(
              `${process.env.REACT_APP_API_URL}${uriPrefix}/register`,
              { phoneNumber, code: verificationCode, name: '' }
            );
            const { user, token } = response.data;
    
            localStorage.setItem(tokenKey, token);
    
            setUser(user);
          } catch (e) {
            setError(e);
          }
        }
    };

    const logout = () => {
        setUser(null);
        localStorage.removeItem(tokenKey);
    };

    return (
        <AuthContext.Provider value={
          { 
            user,
            error,
            logister, 
            logout, 
            isWaitingForVerificationCode, 
            getVerificationCode
          }}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => {
    return useContext(AuthContext);
}