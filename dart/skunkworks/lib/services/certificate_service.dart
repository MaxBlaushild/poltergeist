import 'dart:convert';
import 'dart:typed_data';
import 'package:asn1lib/asn1lib.dart';
import 'package:crypto/crypto.dart';
import 'package:dio/dio.dart';
import 'package:ecdsa/ecdsa.dart' as ecdsa;
import 'package:elliptic/elliptic.dart' as elliptic;
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

  /// Generates a new ECDSA keypair (P-256) using elliptic package
  /// For iOS, this should eventually use Secure Enclave via platform channels
  Future<elliptic.PrivateKey> generateKeyPair() async {
    final ec = elliptic.getP256();
    final privateKey = ec.generatePrivateKey();
    return privateKey;
  }

  /// Extracts the public key in PEM format from an elliptic keypair
  /// Properly encodes the ECDSA public key in X.509 SubjectPublicKeyInfo format
  Future<String> getPublicKeyPEM(elliptic.PrivateKey privateKey) async {
    final publicKey = privateKey.publicKey;
    
    // Get the public key coordinates from PublicKey
    // Based on elliptic package API: PublicKey(ec, X, Y) where X and Y are BigInt
    // The properties are uppercase X and Y
    final x = (publicKey as dynamic).X as BigInt;
    final y = (publicKey as dynamic).Y as BigInt;
    
    // Convert BigInt to byte arrays (32 bytes each for P-256)
    final xBytes = _bigIntToBytes(x, 32);
    final yBytes = _bigIntToBytes(y, 32);
    
    // Uncompressed point format: 0x04 + x (32 bytes) + y (32 bytes) = 65 bytes
    final publicKeyBytes = Uint8List(65);
    publicKeyBytes[0] = 0x04;
    publicKeyBytes.setRange(1, 33, xBytes);
    publicKeyBytes.setRange(33, 65, yBytes);
    
    // Create SubjectPublicKeyInfo structure using asn1lib
    final algorithmOID = ASN1ObjectIdentifier([1, 2, 840, 10045, 2, 1]); // ecPublicKey
    final curveOID = ASN1ObjectIdentifier([1, 2, 840, 10045, 3, 1, 7]); // prime256v1
    final algorithmIdentifier = ASN1Sequence();
    algorithmIdentifier.add(algorithmOID);
    algorithmIdentifier.add(curveOID);
    
    final subjectPublicKeyInfo = ASN1Sequence();
    subjectPublicKeyInfo.add(algorithmIdentifier);
    subjectPublicKeyInfo.add(ASN1BitString(publicKeyBytes));
    
    final derBytes = subjectPublicKeyInfo.encodedBytes;
    final base64Key = base64Encode(derBytes);
    final lines = <String>[];
    for (int i = 0; i < base64Key.length; i += 64) {
      lines.add(base64Key.substring(i, (i + 64 < base64Key.length) ? i + 64 : base64Key.length));
    }
    return '-----BEGIN PUBLIC KEY-----\n${lines.join('\n')}\n-----END PUBLIC KEY-----';
  }

  /// Converts a BigInt to a byte array of specified length (pads with zeros if needed)
  Uint8List _bigIntToBytes(BigInt value, int length) {
    final bytes = <int>[];
    var temp = value;
    while (temp > BigInt.zero || bytes.length < length) {
      bytes.insert(0, (temp & BigInt.from(0xff)).toInt());
      temp = temp >> 8;
    }
    // Pad with zeros if needed
    while (bytes.length < length) {
      bytes.insert(0, 0);
    }
    // Trim to exact length (take last 'length' bytes)
    if (bytes.length > length) {
      return Uint8List.fromList(bytes.sublist(bytes.length - length));
    }
    return Uint8List.fromList(bytes);
  }

  /// Signs a challenge with the private key using ecdsa package
  /// Note: challenge is already a hash (32 bytes), so we sign it directly
  Future<Uint8List> signChallenge(Uint8List challenge, elliptic.PrivateKey privateKey) async {
    // The challenge is already a SHA-256 hash (32 bytes)
    // ECDSA signing expects the hash directly, not the raw message
    // Convert Uint8List to List<int> for ecdsa package
    final hash = challenge.toList();
    
    // Sign using ECDSA package
    final sig = ecdsa.signature(privateKey, hash);
    
    // The ecdsa package's Signature class structure is unclear
    // Try different methods to extract r and s
    final sigDynamic = sig as dynamic;
    
    BigInt r, s;
    
    // Try multiple approaches to access r and s
    // Approach 1: Check if signature is already bytes
    if (sigDynamic is List<int> || sigDynamic is Uint8List) {
      final sigBytes = sigDynamic is List<int> 
          ? Uint8List.fromList(sigDynamic) 
          : sigDynamic as Uint8List;
      if (sigBytes.length == 64) {
        return sigBytes; // Already in r||s format
      }
    }
    
    // Approach 2: Check if signature might be in ASN.1 DER format
    // ECDSA signatures are often returned in ASN.1 format which needs parsing
    try {
      // Try to get signature as bytes (might be ASN.1 encoded)
      final sigBytes = (sigDynamic as dynamic).toBytes?.call() as List<int>?;
      if (sigBytes != null) {
        // If it's 64 bytes, it's already in r||s format
        if (sigBytes.length == 64) {
          return Uint8List.fromList(sigBytes);
        }
        // If it's longer, it might be ASN.1 DER format - try to parse it
        if (sigBytes.length > 64) {
          // Parse ASN.1 DER format signature
          return _parseASN1Signature(sigBytes);
        }
      }
    } catch (_) {
      // Continue to next approach
    }
    
    // Approach 3: Access R and S properties (capitalized as per ecdsa package documentation)
    // The Signature class has R and S properties (getter/setter pair)
    r = sigDynamic.R as BigInt;
    s = sigDynamic.S as BigInt;
    
    // Convert to 64 bytes: r (32 bytes) || s (32 bytes)
    final rBytes = _bigIntToBytes(r, 32);
    final sBytes = _bigIntToBytes(s, 32);
    
    // Concatenate r and s
    final signatureBytes = Uint8List(64);
    signatureBytes.setRange(0, 32, rBytes);
    signatureBytes.setRange(32, 64, sBytes);
    
    return signatureBytes;
  }

  /// Creates a challenge based on user ID and public key
  Future<Uint8List> createChallenge(String userId, String publicKeyPEM) async {
    final data = '$userId:$publicKeyPEM';
    return _sha256(utf8.encode(data));
  }
  
  /// Simple SHA-256 hash implementation using crypto package
  Uint8List _sha256(List<int> data) {
    final digest = sha256.convert(data);
    return Uint8List.fromList(digest.bytes);
  }
  
  /// Parses an ASN.1 DER encoded ECDSA signature to extract r and s
  /// Returns 64 bytes: r (32 bytes) || s (32 bytes)
  Uint8List _parseASN1Signature(List<int> derBytes) {
    try {
      // Parse ASN.1 DER format
      // Format: 0x30 [length] 0x02 [r_length] [r_bytes] 0x02 [s_length] [s_bytes]
      if (derBytes.isEmpty || derBytes[0] != 0x30) {
        throw Exception('Invalid ASN.1 signature format');
      }
      
      int pos = 2; // Skip 0x30 and length byte
      
      // Parse r
      if (derBytes[pos] != 0x02) {
        throw Exception('Invalid ASN.1 signature: expected INTEGER for r');
      }
      pos++;
      final rLength = derBytes[pos];
      pos++;
      var rBytes = derBytes.sublist(pos, pos + rLength);
      // Remove leading zero if present (ASN.1 INTEGER can have leading zeros)
      if (rBytes.isNotEmpty && rBytes[0] == 0x00 && rBytes.length > 32) {
        rBytes = rBytes.sublist(1);
      }
      pos += rLength;
      
      // Parse s
      if (derBytes[pos] != 0x02) {
        throw Exception('Invalid ASN.1 signature: expected INTEGER for s');
      }
      pos++;
      final sLength = derBytes[pos];
      pos++;
      var sBytes = derBytes.sublist(pos, pos + sLength);
      // Remove leading zero if present
      if (sBytes.isNotEmpty && sBytes[0] == 0x00 && sBytes.length > 32) {
        sBytes = sBytes.sublist(1);
      }
      
      // Convert to BigInt and back to 32-byte arrays
      final r = _bytesToBigInt(rBytes);
      final s = _bytesToBigInt(sBytes);
      
      final rBytes32 = _bigIntToBytes(r, 32);
      final sBytes32 = _bigIntToBytes(s, 32);
      
      // Concatenate
      final signatureBytes = Uint8List(64);
      signatureBytes.setRange(0, 32, rBytes32);
      signatureBytes.setRange(32, 64, sBytes32);
      
      return signatureBytes;
    } catch (e) {
      throw Exception('Failed to parse ASN.1 signature: $e');
    }
  }
  
  /// Converts a byte array to BigInt
  BigInt _bytesToBigInt(List<int> bytes) {
    BigInt result = BigInt.zero;
    for (int i = 0; i < bytes.length; i++) {
      result = result << 8;
      result = result | BigInt.from(bytes[i] & 0xff);
    }
    return result;
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
      // Generate keypair using elliptic package
      final privateKey = await generateKeyPair();
      
      // Extract public key in PEM format
      final publicKeyPEM = await getPublicKeyPEM(privateKey);
      
      // Create challenge
      final challenge = await createChallenge(user.id!, publicKeyPEM);
      
      // Sign challenge
      final signature = await signChallenge(challenge, privateKey);
      
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
