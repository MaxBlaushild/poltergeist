var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
export const usePartyMembers = () => {
    const [partyMembers, setPartyMembers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const { apiClient } = useAPI();
    useEffect(() => {
        const fetchPartyMembers = () => __awaiter(void 0, void 0, void 0, function* () {
            try {
                const members = yield apiClient.get(`/sonar/party/members`);
                setPartyMembers(members);
            }
            catch (error) {
                setError(error);
            }
            finally {
                setLoading(false);
            }
        });
        fetchPartyMembers();
    }, []);
    return {
        partyMembers,
        loading,
        error,
    };
};
export const joinParty = (inviterID) => __awaiter(void 0, void 0, void 0, function* () {
    const { apiClient } = useAPI();
    try {
        const response = yield apiClient.post(`/sonar/party/join`, { inviterID });
    }
    catch (error) {
        console.error(error);
    }
});
