import 'dart:convert';
import 'dart:typed_data';
import 'package:cbor/cbor.dart';
import 'package:crypto/crypto.dart';
import 'package:skunkworks/services/certificate_service.dart';

/// Service for creating and managing C2PA manifests
class C2PAService {
  final CertificateService _certificateService;

  C2PAService(this._certificateService);

  /// Creates a C2PA manifest for an image
  /// 
  /// [imageUrl] - URL of the image being signed
  /// [assetId] - Optional asset identifier
  /// 
  /// Returns a map containing:
  /// - manifestBytes: The CBOR-encoded manifest
  /// - manifestHash: SHA-256 hash of the manifest bytes
  /// - certFingerprint: Certificate fingerprint used for signing
  Future<Map<String, dynamic>> createManifest(
    String imageUrl, {
    String? assetId,
  }) async {
    // Get certificate from local storage
    final certificate = await _certificateService.getCertificateLocally();
    if (certificate == null) {
      throw Exception('Certificate not found. Please enroll a certificate first.');
    }

    // Get certificate fingerprint (hex string to bytes)
    final fingerprintBytes = _hexToBytes(certificate.fingerprint);

    // Create C2PA manifest structure
    // Simplified C2PA manifest with required fields
    final manifest = _createManifestStructure(
      imageUrl: imageUrl,
      assetId: assetId,
      certificatePEM: certificate.certificatePem,
      fingerprint: fingerprintBytes,
    );

    // Encode manifest as CBOR
    final encoder = CborEncoder();
    final manifestBytesList = encoder.convert(manifest);
    final manifestBytes = Uint8List.fromList(manifestBytesList);

    // Compute manifest hash
    final manifestHash = sha256.convert(manifestBytes).bytes;

    // Sign the manifest (we'll sign the hash)
    // Note: In a full implementation, we'd sign the manifest properly
    // For now, we'll create a signature assertion in the manifest structure
    final signedManifest = _addSignatureToManifest(
      manifest,
      manifestHash,
      certificate.certificatePem,
    );

    // Re-encode with signature
    final finalEncoder = CborEncoder();
    final finalManifestBytesList = finalEncoder.convert(signedManifest);
    final finalManifestBytes = Uint8List.fromList(finalManifestBytesList);

    // Recompute hash with signature
    final finalManifestHash = sha256.convert(finalManifestBytes).bytes;

    return {
      'manifestBytes': finalManifestBytes,
      'manifestHash': Uint8List.fromList(finalManifestHash),
      'certFingerprint': fingerprintBytes,
    };
  }

  /// Creates the basic C2PA manifest structure
  Map<String, dynamic> _createManifestStructure({
    required String imageUrl,
    String? assetId,
    required String certificatePEM,
    required Uint8List fingerprint,
  }) {
    final now = DateTime.now().toUtc();
    final timestamp = now.toIso8601String();

    // C2PA claim structure
    final claim = {
      'alg': 'sha256',
      'claim_generator': 'verifiable-sn/1.0.0',
      'claim_generator_info': [
        {
          'name': 'verifiable-sn',
          'version': '1.0.0',
        }
      ],
      'dc:format': 'image/jpeg', // Default, should be determined from image
      'dc:title': assetId ?? imageUrl,
      'dc:creator': [
        {
          '@type': 'Person',
        }
      ],
      'dc:dateCreated': timestamp,
      'dc:dateModified': timestamp,
    };

    // Actions assertion - required by C2PA spec
    final actionsAssertion = {
      'label': 'c2pa.actions',
      'data': {
        'actions': [
          {
            'action': 'c2pa.created',
            'when': timestamp,
            'softwareAgent': {
              'name': 'verifiable-sn',
              'version': '1.0.0',
            }
          }
        ]
      }
    };

    // Basic manifest structure
    final manifest = {
      'claim': claim,
      'assertions': [
        actionsAssertion,
      ],
      'signatures': [], // Will be populated by _addSignatureToManifest
    };

    return manifest;
  }

  /// Adds signature assertion to the manifest
  Map<String, dynamic> _addSignatureToManifest(
    Map<String, dynamic> manifest,
    Uint8List manifestHash,
    String certificatePEM,
  ) {
    // Extract certificate DER from PEM
    final certDER = _pemToDER(certificatePEM);

    // Create signature assertion
    // Note: In a full implementation, we would actually sign the manifest
    // with the private key. For now, we'll create the structure.
    final signatureAssertion = {
      'label': 'c2pa.signature',
      'data': {
        'alg': 'ES256', // ECDSA with SHA-256
        'cert_chain': [
          base64Encode(certDER),
        ],
        'signature': base64Encode(manifestHash), // Placeholder - should be actual signature
        'time': DateTime.now().toUtc().toIso8601String(),
      }
    };

    // Add signature to assertions
    final assertions = List<Map<String, dynamic>>.from(manifest['assertions'] as List);
    assertions.add(signatureAssertion);

    final signedManifest = Map<String, dynamic>.from(manifest);
    signedManifest['assertions'] = assertions;

    return signedManifest;
  }

  /// Converts hex string to bytes
  Uint8List _hexToBytes(String hex) {
    final cleanHex = hex.replaceAll(' ', '').replaceAll('0x', '');
    final bytes = <int>[];
    for (int i = 0; i < cleanHex.length; i += 2) {
      final byteStr = cleanHex.substring(i, i + 2);
      bytes.add(int.parse(byteStr, radix: 16));
    }
    return Uint8List.fromList(bytes);
  }

  /// Converts PEM certificate to DER bytes
  Uint8List _pemToDER(String pem) {
    // Remove PEM headers and whitespace
    final cleanPem = pem
        .replaceAll('-----BEGIN CERTIFICATE-----', '')
        .replaceAll('-----END CERTIFICATE-----', '')
        .replaceAll('\n', '')
        .replaceAll(' ', '');
    
    return base64Decode(cleanPem);
  }

  /// Gets certificate fingerprint from stored certificate
  Future<Uint8List> getCertFingerprint() async {
    final certificate = await _certificateService.getCertificateLocally();
    if (certificate == null) {
      throw Exception('Certificate not found');
    }
    return _hexToBytes(certificate.fingerprint);
  }

  /// Computes SHA-256 hash of manifest bytes
  Uint8List getManifestHash(Uint8List manifestBytes) {
    final digest = sha256.convert(manifestBytes);
    return Uint8List.fromList(digest.bytes);
  }
}
