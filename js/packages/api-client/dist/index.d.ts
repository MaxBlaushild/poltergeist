export declare class APIClient {
    private client;
    constructor(baseURL: string);
    get<T>(url: string, params?: any): Promise<T>;
    post<T>(url: string, data?: any): Promise<T>;
    patch<T>(url: string, data?: any): Promise<T>;
    delete<T>(url: string): Promise<T>;
}
export default APIClient;
