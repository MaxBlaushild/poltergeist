export type LogisterProps = {
    logister: (phoneNumber: string, verificationCode: string) => void;
    getVerificationCode: (phoneNumber: string) => void;
    isWaitingOnVerificationCode: boolean;
};
export declare function Logister(props: LogisterProps): import("react/jsx-runtime").JSX.Element;
