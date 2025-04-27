// Helper function to recursively get tags from a quest node
function getTagsFromNode(node) {
    const tags = new Set();
    // Add tags from current node's point of interest
    node.pointOfInterest.tags.forEach(tag => tags.add(tag.name));
    // Recursively get tags from children
    Object.values(node.children).forEach(child => {
        getTagsFromNode(child).forEach(tag => tags.add(tag));
    });
    return Array.from(tags);
}
// Implementation of getTags for Quest
export function getQuestTags(quest) {
    return getTagsFromNode(quest.rootNode);
}
