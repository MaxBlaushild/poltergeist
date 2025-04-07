import { Tag } from "./tag";
export type TagGroup = {
    id: string;
    name: string;
    createdAt: string;
    updatedAt: string;
    tags: Tag[];
};
