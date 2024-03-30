import React, { useEffect } from 'react';
import { useAuth } from '@poltergeist/contexts';
import { Logister } from '@poltergeist/components';
import { useNavigate } from 'react-router-dom';

export const Login = () => {
    const navigate = useNavigate();
    const { user, logister, getVerificationCode, isWaitingForVerificationCode, isRegister } = useAuth();

    useEffect(() => {
        if (user) {
            navigate('/surveys');
        }
    }, [user])

    console.log(isRegister)
    
    return <div>
        <Logister 
            logister={logister} 
            getVerificationCode={getVerificationCode}
            isRegister={isRegister}
            isWaitingOnVerificationCode={isWaitingForVerificationCode}
        />
    </div>
}