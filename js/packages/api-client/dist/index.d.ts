export declare class APIClient {
    private client;
    constructor(baseURL: string);
    get<T>(url: string, params?: Record<string, unknown>): Promise<T>;
    post<T>(url: string, data?: Record<string, unknown>): Promise<T>;
    patch<T>(url: string, data?: Record<string, unknown>): Promise<T>;
    delete<T>(url: string): Promise<T>;
}
export default APIClient;
