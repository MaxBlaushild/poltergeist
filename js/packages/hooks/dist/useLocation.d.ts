interface Location {
    latitude: number | null;
    longitude: number | null;
    accuracy: number | null;
}
export declare const useLocation: () => {
    location: Location;
    error: string | null;
};
export default useLocation;
