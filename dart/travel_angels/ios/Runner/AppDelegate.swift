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
      // Debug: Print all Info.plist keys and values
      print("DEBUG: Info.plist contents:")
      if let infoDict = Bundle.main.infoDictionary {
        for (key, value) in infoDict.sorted(by: { $0.key < $1.key }) {
          print("  \(key): \(value)")
        }
      } else {
        print("  (Info.plist is nil)")
      }
      
      fatalError("""
      ERROR: Google Maps API key not found in Info.plist!
      
      The build setting INFOPLIST_KEY_GMS_API_KEY = $(GMS_API_KEY) is not working.
      The GMS_API_KEY variable from GoogleMaps.xcconfig is not being injected into Info.plist.
      
      Please ensure:
      1. The file ios/Flutter/GoogleMaps.xcconfig exists and contains: GMS_API_KEY = YOUR_API_KEY
      2. Debug.xcconfig and Release.xcconfig include: #include? "GoogleMaps.xcconfig"
      3. Clean build folder (Product → Clean Build Folder in Xcode)
      4. Rebuild the app
      
      See ios/Flutter/GoogleMaps.xcconfig.example for reference.
      """)
    }
    
    GeneratedPluginRegistrant.register(with: self)
    return super.application(application, didFinishLaunchingWithOptions: launchOptions)
  }
}
