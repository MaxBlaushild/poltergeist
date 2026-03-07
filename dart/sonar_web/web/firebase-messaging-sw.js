/* eslint-disable no-undef */

// Fill these with your Firebase Web app values for background web push support.
// They should match the FIREBASE_* --dart-define values used by the app build.
const firebaseConfig = {
  apiKey: "",
  appId: "",
  messagingSenderId: "",
  projectId: "",
  authDomain: "",
  storageBucket: "",
  measurementId: "",
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
