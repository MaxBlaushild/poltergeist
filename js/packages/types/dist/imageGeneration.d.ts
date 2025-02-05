export declare enum GenerationStatus {
    Unidentified = 0,
    Pending = 1,
    InProgress = 2,
    GenerateImageOptions = 3,
    Complete = 4,
    Failed = 5
}
export declare enum GenerationBackend {
    Unidentified = 0,
    Imagine = 1,
    UseApi = 2
}
export type ImageGeneration = {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    userId: string;
    generationId: string;
    generationBackendId: GenerationBackend;
    status: GenerationStatus;
    optionOne: string | null;
    optionTwo: string | null;
    optionThree: string | null;
    optionFour: string | null;
};
