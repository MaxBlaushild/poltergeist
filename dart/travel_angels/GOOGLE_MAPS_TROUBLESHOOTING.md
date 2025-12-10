# Google Maps Troubleshooting Guide

## Issue: Map shows pin and "Google" watermark but tiles don't load

This indicates the Google Maps SDK is partially working but map tiles aren't loading. Common causes:

### 1. Check API Key Restrictions in Google Cloud Console

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to "APIs & Services" → "Credentials"
3. Click on your API key
4. Check "API restrictions":
   - **For Android**: Ensure "Maps SDK for Android" is enabled
   - **For iOS**: Ensure "Maps SDK for iOS" is enabled
   - **For Backend**: Ensure "Places API" is enabled (for search)

### 2. Check Application Restrictions

**For Android:**
- Under "Application restrictions", select "Android apps"
- Add your app's package name: `com.example.travel_angels`
- Add your SHA-1 certificate fingerprint (get it with):
  ```bash
  keytool -list -v -keystore ~/.android/debug.keystore -alias androiddebugkey -storepass android -keypass android
  ```

**For iOS:**
- Under "Application restrictions", select "iOS apps"
- Add your bundle identifier: `com.example.travelAngels`

### 3. Verify API Key is Correctly Injected

**Android:**
- Check `android/local.properties` contains: `google.maps.key=YOUR_KEY`
- Rebuild the app: `flutter clean && flutter build apk`

**iOS:**
- Check `ios/Flutter/Secrets.xcconfig` contains: `GOOGLE_MAPS_API_KEY=YOUR_KEY`
- Clean build folder in Xcode: Product → Clean Build Folder
- Rebuild the app

### 4. Check Console Logs

Look for these error messages:
- `"Google Maps API key not found"` - Key not injected properly
- `"This API project is not authorized to use this API"` - API not enabled
- `"The provided API key is invalid"` - Wrong key or restrictions too strict

### 5. Enable Required APIs

In Google Cloud Console → "APIs & Services" → "Library", enable:
- ✅ Maps SDK for Android (for Android builds)
- ✅ Maps SDK for iOS (for iOS builds)
- ✅ Places API (for location search)
- ✅ Geocoding API (optional, for reverse geocoding)

### 6. Test API Key Directly

Test your API key with a simple HTTP request:
```bash
# Test Places API (for search)
curl "https://maps.googleapis.com/maps/api/place/textsearch/json?query=New+York&key=YOUR_API_KEY"

# Test Maps SDK (check if key is valid)
curl "https://maps.googleapis.com/maps/api/staticmap?center=40.7128,-74.0060&zoom=13&size=400x400&key=YOUR_API_KEY"
```

### 7. Common Issues

**Issue**: Map shows but tiles are blank/gray
- **Solution**: Maps SDK for Android/iOS not enabled in API restrictions

**Issue**: Search returns no results
- **Solution**: Places API not enabled or API key restrictions too strict

**Issue**: Works on one platform but not the other
- **Solution**: Check platform-specific API restrictions and ensure correct SDK is enabled

**Issue**: Works in debug but not release
- **Solution**: Add release keystore SHA-1 fingerprint to API key restrictions

### 8. Verify Build Configuration

**Android:**
- Check `android/app/build.gradle.kts` loads `local.properties`
- Verify `AndroidManifest.xml` has `${googleMapsApiKey}` placeholder

**iOS:**
- Check `Debug.xcconfig` and `Release.xcconfig` include `Secrets.xcconfig`
- Verify `Info.plist` has `GMSApiKey` key referencing `$(GOOGLE_MAPS_API_KEY)`

### 9. Network Issues

If you're behind a firewall or proxy:
- Ensure network allows connections to `*.googleapis.com`
- Check if corporate firewall blocks Google Maps

### 10. Billing

Google Maps requires a billing account to be set up, even for free tier usage. Ensure billing is enabled in Google Cloud Console.

## Quick Checklist

- [ ] API key has no restrictions OR restrictions allow your app
- [ ] Maps SDK for Android enabled (for Android)
- [ ] Maps SDK for iOS enabled (for iOS)
- [ ] Places API enabled (for search)
- [ ] API key correctly set in `local.properties` (Android)
- [ ] API key correctly set in `Secrets.xcconfig` (iOS)
- [ ] App rebuilt after adding API key
- [ ] Billing enabled in Google Cloud Console
- [ ] Check console logs for specific error messages

