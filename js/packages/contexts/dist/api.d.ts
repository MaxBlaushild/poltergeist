import React, { ReactNode } from 'react';
import APIClient from '@poltergeist/api-client';
interface APIContextType {
    apiClient: APIClient;
}
interface APIProviderProps {
    children: ReactNode;
}
export declare const APIProvider: React.FC<APIProviderProps>;
export declare const useAPI: () => APIContextType;
export {};
