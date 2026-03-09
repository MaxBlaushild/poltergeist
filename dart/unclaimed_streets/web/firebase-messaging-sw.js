/* eslint-disable no-undef */

const currentUrl = new URL(self.location.href);
const firebaseConfig = {
  apiKey: currentUrl.searchParams.get("apiKey") || "",
  appId: currentUrl.searchParams.get("appId") || "",
  messagingSenderId: currentUrl.searchParams.get("messagingSenderId") || "",
  projectId: currentUrl.searchParams.get("projectId") || "",
  authDomain: currentUrl.searchParams.get("authDomain") || "",
  storageBucket: currentUrl.searchParams.get("storageBucket") || "",
  measurementId: currentUrl.searchParams.get("measurementId") || "",
};

const requiredKeys = ["apiKey", "appId", "messagingSenderId", "projectId"];
const isConfigured = requiredKeys.every((key) => {
  const value = firebaseConfig[key];
  return typeof value === "string" && value.trim().length > 0;
});

if (isConfigured) {
  importScripts("https://www.gstatic.com/firebasejs/10.13.2/firebase-app-compat.js");
  importScripts(
    "https://www.gstatic.com/firebasejs/10.13.2/firebase-messaging-compat.js",
  );
  firebase.initializeApp(firebaseConfig);
  firebase.messaging();
}
