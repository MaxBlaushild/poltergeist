import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useState } from 'react';
import PhoneInput from 'react-phone-number-input/input';
import { isValidPhoneNumber } from 'react-phone-number-input';
export function Logister(props) {
    const { logister, getVerificationCode, isWaitingOnVerificationCode } = props;
    const [code, setCode] = useState('');
    const [phoneNumber, setPhoneNumber] = useState(undefined);
    const validPhoneNumber = typeof phoneNumber === 'string' && isValidPhoneNumber(phoneNumber);
    return (_jsx("div", { children: _jsxs("div", { children: [_jsx("p", { children: "Sign in or sign up" }), _jsxs("div", { children: [_jsxs("div", { children: [_jsx(PhoneInput, { value: phoneNumber, placeholder: "+1 234 567 8900", onChange: setPhoneNumber }), !isWaitingOnVerificationCode && (_jsx("button", Object.assign({ onClick: () => getVerificationCode(phoneNumber), disabled: !validPhoneNumber }, { children: "Login" })))] }), isWaitingOnVerificationCode && (_jsxs("div", { children: [_jsx("input", { type: "text", inputMode: "numeric", pattern: "[0-9]*", value: code, autoComplete: "one-time-code", onChange: (e) => {
                                        const inputValue = e.target.value;
                                        if (/^\d*$/.test(inputValue) && inputValue.length <= 6) {
                                            setCode(inputValue);
                                        }
                                    } }), _jsx("button", Object.assign({ onClick: () => logister(phoneNumber, code), disabled: code.length !== 6 }, { children: "Enter code" }))] }))] })] }) }));
}
