export enum GenerationStatus {
  Unidentified,
  Pending,
  InProgress,
  GenerateImageOptions,
  Complete,
  Failed
}

export enum GenerationBackend {
  Unidentified,
  Imagine,
  UseApi
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
