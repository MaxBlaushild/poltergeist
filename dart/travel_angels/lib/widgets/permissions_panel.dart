import 'dart:async';

import 'package:app_links/app_links.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/google_drive_service.dart';
import 'package:url_launcher/url_launcher.dart';

/// Permissions panel widget for managing third-party integrations
class PermissionsPanel extends StatefulWidget {
  const PermissionsPanel({super.key});

  @override
  State<PermissionsPanel> createState() => _PermissionsPanelState();
}

class _PermissionsPanelState extends State<PermissionsPanel> {
  final GoogleDriveService _googleDriveService = GoogleDriveService(
    APIClient(ApiConstants.baseUrl),
  );
  final AppLinks _appLinks = AppLinks();
  StreamSubscription<Uri>? _linkSubscription;
  bool _isGoogleDriveConnected = false;
  bool _isLoading = true;
  bool _isToggling = false;
  bool _expectingOAuthCallback = false;

  @override
  void initState() {
    super.initState();
    _loadStatus();
    _initDeepLinkListener();
  }

  void _initDeepLinkListener() {
    // Listen for NEW deep links (stream events only, not initial link)
    // The initial link is handled once in main.dart when app starts
    _linkSubscription = _appLinks.uriLinkStream.listen(
      (Uri uri) {
        _handleDeepLink(uri);
      },
      onError: (err) {
        // Handle error silently or log it
        debugPrint('Deep link error: $err');
      },
    );
  }

  void _handleDeepLink(Uri uri) {
    if (uri.scheme == 'travelangels' && uri.host == 'oauth-callback') {
      final success = uri.queryParameters['success'];
      final service = uri.queryParameters['service'];

      if (success == 'true' && service == 'google-drive') {
        // Refresh status to check if connection was established
        final wasConnected = _isGoogleDriveConnected;
        _loadStatus();
        
        // Only show success message if we were expecting the callback
        // (i.e., we initiated the OAuth flow) or if status changed from disconnected to connected
        if (_expectingOAuthCallback) {
          _expectingOAuthCallback = false;
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('Google Drive connected successfully!'),
                backgroundColor: Colors.green,
              ),
            );
          }
        } else if (!wasConnected) {
          // Status changed from disconnected to connected (e.g., app opened via deep link)
          // Refresh silently without showing message
        }
      }
    }
  }

  @override
  void dispose() {
    _linkSubscription?.cancel();
    super.dispose();
  }

  Future<void> _loadStatus() async {
    try {
      setState(() {
        _isLoading = true;
      });
      final status = await _googleDriveService.getStatus();
      setState(() {
        _isGoogleDriveConnected = status['connected'] ?? false;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
      });
      if (mounted) {
        String errorMessage = 'Failed to load permissions status';
        if (e is DioException) {
          if (e.response != null) {
            errorMessage = '${errorMessage}: ${e.response?.statusCode} - ${e.response?.statusMessage}';
            if (e.response?.data != null && e.response?.data is Map) {
              final errorData = e.response?.data as Map<String, dynamic>;
              errorMessage = errorData['error']?.toString() ?? errorMessage;
            }
          } else {
            errorMessage = '${errorMessage}: ${e.message ?? e.toString()}';
          }
        } else {
          errorMessage = '$errorMessage: $e';
        }
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(errorMessage),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  Future<void> _handleGoogleDriveToggle(bool value) async {
    if (_isToggling) return;

    setState(() {
      _isToggling = true;
    });

    try {
      if (value) {
        // Enable: Get auth URL and open in browser
        final authResponse = await _googleDriveService.getAuthUrl();
        final authUrl = authResponse['authUrl'] as String?;

        if (authUrl == null || authUrl.isEmpty) {
          throw Exception('No auth URL received');
        }

        // Open auth URL in browser
        final uri = Uri.parse(authUrl);
        if (await canLaunchUrl(uri)) {
          await launchUrl(uri, mode: LaunchMode.externalApplication);
          // Set flag to expect OAuth callback
          setState(() {
            _expectingOAuthCallback = true;
          });
          // Show message that user should complete OAuth flow
          // The deep link listener will handle the callback and refresh status
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text(
                  'Please complete the Google Drive authorization in your browser. '
                  'You will be redirected back to the app automatically.',
                ),
                duration: Duration(seconds: 5),
              ),
            );
          }
        } else {
          throw Exception('Could not launch URL: $authUrl');
        }
      } else {
        // Disable: Revoke access
        await _googleDriveService.revoke();
        setState(() {
          _isGoogleDriveConnected = false;
        });
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Google Drive access revoked successfully'),
              backgroundColor: Colors.green,
            ),
          );
        }
      }
    } catch (e) {
      String errorMessage = 'Failed to ${value ? 'connect' : 'disconnect'} Google Drive';
      if (e is DioException) {
        if (e.response != null) {
          errorMessage = '${errorMessage}: ${e.response?.statusCode} - ${e.response?.statusMessage}';
          if (e.response?.data != null && e.response?.data is Map) {
            final errorData = e.response?.data as Map<String, dynamic>;
            errorMessage = errorData['error']?.toString() ?? errorMessage;
          }
        } else {
          errorMessage = '${errorMessage}: ${e.message ?? e.toString()}';
        }
      } else {
        errorMessage = '$errorMessage: $e';
      }
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(errorMessage),
            backgroundColor: Colors.red,
          ),
        );
      }
      // Revert toggle state on error
      setState(() {
        _isGoogleDriveConnected = !value;
        _expectingOAuthCallback = false; // Reset flag on error
      });
    } finally {
      setState(() {
        _isToggling = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Permissions',
              style: theme.textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 16),
            if (_isLoading)
              const Center(
                child: Padding(
                  padding: EdgeInsets.all(16.0),
                  child: CircularProgressIndicator(),
                ),
              )
            else
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Row(
                    children: [
                      Icon(
                        Icons.cloud,
                        color: theme.colorScheme.primary,
                      ),
                      const SizedBox(width: 12),
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Google Drive',
                            style: theme.textTheme.bodyLarge?.copyWith(
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                          Text(
                            _isGoogleDriveConnected
                                ? 'Connected'
                                : 'Not connected',
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: _isGoogleDriveConnected
                                  ? Colors.green
                                  : theme.colorScheme.onSurface.withOpacity(0.6),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                  Switch(
                    value: _isGoogleDriveConnected,
                    onChanged: _isToggling ? null : _handleGoogleDriveToggle,
                  ),
                ],
              ),
          ],
        ),
      ),
    );
  }
}

