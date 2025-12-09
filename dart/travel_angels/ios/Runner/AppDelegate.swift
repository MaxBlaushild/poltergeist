import Flutter
import UIKit
import GoogleMaps

@main
@objc class AppDelegate: FlutterAppDelegate {
  override func application(
    _ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
  ) -> Bool {
    // Get API key from Info.plist (injected from GoogleMaps.xcconfig via build settings)
    if let apiKey = Bundle.main.object(forInfoDictionaryKey: "GMS_API_KEY") as? String,
       !apiKey.isEmpty {
      GMSServices.provideAPIKey(apiKey)
    } else {
      // Warn but don't crash - app can still run without Google Maps
      print("WARNING: Google Maps API key not found in Info.plist!")
      print("Google Maps features will not work until the API key is configured.")
      print("")
      print("To fix this, ensure:")
      print("1. The file ios/Flutter/GoogleMaps.xcconfig exists and contains: GMS_API_KEY = YOUR_API_KEY")
      print("2. Debug.xcconfig and Release.xcconfig include: #include? \"GoogleMaps.xcconfig\"")
      print("3. Clean build folder (Product → Clean Build Folder in Xcode)")
      print("4. Rebuild the app")
      print("")
      print("See ios/Flutter/GoogleMaps.xcconfig.example for reference.")
    }
    
    GeneratedPluginRegistrant.register(with: self)
    return super.application(application, didFinishLaunchingWithOptions: launchOptions)
  }
}
