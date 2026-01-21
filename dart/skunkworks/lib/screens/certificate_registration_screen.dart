import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/models/certificate.dart';
import 'package:skunkworks/providers/auth_provider.dart';
import 'package:skunkworks/providers/certificate_provider.dart';

class CertificateRegistrationScreen extends StatefulWidget {
  const CertificateRegistrationScreen({super.key});

  @override
  State<CertificateRegistrationScreen> createState() =>
      _CertificateRegistrationScreenState();
}

class _CertificateRegistrationScreenState
    extends State<CertificateRegistrationScreen> {
  @override
  void initState() {
    super.initState();
    // Don't check here - main.dart handles the initial check
    // This prevents duplicate checks and infinite loops
  }

  Future<void> _handleEnrollCertificate() async {
    final authProvider = context.read<AuthProvider>();
    final certProvider = context.read<CertificateProvider>();

    if (authProvider.user == null) {
      return;
    }

    try {
      await certProvider.enrollCertificate(authProvider.user!);
      // Success - the provider will update state
      if (mounted && certProvider.hasCertificate) {
        // Navigate back or show success message
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Certificate generated successfully!'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (e) {
      // Error is handled by CertificateProvider and displayed via error getter
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to generate certificate: ${certProvider.error ?? e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        title: const Text('Certificate Registration'),
        backgroundColor: Colors.white,
        elevation: 0,
      ),
      body: SafeArea(
        child: Consumer<CertificateProvider>(
          builder: (context, certProvider, child) {
            // If already has certificate, show success state
            if (certProvider.hasCertificate && certProvider.certificate != null) {
              return _buildSuccessState(certProvider.certificate!);
            }

            // Loading state
            if (certProvider.loading) {
              return const Center(
                child: CircularProgressIndicator(),
              );
            }

            // Registration form
            return SingleChildScrollView(
              padding: const EdgeInsets.all(24.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const SizedBox(height: 40),
                  // Title
                  const Text(
                    'Generate Certificate',
                    style: TextStyle(
                      fontSize: 28,
                      fontWeight: FontWeight.bold,
                      color: Colors.black87,
                    ),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 16),
                  // Description
                  Text(
                    'A certificate is required to sign and verify your content. '
                    'Your private key will be stored securely on your device.',
                    style: TextStyle(
                      fontSize: 16,
                      color: Colors.grey.shade600,
                    ),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 40),
                  // Error message
                  if (certProvider.error != null) ...[
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.red.shade50,
                        borderRadius: BorderRadius.circular(4),
                        border: Border.all(color: Colors.red.shade200),
                      ),
                      child: Text(
                        certProvider.error!,
                        style: TextStyle(
                          color: Colors.red.shade700,
                          fontSize: 14,
                        ),
                        textAlign: TextAlign.center,
                      ),
                    ),
                    const SizedBox(height: 20),
                  ],
                  // Generate button
                  SizedBox(
                    height: 44,
                    child: ElevatedButton(
                      onPressed: certProvider.loading ? null : _handleEnrollCertificate,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.blue.shade600,
                        foregroundColor: Colors.white,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(4),
                        ),
                        elevation: 0,
                      ),
                      child: const Text(
                        'Generate Certificate',
                        style: TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            );
          },
        ),
      ),
    );
  }

  Widget _buildSuccessState(Certificate certificate) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const SizedBox(height: 40),
          // Success icon
          Icon(
            Icons.check_circle,
            size: 80,
            color: Colors.green.shade600,
          ),
          const SizedBox(height: 24),
          // Success message
          const Text(
            'Certificate Generated',
            style: TextStyle(
              fontSize: 28,
              fontWeight: FontWeight.bold,
              color: Colors.black87,
            ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 16),
          Text(
            'Your certificate has been successfully generated and stored.',
            style: TextStyle(
              fontSize: 16,
              color: Colors.grey.shade600,
            ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 40),
          // Certificate info
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.grey.shade50,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: Colors.grey.shade200),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Fingerprint',
                  style: TextStyle(
                    fontSize: 12,
                    color: Colors.grey.shade600,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  certificate.fingerprint,
                  style: const TextStyle(
                    fontSize: 14,
                    fontFamily: 'monospace',
                    color: Colors.black87,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
