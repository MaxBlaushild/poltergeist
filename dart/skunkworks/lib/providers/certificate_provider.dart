import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:skunkworks/models/certificate.dart';
import 'package:skunkworks/models/user.dart';
import 'package:skunkworks/services/certificate_service.dart';

class CertificateProvider extends ChangeNotifier {
  final CertificateService _certificateService;
  bool _hasCertificate = false;
  Certificate? _certificate;
  bool _loading = false;
  String? _error;

  CertificateProvider(this._certificateService);

  bool get hasCertificate => _hasCertificate;
  Certificate? get certificate => _certificate;
  bool get loading => _loading;
  String? get error => _error;

  /// Checks if the user has a certificate
  Future<void> checkCertificate() async {
    _loading = true;
    _error = null;
    notifyListeners();

    try {
      _hasCertificate = await _certificateService.hasCertificate();
      
      // If certificate exists, try to fetch it
      if (_hasCertificate) {
        _certificate = await _certificateService.getCertificate();
        // Also try to get from local storage as fallback
        if (_certificate == null) {
          _certificate = await _certificateService.getCertificateLocally();
        }
      }
    } catch (e) {
      _error = _extractErrorMessage(e);
      // On error, assume no certificate
      _hasCertificate = false;
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  /// Enrolls a certificate for the user
  Future<void> enrollCertificate(User user) async {
    _loading = true;
    _error = null;
    notifyListeners();

    try {
      _certificate = await _certificateService.enrollCertificate(user);
      _hasCertificate = true;
    } catch (e) {
      _error = _extractErrorMessage(e);
      rethrow;
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  /// Clears the certificate state
  void clearCertificate() {
    _hasCertificate = false;
    _certificate = null;
    _error = null;
    notifyListeners();
  }

  /// Extracts error message from DioException response body
  String _extractErrorMessage(dynamic error) {
    if (error is DioException) {
      // Try to get error message from response body
      if (error.response?.data != null) {
        final data = error.response!.data;
        if (data is Map<String, dynamic> && data.containsKey('error')) {
          return data['error'] as String;
        }
        if (data is String) {
          return data;
        }
      }
      // Fall back to DioException message
      return error.message ?? 'An error occurred';
    }
    // Fall back for non-Dio errors
    return error.toString();
  }
}
