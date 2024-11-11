import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useState } from 'react';
import PhoneInput from 'react-phone-number-input/input';
import { isValidPhoneNumber } from 'react-phone-number-input';
export function Logister(props) {
    const { logister, getVerificationCode, isWaitingOnVerificationCode, isRegister, error, } = props;
    const [code, setCode] = useState('');
    const [phoneNumber, setPhoneNumber] = useState(undefined);
    const [name, setName] = useState('');
    const [profilePictureUrl, setProfilePictureUrl] = useState(undefined);
    const validPhoneNumber = typeof phoneNumber === 'string' && isValidPhoneNumber(phoneNumber);
    return (_jsxs("div", { className: "Logister__container", children: [_jsxs("div", { className: "Logister__inputs", children: [_jsxs("div", { children: [_jsx(PhoneInput, { value: phoneNumber, placeholder: "Phone Number", country: "US", onChange: setPhoneNumber }), isWaitingOnVerificationCode && (_jsx("p", { className: "Logister__disclaimer", children: "We've just sent a 6-digit verification code. It may take a moment to arrive." }))] }), isRegister && isWaitingOnVerificationCode && (_jsx("div", { children: _jsx("input", { placeholder: "Name", type: "text", value: name, onChange: (e) => {
                                const inputValue = e.target.value;
                                setName(inputValue);
                            } }) })), isWaitingOnVerificationCode && (_jsxs("div", { children: [_jsx("input", { type: "text", inputMode: "numeric", pattern: "[0-9]*", placeholder: "Verification code", value: code, autoComplete: "one-time-code", onChange: (e) => {
                                    const inputValue = e.target.value;
                                    if (/^\d*$/.test(inputValue) && inputValue.length <= 6) {
                                        setCode(inputValue);
                                    }
                                } }), error && _jsx("p", { className: "Logister__error", children: error })] }))] }), _jsxs("div", { className: "Logister__buttonBar", children: [!isWaitingOnVerificationCode && validPhoneNumber ? (_jsx("button", { className: "Logister__button", onClick: () => getVerificationCode(phoneNumber), children: "Get code" })) : null, isWaitingOnVerificationCode ? (_jsx("button", { onClick: () => logister(phoneNumber, code, name, isRegister), disabled: code.length !== 6 || (isRegister && name.length === 0), className: "Logister__button", children: isRegister ? 'Register' : 'Login' })) : null] })] }));
}
