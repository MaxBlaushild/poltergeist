var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import { jsx as _jsx } from "react/jsx-runtime";
import { createContext, useState, useEffect, useContext } from 'react';
import { useAPI, useAuth } from '@poltergeist/contexts';
export const TagContext = createContext({
    tagGroups: [],
    selectedTags: [],
    setSelectedTags: () => { },
    createTagGroup: () => { },
    moveTagToTagGroup: () => { },
});
export const TagProvider = ({ children }) => {
    const { apiClient } = useAPI();
    const { user } = useAuth();
    const [tags, setTags] = useState([]);
    const [tagGroups, setTagGroups] = useState([]);
    const [selectedTags, setSelectedTags] = useState([]);
    const fetchTagGroups = () => __awaiter(void 0, void 0, void 0, function* () {
        const response = yield apiClient.get('/sonar/tagGroups');
        setTagGroups(response);
    });
    const createTagGroup = (tagGroup) => __awaiter(void 0, void 0, void 0, function* () {
        const response = yield apiClient.post('/sonar/tagGroups', tagGroup);
        setTagGroups([...tagGroups, response]);
    });
    const moveTagToTagGroup = (tagID, tagGroupID) => __awaiter(void 0, void 0, void 0, function* () {
        yield apiClient.post(`/sonar/tags/move`, { tagID, tagGroupID });
        fetchTagGroups();
    });
    useEffect(() => {
        if (!user) {
            setTagGroups([]);
            return;
        }
        fetchTagGroups();
    }, [user]);
    return (_jsx(TagContext.Provider, Object.assign({ value: { tagGroups, selectedTags, setSelectedTags, createTagGroup, moveTagToTagGroup } }, { children: children })));
};
export const useTagContext = () => {
    return useContext(TagContext);
};
