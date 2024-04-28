export type LogisterProps = {
    logister: (phoneNumber: string, verificationCode: string, name: string) => void;
    getVerificationCode: (phoneNumber: string) => void;
    isRegister: boolean;
    isWaitingOnVerificationCode: boolean;
    error: string | undefined;
};
export declare function Logister(props: LogisterProps): import("react/jsx-runtime").JSX.Element;
