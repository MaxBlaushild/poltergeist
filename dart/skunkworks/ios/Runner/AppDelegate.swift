import Flutter
import UIKit
import Security
import Foundation

@main
@objc class AppDelegate: FlutterAppDelegate {
  override func application(
    _ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
  ) -> Bool {
    GeneratedPluginRegistrant.register(with: self)
    
    guard let controller = window?.rootViewController as? FlutterViewController else {
      return super.application(application, didFinishLaunchingWithOptions: launchOptions)
    }
    
    let secureEnclaveChannel = FlutterMethodChannel(name: "com.verifiablesn/secure_enclave",
                                                    binaryMessenger: controller.binaryMessenger)
    
    secureEnclaveChannel.setMethodCallHandler({
      (call: FlutterMethodCall, result: @escaping FlutterResult) -> Void in
      
      switch call.method {
      case "generateKeyPair":
        self.generateKeyPair(result: result)
      case "signData":
        guard let args = call.arguments as? [String: Any],
              let data = args["data"] as? FlutterStandardTypedData,
              let keyTag = args["keyTag"] as? String else {
          result(FlutterError(code: "INVALID_ARGS", message: "Invalid arguments", details: nil))
          return
        }
        self.signData(data: data.data, keyTag: keyTag, result: result)
      case "getPublicKey":
        guard let args = call.arguments as? [String: Any],
              let keyTag = args["keyTag"] as? String else {
          result(FlutterError(code: "INVALID_ARGS", message: "Invalid arguments", details: nil))
          return
        }
        self.getPublicKey(keyTag: keyTag, result: result)
      case "isAvailable":
        // Check if device supports Secure Enclave
        // Secure Enclave is available on devices with A7 processor and later
        #if targetEnvironment(simulator)
        result(false)
        #else
        result(true)
        #endif
      default:
        result(FlutterMethodNotImplemented)
      }
    })
    
    return super.application(application, didFinishLaunchingWithOptions: launchOptions)
  }
  
  private func generateKeyPair(result: @escaping FlutterResult) {
    let keyTag = "com.verifiablesn.certificate.key"
    let access = SecAccessControlCreateWithFlags(
      kCFAllocatorDefault,
      kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
      [.privateKeyUsage, .biometryAny],
      nil
    )!
    
    let attributes: [String: Any] = [
      kSecAttrKeyType as String: kSecAttrKeyTypeECSECPrimeRandom,
      kSecAttrKeySizeInBits as String: 256,
      kSecAttrTokenID as String: kSecAttrTokenIDSecureEnclave,
      kSecPrivateKeyAttrs as String: [
        kSecAttrIsPermanent as String: true,
        kSecAttrApplicationTag as String: keyTag.data(using: .utf8)!,
        kSecAttrAccessControl as String: access,
      ]
    ]
    
    var error: Unmanaged<CFError>?
    guard let privateKey = SecKeyCreateRandomKey(attributes as CFDictionary, &error) else {
      result(FlutterError(code: "KEY_GENERATION_FAILED", message: error?.takeRetainedValue().localizedDescription, details: nil))
      return
    }
    
    result(keyTag)
  }
  
  private func signData(data: Data, keyTag: String, result: @escaping FlutterResult) {
    let query: [String: Any] = [
      kSecClass as String: kSecClassKey,
      kSecAttrApplicationTag as String: keyTag.data(using: .utf8)!,
      kSecAttrKeyType as String: kSecAttrKeyTypeECSECPrimeRandom,
      kSecReturnRef as String: true
    ]
    
    var item: CFTypeRef?
    let status = SecItemCopyMatching(query as CFDictionary, &item)
    
    guard status == errSecSuccess, let keyRef = item else {
      result(FlutterError(code: "KEY_NOT_FOUND", message: "Key not found", details: nil))
      return
    }
    
    let privateKey = keyRef as! SecKey
    
    var error: Unmanaged<CFError>?
    guard let signature = SecKeyCreateSignature(
      privateKey,
      .ecdsaSignatureMessageX962SHA256,
      data as CFData,
      &error
    ) as Data? else {
      result(FlutterError(code: "SIGN_FAILED", message: error?.takeRetainedValue().localizedDescription, details: nil))
      return
    }
    
    result(FlutterStandardTypedData(bytes: signature))
  }
  
  private func getPublicKey(keyTag: String, result: @escaping FlutterResult) {
    let query: [String: Any] = [
      kSecClass as String: kSecClassKey,
      kSecAttrApplicationTag as String: keyTag.data(using: .utf8)!,
      kSecAttrKeyType as String: kSecAttrKeyTypeECSECPrimeRandom,
      kSecReturnRef as String: true
    ]
    
    var item: CFTypeRef?
    let status = SecItemCopyMatching(query as CFDictionary, &item)
    
    guard status == errSecSuccess, let keyRef = item else {
      result(FlutterError(code: "KEY_NOT_FOUND", message: "Key not found", details: nil))
      return
    }
    
    let privateKey = keyRef as! SecKey
    
    guard let publicKey = SecKeyCopyPublicKey(privateKey) else {
      result(FlutterError(code: "PUBLIC_KEY_FAILED", message: "Failed to extract public key", details: nil))
      return
    }
    
    var error: Unmanaged<CFError>?
    guard let publicKeyData = SecKeyCopyExternalRepresentation(publicKey, &error) as Data? else {
      result(FlutterError(code: "EXPORT_FAILED", message: error?.takeRetainedValue().localizedDescription, details: nil))
      return
    }
    
    // Convert to PEM format (simplified - in production, use proper ASN.1 encoding)
    let base64Key = publicKeyData.base64EncodedString()
    let pem = "-----BEGIN PUBLIC KEY-----\n\(base64Key)\n-----END PUBLIC KEY-----"
    
    result(pem)
  }
}
