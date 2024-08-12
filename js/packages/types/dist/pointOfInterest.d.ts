export interface PointOfInterest {
    ID: string;
    createdAt: Date;
    updatedAt: Date;
    name: string;
    clue: string;
    captureChallenge: string;
    attuneChallenge: string;
    lat: string;
    lng: string;
    imageURL: string;
}
