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
import { useAPI } from '@poltergeist/contexts';
export const TagContext = createContext({
    tagGroups: [],
    selectedTags: [],
    setSelectedTags: () => { },
});
export const TagProvider = ({ children }) => {
    const { apiClient } = useAPI();
    const [tags, setTags] = useState([]);
    const [tagGroups, setTagGroups] = useState([]);
    const [selectedTags, setSelectedTags] = useState([]);
    const fetchTagGroups = () => __awaiter(void 0, void 0, void 0, function* () {
        const response = yield apiClient.get('/sonar/tagGroups');
        setTagGroups(response);
        setSelectedTags(response.flatMap(group => group.tags));
    });
    useEffect(() => {
        fetchTagGroups();
    }, []);
    return (_jsx(TagContext.Provider, Object.assign({ value: { tagGroups, selectedTags, setSelectedTags } }, { children: children })));
};
export const useTagContext = () => {
    return useContext(TagContext);
};
