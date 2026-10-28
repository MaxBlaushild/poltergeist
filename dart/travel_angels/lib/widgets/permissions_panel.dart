import 'dart:async';

import 'package:app_links/app_links.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/google_drive_service.dart';
import 'package:url_launcher/url_launcher.dart';

/// A connection is shown as active only after the server confirms its status.
class PermissionsPanel extends StatefulWidget {
  const PermissionsPanel({
    super.key,
    this.service,
    this.deepLinkStream,
    this.launchAuthorization,
    this.listenForDeepLinks = true,
  });

  final GoogleDriveService? service;
  final Stream<Uri>? deepLinkStream;
  final Future<bool> Function(Uri)? launchAuthorization;
  final bool listenForDeepLinks;

  @override
  State<PermissionsPanel> createState() => _PermissionsPanelState();
}

class _PermissionsPanelState extends State<PermissionsPanel>
    with WidgetsBindingObserver {
  late final GoogleDriveService _service =
      widget.service ?? GoogleDriveService(APIClient(ApiConstants.baseUrl));
  StreamSubscription<Uri>? _linkSubscription;
  AuthProvider? _auth;
  String? _identity;
  bool _started = false;
  bool? _connected;
  bool _loading = true;
  bool _toggling = false;
  bool _awaitingAuthorization = false;
  String? _error;
  int _statusGeneration = 0;
  int _actionGeneration = 0;
  bool get _canLoad => _auth == null || _identity?.isNotEmpty == true;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _listenForLinks();
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final auth = context.read<AuthProvider?>();
    if (auth != _auth) {
      _auth?.removeListener(_authChanged);
      _auth = auth;
      _auth?.addListener(_authChanged);
    }
    if (!_started) {
      _started = true;
      _identity = _auth?.user?.id;
      if (_canLoad) {
        _loadStatus();
      } else {
        _loading = false;
      }
    } else {
      _authChanged();
    }
  }

  void _authChanged() {
    if (!mounted) return;
    final identity = _auth?.user?.id;
    if (identity == _identity) return;
    _identity = identity;
    _statusGeneration++;
    _actionGeneration++;
    setState(() {
      _connected = null;
      _loading = false;
      _toggling = false;
      _awaitingAuthorization = false;
      _error = null;
    });
    if (_canLoad) _loadStatus();
  }

  void _listenForLinks() {
    if (!widget.listenForDeepLinks) return;
    // Browser OAuth does not return through the native app-links plugin.
    // App-resume checks also work when no deep link is delivered.
    if (widget.deepLinkStream == null && kIsWeb) return;
    try {
      _linkSubscription = (widget.deepLinkStream ?? AppLinks().uriLinkStream)
          .listen(_handleDeepLink, onError: (Object _) {});
    } catch (_) {
      // Some host platforms have no link listener. App-resume checks still work.
    }
  }

  void _handleDeepLink(Uri uri) {
    if (!mounted ||
        uri.scheme != 'travelangels' ||
        uri.host != 'oauth-callback' ||
        uri.queryParameters['service'] != 'google-drive') {
      return;
    }
    if (uri.queryParameters['success'] == 'true') {
      _loadStatus();
    } else if (_awaitingAuthorization) {
      setState(() {
        _awaitingAuthorization = false;
        _error =
            'Google Drive authorization was not completed. Try connecting again.';
      });
    }
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed && mounted) _loadStatus();
  }

  @override
  void dispose() {
    _statusGeneration++;
    _actionGeneration++;
    WidgetsBinding.instance.removeObserver(this);
    _auth?.removeListener(_authChanged);
    _linkSubscription?.cancel();
    super.dispose();
  }

  Future<void> _loadStatus() async {
    if (!mounted || !_canLoad) return;
    final generation = ++_statusGeneration;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final status = await _service.getStatus();
      if (!mounted || generation != _statusGeneration) return;
      final connected = status['connected'];
      if (connected is! bool) {
        throw const FormatException('Missing connection status');
      }
      setState(() {
        _connected = connected;
        _loading = false;
        if (connected) _awaitingAuthorization = false;
      });
    } catch (_) {
      if (!mounted || generation != _statusGeneration) return;
      setState(() {
        _loading = false;
        _error = 'Google Drive status could not be checked. Please try again.';
      });
    }
  }

  Future<void> _toggle(bool connect) async {
    if (!_canLoad || _toggling || _loading || _connected == null) return;
    final action = ++_actionGeneration;
    _statusGeneration++;
    setState(() {
      _toggling = true;
      _error = null;
    });
    try {
      if (connect) {
        final response = await _service.getAuthUrl();
        if (!mounted || action != _actionGeneration) return;
        final uri = Uri.tryParse(response['authUrl']?.toString() ?? '');
        if (uri == null || uri.scheme != 'https' || uri.host.isEmpty) {
          throw const FormatException('Invalid authorization URL');
        }
        setState(() => _awaitingAuthorization = true);
        final launched =
            await (widget.launchAuthorization?.call(uri) ??
                launchUrl(uri, mode: LaunchMode.externalApplication));
        if (!mounted || action != _actionGeneration) return;
        if (!launched) throw StateError('Authorization could not be opened');
      } else {
        await _service.revoke();
        if (!mounted || action != _actionGeneration) return;
        setState(() {
          _connected = null;
          _awaitingAuthorization = false;
        });
        await _loadStatus();
      }
    } catch (_) {
      if (!mounted || action != _actionGeneration) return;
      setState(() {
        _awaitingAuthorization = false;
        _error = connect
            ? 'Google Drive authorization could not be opened. Please try again.'
            : 'Google Drive could not be disconnected. Please try again.';
      });
    } finally {
      if (mounted && action == _actionGeneration) {
        setState(() => _toggling = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final status = !_canLoad
        ? 'Sign in to manage connected accounts.'
        : _loading
        ? 'Checking connection…'
        : _connected == true
        ? 'Connected'
        : _connected == false
        ? 'Not connected'
        : 'Connection status unavailable';
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              children: [
                Icon(
                  Icons.add_to_drive_outlined,
                  color: theme.colorScheme.primary,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Google Drive', style: theme.textTheme.titleLarge),
                      const SizedBox(height: 4),
                      Semantics(
                        liveRegion: true,
                        child: Text(
                          status,
                          key: const ValueKey('drive-connection-status'),
                          style: theme.textTheme.bodyMedium?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 8),
                Semantics(
                  label: 'Google Drive connection',
                  child: Switch(
                    key: const ValueKey('drive-connection-toggle'),
                    value: _connected ?? false,
                    onChanged: _loading || _toggling || _connected == null
                        ? null
                        : _toggle,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            const Text(
              'Bring Google Docs and Sheets into your travel library.',
            ),
            if (_loading || _toggling) ...[
              const SizedBox(height: 16),
              const LinearProgressIndicator(minHeight: 2),
            ],
            if (_awaitingAuthorization) ...[
              const SizedBox(height: 16),
              const Text(
                'Complete authorization in your browser, then return here.',
              ),
            ],
            if (_error != null) ...[
              const SizedBox(height: 16),
              Text(
                _error!,
                key: const ValueKey('drive-connection-error'),
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: theme.colorScheme.error,
                ),
              ),
              const SizedBox(height: 12),
              Align(
                alignment: Alignment.centerLeft,
                child: TextButton.icon(
                  key: const ValueKey('drive-connection-retry'),
                  onPressed: !_canLoad || _loading || _toggling
                      ? null
                      : _loadStatus,
                  icon: const Icon(Icons.replay, size: 18),
                  label: const Text('Retry'),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
