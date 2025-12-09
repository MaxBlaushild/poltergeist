import React, { createContext, useContext, useState, ReactNode, useEffect } from 'react';
import { User } from '@poltergeist/types';
import axios from 'axios';

const tokenKey = 'token';

// Support both Vite and CRA environment variables
  const getApiUrl = () => {
    return 'https://api.unclaimedstreets.com';
  }

type AuthContextType = {
  user: User | null;
  isWaitingForVerificationCode: boolean;
  error: unknown;
  loading: boolean;
  getVerificationCode: (phoneNumber: string) => void;
  logister: (
    phoneNumber: string,
    verificationCode: string,
    name: string
  ) => void;
  logout: () => void;
  isRegister: boolean;
};

const AuthContext = createContext<AuthContextType>({
  user: null,
  isWaitingForVerificationCode: false,
  error: null,
  getVerificationCode: () => {},
  logister: () => {},
  isRegister: false,
  logout: () => {},
  loading: false,
});

type AuthProviderProps = {
  children: ReactNode;
  appName: string;
  uriPrefix: string;
};

export const AuthProvider = ({
  children,
  appName,
  uriPrefix,
}: AuthProviderProps) => {
  const token = localStorage.getItem(tokenKey);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<unknown>(null);
  const [isRegister, setIsRegister] = useState<boolean>(false);
  
  const [isWaitingForVerificationCode, setIsWaitingOnVerificationCode] =
    useState<boolean>(false);

  useEffect(() => {
    if (token) {
      setLoading(true);
      const verifyToken = async () => {
        try {
          const response = await axios.post(
            `${getApiUrl()}/authenticator/token/verify`,
            { token },
          );
          console.log(response.data);
          setUser(response.data);
        } catch (e) {
          setError(e);
          setUser(null);
        } finally {
          setLoading(false);
        }
      };
      verifyToken();
    } else {
      setLoading(false);
    }
  }, [token]);

  const getVerificationCode = async (phoneNumber: string) => {
    try {
      const { data } = await axios.post(
        `${getApiUrl()}/authenticator/text/verification-code`,
        { phoneNumber, appName }
      );
      setIsWaitingOnVerificationCode(true);
      setIsRegister(!data);
    } catch (e) {
      setError(e);
      setIsWaitingOnVerificationCode(false);
    }
  };

  const logister = async (
    phoneNumber: string,
    verificationCode: string,
  ) => {
    try {
      const response = await axios.post(
        `${getApiUrl()}${uriPrefix}/login`,
        { phoneNumber, code: verificationCode }
      );
      const { user, token } = response.data;

      localStorage.setItem(tokenKey, token);

      setUser(user);
    } catch (e) {
      try {
        const response = await axios.post(
          `${getApiUrl()}${uriPrefix}/register`,
          { phoneNumber, code: verificationCode }
        );
        const { user, token } = response.data;

        localStorage.setItem(tokenKey, token);
        setIsRegister(true);
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
    <AuthContext.Provider
      value={{
        user,
        error,
        logister,
        logout,
        isWaitingForVerificationCode,
        getVerificationCode,
        isRegister,
        loading,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
