import { ReactNode } from 'react';
interface Location {
    latitude: number | null;
    longitude: number | null;
    accuracy: number | null;
}
interface LocationContextType {
    location: Location | null;
    isLoading: boolean;
    error: string | null;
}
export declare const LocationProvider: ({ children }: {
    children: ReactNode;
}) => import("react/jsx-runtime").JSX.Element;
export declare const useLocation: () => LocationContextType;
export {};
