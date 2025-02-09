interface UseCityNameResult {
    cityName: string | null;
    loading: boolean;
    error: Error | null;
}
export declare const useCityName: (latitude: string, longitude: string) => UseCityNameResult;
export {};
