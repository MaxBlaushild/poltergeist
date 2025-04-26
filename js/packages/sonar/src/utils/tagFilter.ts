import { Tag } from "@poltergeist/types";

export const tagsToFilter = [
  'store',
  'establishment',
  'point_of_interest',
  'food',
  'health',
  'food_store',
  'supermarket',
  'market'
]

const tagFilter = (tags: Tag[]): Tag[] => {
  return tags.filter(tag => !tagsToFilter.includes(tag.name));
}

export default tagFilter;