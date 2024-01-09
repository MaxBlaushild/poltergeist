Object.defineProperty(exports, '__esModule', { value: true });
exports.getUserID = void 0;
const uuid_1 = require('uuid');
const getUserID = () => {
  const id = localStorage.getItem('user-id');
  if (id) {
    try {
      return {
        userId: id,
      };
    } catch (e) {
      localStorage.removeItem('user-id');
    }
  }
  let ephemeralUserId = localStorage.getItem('ephemeral-user-id');
  if (!ephemeralUserId) {
    ephemeralUserId = (0, uuid_1.v4)();
    localStorage.setItem('ephemeral-user-id', ephemeralUserId);
  }
  return { ephemeralUserId };
};
exports.getUserID = getUserID;
//# sourceMappingURL=util.js.map
