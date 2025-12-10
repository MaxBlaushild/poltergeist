# Google Maps API Key Setup

This document explains how to securely configure your Google Maps API key for both Android and iOS platforms.

## Security Best Practices

**Never commit API keys directly to version control.** The configuration files that contain keys are gitignored.

## Android Setup

1. **Create or edit `android/local.properties`** (this file is already gitignored):
   ```properties
   google.maps.key=YOUR_GOOGLE_MAPS_API_KEY_HERE
   ```

2. The key will be automatically injected into `AndroidManifest.xml` at build time via `build.gradle.kts`.

3. **Verify setup**: The `AndroidManifest.xml` should reference `${googleMapsApiKey}` which will be replaced at build time.

## iOS Setup

1. **Create `ios/Flutter/Secrets.xcconfig`** (copy from the example file):
   ```bash
   cp ios/Flutter/Secrets.xcconfig.example ios/Flutter/Secrets.xcconfig
   ```

2. **Edit `ios/Flutter/Secrets.xcconfig`** and add your actual API key:
   ```
   GOOGLE_MAPS_API_KEY=YOUR_GOOGLE_MAPS_API_KEY_HERE
   ```

3. The `Debug.xcconfig` and `Release.xcconfig` files already include `Secrets.xcconfig`, so the key will be available at build time.

4. **Verify setup**: The `Info.plist` references `$(GOOGLE_MAPS_API_KEY)` which will be replaced at build time.

## Getting Your Google Maps API Key

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Maps SDK for Android and/or Maps SDK for iOS
4. Go to "Credentials" → "Create Credentials" → "API Key"
5. Copy your API key

## Important: Restrict Your API Key

**Always restrict your API key in Google Cloud Console:**

1. Go to your API key settings
2. Under "Application restrictions":
   - For Android: Add your app's package name and SHA-1 certificate fingerprint
   - For iOS: Add your app's bundle identifier
3. Under "API restrictions": Restrict to only the Maps SDK APIs you need

This prevents unauthorized usage even if the key is leaked.

## Troubleshooting

### Android
- Ensure `android/local.properties` exists and contains `google.maps.key=YOUR_KEY`
- Clean and rebuild: `flutter clean && flutter build apk`

### iOS
- Ensure `ios/Flutter/Secrets.xcconfig` exists and contains `GOOGLE_MAPS_API_KEY=YOUR_KEY`
- Clean build folder in Xcode: Product → Clean Build Folder
- Rebuild the app

## Files Reference

- `android/local.properties` - Android API key (gitignored)
- `ios/Flutter/Secrets.xcconfig` - iOS API key (gitignored)
- `android/local.properties.example` - Example template for Android
- `ios/Flutter/Secrets.xcconfig.example` - Example template for iOS

