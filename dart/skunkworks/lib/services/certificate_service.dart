import 'dart:convert';
import 'dart:typed_data';
import 'package:cryptography/cryptography.dart';
import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/models/certificate.dart';
import 'package:skunkworks/models/user.dart';
import 'package:skunkworks/services/api_client.dart';

class CertificateService {
  final APIClient _apiClient;
  final FlutterSecureStorage _secureStorage;
  static const String _keyPairTag = 'certificate_keypair';
  static const String _certificateKey = 'user_certificate';

  CertificateService(this._apiClient)
      : _secureStorage = const FlutterSecureStorage();

  /// Generates a new ECDSA keypair (P-256)
  /// For iOS, this should eventually use Secure Enclave via platform channels
  Future<SimpleKeyPair> generateKeyPair() async {
    final algorithm = EcdsaP256();
    return await algorithm.newKeyPair();
  }

  /// Extracts the public key in PEM format
  /// Note: This is a simplified implementation. For production, proper ASN.1 encoding
  /// should be used to create a valid X.509 SubjectPublicKeyInfo structure.
  Future<String> getPublicKeyPEM(SimpleKeyPair keyPair) async {
    final algorithm = EcdsaP256();
    final publicKey = await keyPair.extractPublicKey();
    
    // The public key bytes from the cryptography package need to be converted
    // to X.509 SubjectPublicKeyInfo format. For now, we'll use a simplified approach.
    // In production, use pointycastle to properly encode the public key.
    final publicKeyBytes = publicKey.bytes;

    // Convert to PEM format (simplified - proper ASN.1 encoding needed for production)
    final base64Key = base64Encode(publicKeyBytes);
    // Split into 64-character lines for PEM format
    final lines = <String>[];
    for (int i = 0; i < base64Key.length; i += 64) {
      lines.add(base64Key.substring(i, (i + 64 < base64Key.length) ? i + 64 : base64Key.length));
    }
    return '-----BEGIN PUBLIC KEY-----\n${lines.join('\n')}\n-----END PUBLIC KEY-----';
  }

  /// Signs a challenge with the private key
  Future<Uint8List> signChallenge(Uint8List challenge, SimpleKeyPair keyPair) async {
    final algorithm = EcdsaP256();
    final signature = await algorithm.sign(
      challenge,
      keyPair: keyPair,
    );
    // ECDSA signature is typically 64 bytes (r: 32 bytes, s: 32 bytes)
    // The cryptography package returns it in a format we need to convert
    return signature.bytes;
  }

  /// Creates a challenge based on user ID and public key
  Future<Uint8List> createChallenge(String userId, String publicKeyPEM) async {
    final data = '$userId:$publicKeyPEM';
    final algorithm = Sha256();
    final hash = await algorithm.hash(utf8.encode(data));
    return Uint8List.fromList(hash.bytes);
  }

  /// Checks if the user has a certificate
  Future<bool> hasCertificate() async {
    try {
      final response = await _apiClient.get<Map<String, dynamic>>(
        ApiConstants.checkCertificateEndpoint,
      );
      return response['hasCertificate'] as bool? ?? false;
    } catch (e) {
      if (e is DioException && e.response?.statusCode == 404) {
        return false;
      }
      rethrow;
    }
  }

  /// Enrolls a certificate by generating a keypair, signing a challenge, and sending to backend
  Future<Certificate> enrollCertificate(User user) async {
    try {
      // Generate keypair
      final keyPair = await generateKeyPair();
      
      // Extract public key in PEM format
      final publicKeyPEM = await getPublicKeyPEM(keyPair);
      
      // Create challenge
      final challenge = await createChallenge(user.id!, publicKeyPEM);
      
      // Sign challenge
      final signature = await signChallenge(challenge, keyPair);
      
      // Encode signature as base64
      final signatureBase64 = base64Encode(signature);
      
      // Send enrollment request
      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.enrollCertificateEndpoint,
        data: {
          'publicKey': publicKeyPEM,
          'challengeSignature': signatureBase64,
        },
      );
      
      final certificate = Certificate.fromJson(response);
      
      // Store certificate locally
      await storeCertificateLocally(certificate);
      
      // Store keypair reference (in a real implementation, this would be a reference to Secure Enclave)
      // For now, we'll store a flag indicating the keypair exists
      await _secureStorage.write(key: _keyPairTag, value: 'exists');
      
      return certificate;
    } catch (e) {
      rethrow;
    }
  }

  /// Gets the certificate from the backend
  Future<Certificate?> getCertificate() async {
    try {
      final response = await _apiClient.get<Map<String, dynamic>>(
        ApiConstants.getCertificateEndpoint,
      );
      final certificate = Certificate.fromJson(response);
      
      // Store certificate locally
      await storeCertificateLocally(certificate);
      
      return certificate;
    } catch (e) {
      if (e is DioException && e.response?.statusCode == 404) {
        return null;
      }
      rethrow;
    }
  }

  /// Stores certificate locally in secure storage
  Future<void> storeCertificateLocally(Certificate certificate) async {
    final certificateJson = jsonEncode(certificate.toJson());
    await _secureStorage.write(key: _certificateKey, value: certificateJson);
  }

  /// Retrieves certificate from local secure storage
  Future<Certificate?> getCertificateLocally() async {
    final certificateJson = await _secureStorage.read(key: _certificateKey);
    if (certificateJson == null) {
      return null;
    }
    final certificateMap = jsonDecode(certificateJson) as Map<String, dynamic>;
    return Certificate.fromJson(certificateMap);
  }

  /// Clears the locally stored certificate
  Future<void> clearCertificate() async {
    await _secureStorage.delete(key: _certificateKey);
    await _secureStorage.delete(key: _keyPairTag);
  }
}
