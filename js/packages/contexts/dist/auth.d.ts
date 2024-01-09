import { ReactNode } from 'react';
import { User } from '@poltergeist/types';
type AuthContextType = {
    user: User | null;
    isWaitingForVerificationCode: boolean;
    error: unknown;
    getVerificationCode: (phoneNumber: string) => void;
    logister: (phoneNumber: string, verificationCode: string) => void;
    logout: () => void;
};
type AuthProviderProps = {
    children: ReactNode;
    appName: string;
    uriPrefix: string;
};
export declare const AuthProvider: ({ children, appName, uriPrefix }: AuthProviderProps) => import("react/jsx-runtime").JSX.Element;
export declare const useAuth: () => AuthContextType;
export {};
