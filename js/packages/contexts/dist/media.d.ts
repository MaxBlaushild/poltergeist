import * as React from 'react';
export type MediaContextType = {
    getPresignedUploadURL: (bucket: string, key: string) => Promise<string | null>;
    uploadMedia: (url: string, file: Blob) => Promise<boolean>;
    openCameraAndCaptureImage: () => Promise<Blob | null>;
};
export declare const MediaContext: React.Context<MediaContextType | undefined>;
export declare const useMediaContext: () => MediaContextType;
export declare const MediaContextProvider: React.FC<{
    children: React.ReactNode;
}>;
