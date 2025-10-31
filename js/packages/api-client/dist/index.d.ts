import { UserZoneReputation } from '@poltergeist/types';
interface Location {
    latitude: number | null;
    longitude: number | null;
    accuracy: number | null;
}
export declare class APIClient {
    private client;
    private getLocation?;
    constructor(baseURL: string, getLocation?: () => Location | null);
    get<T>(url: string, params?: any): Promise<T>;
    post<T>(url: string, data?: any): Promise<T>;
    put<T>(url: string, data?: any): Promise<T>;
    patch<T>(url: string, data?: any): Promise<T>;
    delete<T>(url: string, data?: any): Promise<T>;
    getUserReputations(): Promise<UserZoneReputation[]>;
}
export default APIClient;
