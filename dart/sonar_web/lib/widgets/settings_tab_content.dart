import 'package:flutter/material.dart';

import '../services/notification_permission_service.dart';

class SettingsTabContent extends StatefulWidget {
  const SettingsTabContent({super.key});

  @override
  State<SettingsTabContent> createState() => _SettingsTabContentState();
}

class _SettingsTabContentState extends State<SettingsTabContent> {
  final NotificationPermissionService _notificationPermissionService =
      NotificationPermissionService();

  NotificationPermissionState _permissionState =
      NotificationPermissionState.notDetermined;
  bool _loading = true;
  bool _requesting = false;

  @override
  void initState() {
    super.initState();
    _loadPermissionState();
  }

  Future<void> _loadPermissionState() async {
    setState(() => _loading = true);
    try {
      final state = await _notificationPermissionService.getPermissionState();
      if (!mounted) return;
      setState(() {
        _permissionState = state;
        _loading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _loading = false);
    }
  }

  Future<void> _onTogglePush(bool enabled) async {
    if (!enabled) {
      if (_permissionState == NotificationPermissionState.granted && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text(
              'Push is enabled. Disable it from browser/system settings.',
            ),
          ),
        );
      }
      return;
    }

    setState(() => _requesting = true);
    try {
      final state = await _notificationPermissionService.requestPermission();
      if (!mounted) return;
      setState(() {
        _permissionState = state;
        _requesting = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _requesting = false);
    }
  }

  String _statusLabel() {
    switch (_permissionState) {
      case NotificationPermissionState.granted:
        return 'Enabled';
      case NotificationPermissionState.denied:
        return 'Blocked';
      case NotificationPermissionState.unsupported:
        return 'Unsupported on this browser';
      case NotificationPermissionState.notDetermined:
        return 'Not enabled';
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    final pushEnabled = _permissionState == NotificationPermissionState.granted;
    final canRequest =
        _permissionState != NotificationPermissionState.unsupported;

    return SingleChildScrollView(
      primary: false,
      padding: const EdgeInsets.only(bottom: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            'Settings',
            style: theme.textTheme.headlineSmall?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Control app preferences and notifications.',
            style: theme.textTheme.bodyMedium?.copyWith(
              color: scheme.onSurfaceVariant,
            ),
          ),
          const SizedBox(height: 16),
          Material(
            color: scheme.surface,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(16),
              side: BorderSide(color: scheme.outlineVariant),
            ),
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    'Notifications',
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 10),
                  Container(
                    decoration: BoxDecoration(
                      color: scheme.surfaceContainerHighest,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: scheme.outlineVariant),
                    ),
                    child: SwitchListTile(
                      value: pushEnabled,
                      onChanged: (_loading || _requesting || !canRequest)
                          ? null
                          : _onTogglePush,
                      title: const Text('Allow push notifications'),
                      subtitle: Text(_statusLabel()),
                      secondary: (_loading || _requesting)
                          ? const SizedBox(
                              width: 18,
                              height: 18,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Icon(Icons.notifications_active_outlined),
                    ),
                  ),
                  if (_permissionState == NotificationPermissionState.denied)
                    Padding(
                      padding: const EdgeInsets.only(top: 10),
                      child: Text(
                        'Push is blocked right now. Enable it in browser settings and then refresh status.',
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: scheme.onSurfaceVariant,
                        ),
                      ),
                    ),
                  const SizedBox(height: 12),
                  Align(
                    alignment: Alignment.centerLeft,
                    child: TextButton.icon(
                      onPressed: _loading || _requesting
                          ? null
                          : _loadPermissionState,
                      icon: const Icon(Icons.refresh),
                      label: const Text('Refresh status'),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
