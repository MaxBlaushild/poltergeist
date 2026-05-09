import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:dio/dio.dart';
import 'package:go_router/go_router.dart';

import '../providers/auth_provider.dart';
import '../providers/map_visual_settings_provider.dart';
import '../providers/user_level_provider.dart';
import '../services/notification_permission_service.dart';
import '../services/push_notification_service.dart';
import '../services/poi_service.dart';

class SettingsTabContent extends StatefulWidget {
  const SettingsTabContent({super.key, this.onTriggerTutorialTest});

  final VoidCallback? onTriggerTutorialTest;

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
  bool _sendingTestPush = false;
  bool _spawningNearbyContent = false;
  bool _triggeringTutorial = false;

  void _handleLogout() {
    final router = GoRouter.of(context);
    Navigator.of(context).pop();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      router.go('/logout');
    });
  }

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
      if (state == NotificationPermissionState.granted) {
        await _syncPushRegistration(force: false);
      }
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
      if (state == NotificationPermissionState.granted) {
        await _syncPushRegistration(force: true);
      }
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

  Future<void> _syncPushRegistration({required bool force}) async {
    final userId = context.read<AuthProvider>().user?.id;
    if (userId == null || userId.isEmpty) return;
    try {
      await context.read<PushNotificationService>().registerDeviceTokenForUser(
        userId,
        force: force,
      );
    } catch (_) {}
  }

  String _pushErrorMessage(Object error) {
    if (error is DioException) {
      final data = error.response?.data;
      if (data is Map && data['error'] is String) {
        return data['error'] as String;
      }
      if (error.message != null && error.message!.trim().isNotEmpty) {
        return error.message!.trim();
      }
    }
    return 'Failed to send test push notification.';
  }

  String _spawnErrorMessage(Object error) {
    if (error is DioException) {
      final data = error.response?.data;
      if (data is Map && data['error'] is String) {
        return data['error'] as String;
      }
      if (error.message != null && error.message!.trim().isNotEmpty) {
        return error.message!.trim();
      }
    }
    return 'Failed to generate nearby scenario and monster encounter.';
  }

  String _formatLevelOffset(int offset) {
    if (offset > 0) {
      return '+$offset';
    }
    if (offset < 0) {
      return '$offset';
    }
    return '0';
  }

  String _contentDifficultySubtitle(int offset, int? userLevel) {
    if (userLevel == null || userLevel <= 0) {
      if (offset == 0) {
        return 'Scale content to match your current level.';
      }
      return 'Scale content ${_formatLevelOffset(offset)} levels from your current level.';
    }
    final effectiveLevel = userLevel + offset <= 1 ? 1 : userLevel + offset;
    if (offset == 0) {
      return 'Content currently matches your level $userLevel.';
    }
    return 'Content scales as if you were level $effectiveLevel (${_formatLevelOffset(offset)} from level $userLevel).';
  }

  String _shortId(dynamic value) {
    final raw = (value ?? '').toString().trim();
    if (raw.isEmpty) return '';
    if (raw.length <= 8) return raw;
    return raw.substring(0, 8);
  }

  Future<void> _spawnNearbyScenarioAndMonster() async {
    setState(() => _spawningNearbyContent = true);
    try {
      final result = await context
          .read<PoiService>()
          .spawnNearbyScenarioAndMonster();
      if (!mounted) return;
      final zoneName = (result['zoneName'] ?? '').toString().trim();
      final scenario = result['scenario'];
      final encounter = result['monsterEncounter'];
      final scenarioId = scenario is Map ? _shortId(scenario['id']) : '';
      final encounterId = encounter is Map ? _shortId(encounter['id']) : '';
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Spawned nearby content${zoneName.isNotEmpty ? ' in $zoneName' : ''}'
            '${scenarioId.isNotEmpty ? ' • Scenario $scenarioId' : ''}'
            '${encounterId.isNotEmpty ? ' • Encounter $encounterId' : ''}.',
          ),
        ),
      );
    } catch (error) {
      if (!mounted) return;
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(_spawnErrorMessage(error))));
    } finally {
      if (mounted) {
        setState(() => _spawningNearbyContent = false);
      }
    }
  }

  Future<void> _sendTestPush({int delaySeconds = 0}) async {
    if (_permissionState != NotificationPermissionState.granted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Enable notifications first.')),
      );
      return;
    }

    setState(() => _sendingTestPush = true);
    final pushNotificationService = context.read<PushNotificationService>();
    try {
      await _syncPushRegistration(force: false);
      if (delaySeconds > 0 && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Sending test push in $delaySeconds seconds. Background the app now.',
            ),
          ),
        );
      }
      final result = await pushNotificationService.sendTestPush(
        delaySeconds: delaySeconds,
      );
      if (!mounted) return;
      final sent = (result['sent'] as num?)?.toInt() ?? 0;
      final failed = (result['failed'] as num?)?.toInt() ?? 0;
      final tokens = (result['tokens'] as num?)?.toInt();
      final total = tokens ?? (sent + failed);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Test push sent to $sent/$total token(s).${failed > 0 ? ' $failed failed.' : ''}',
          ),
        ),
      );
    } catch (error) {
      if (!mounted) return;
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(_pushErrorMessage(error))));
    } finally {
      if (mounted) {
        setState(() => _sendingTestPush = false);
      }
    }
  }

  Future<void> _triggerTutorialTest() async {
    setState(() => _triggeringTutorial = true);
    try {
      widget.onTriggerTutorialTest?.call();
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Replaying tutorial on single player.')),
      );
    } finally {
      if (mounted) {
        setState(() => _triggeringTutorial = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    final pushEnabled = _permissionState == NotificationPermissionState.granted;
    final canRequest =
        _permissionState != NotificationPermissionState.unsupported;
    final mapVisualSettings = context.watch<MapVisualSettingsProvider>();
    final userLevel = context.watch<UserLevelProvider>().userLevel?.level;
    final contentLevelOffset = mapVisualSettings.contentLevelOffset;

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
                        'Push is blocked right now. Enable it in browser settings.',
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: scheme.onSurfaceVariant,
                        ),
                      ),
                    ),
                  const SizedBox(height: 12),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: [
                      FilledButton.tonalIcon(
                        onPressed:
                            (_loading ||
                                _requesting ||
                                _sendingTestPush ||
                                _spawningNearbyContent ||
                                _triggeringTutorial)
                            ? null
                            : () => _sendTestPush(),
                        icon: _sendingTestPush
                            ? const SizedBox(
                                width: 16,
                                height: 16,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : const Icon(Icons.send_rounded),
                        label: const Text('Send test push now'),
                      ),
                      OutlinedButton.icon(
                        onPressed:
                            (_loading ||
                                _requesting ||
                                _sendingTestPush ||
                                _spawningNearbyContent ||
                                _triggeringTutorial)
                            ? null
                            : () => _sendTestPush(delaySeconds: 10),
                        icon: const Icon(Icons.timer_outlined),
                        label: const Text('Send test push in 10s'),
                      ),
                      FilledButton.icon(
                        onPressed:
                            (_loading ||
                                _requesting ||
                                _sendingTestPush ||
                                _spawningNearbyContent ||
                                _triggeringTutorial)
                            ? null
                            : _spawnNearbyScenarioAndMonster,
                        icon: _spawningNearbyContent
                            ? const SizedBox(
                                width: 16,
                                height: 16,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : const Icon(Icons.auto_awesome_outlined),
                        label: const Text('Generate nearby scenario + monster'),
                      ),
                      OutlinedButton.icon(
                        onPressed:
                            (_loading ||
                                _requesting ||
                                _sendingTestPush ||
                                _spawningNearbyContent ||
                                _triggeringTutorial)
                            ? null
                            : _triggerTutorialTest,
                        icon: _triggeringTutorial
                            ? const SizedBox(
                                width: 16,
                                height: 16,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : const Icon(Icons.menu_book_outlined),
                        label: const Text('Replay tutorial'),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  const Divider(),
                  const SizedBox(height: 12),
                  Text(
                    'Map',
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
                      value: mapVisualSettings.zoneKindMapStylingEnabled,
                      onChanged: mapVisualSettings.setZoneKindMapStylingEnabled,
                      title: const Text('Show discovered zone kind styling'),
                      subtitle: const Text(
                        'Adds zone-kind tinting and selected-zone tile textures to discovered zones on the map.',
                      ),
                      secondary: const Icon(Icons.layers_outlined),
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
                      value: mapVisualSettings.unselectedZoneKindTilingEnabled,
                      onChanged: mapVisualSettings.zoneKindMapStylingEnabled
                          ? mapVisualSettings.setUnselectedZoneKindTilingEnabled
                          : null,
                      title: const Text('Tile unselected discovered zones'),
                      subtitle: const Text(
                        'Keeps each discovered zone filled with its zone-kind tile even when it is not selected.',
                      ),
                      secondary: const Icon(Icons.grid_on_rounded),
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
                      value: mapVisualSettings.proximityBypassEnabled,
                      onChanged: mapVisualSettings.setProximityBypassEnabled,
                      title: const Text('Bypass proximity checks'),
                      subtitle: const Text(
                        'Reveal nearby-locked map content and allow interactions from anywhere.',
                      ),
                      secondary: const Icon(Icons.my_location_outlined),
                    ),
                  ),
                  const SizedBox(height: 16),
                  const Divider(),
                  const SizedBox(height: 12),
                  Text(
                    'Gameplay',
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
                    child: Padding(
                      padding: const EdgeInsets.fromLTRB(16, 14, 16, 12),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              const Icon(Icons.tune_rounded),
                              const SizedBox(width: 12),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      'Content difficulty',
                                      style: theme.textTheme.titleSmall
                                          ?.copyWith(
                                            fontWeight: FontWeight.w700,
                                          ),
                                    ),
                                    const SizedBox(height: 4),
                                    Text(
                                      _contentDifficultySubtitle(
                                        contentLevelOffset,
                                        userLevel,
                                      ),
                                      style: theme.textTheme.bodySmall
                                          ?.copyWith(
                                            color: scheme.onSurfaceVariant,
                                          ),
                                    ),
                                  ],
                                ),
                              ),
                              const SizedBox(width: 12),
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 10,
                                  vertical: 6,
                                ),
                                decoration: BoxDecoration(
                                  color: scheme.primary.withValues(alpha: 0.12),
                                  borderRadius: BorderRadius.circular(999),
                                ),
                                child: Text(
                                  _formatLevelOffset(contentLevelOffset),
                                  style: theme.textTheme.labelLarge?.copyWith(
                                    color: scheme.primary,
                                    fontWeight: FontWeight.w700,
                                  ),
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: 14),
                          Slider(
                            value: contentLevelOffset.toDouble(),
                            min: MapVisualSettingsProvider.minContentLevelOffset
                                .toDouble(),
                            max: MapVisualSettingsProvider.maxContentLevelOffset
                                .toDouble(),
                            divisions:
                                MapVisualSettingsProvider
                                    .maxContentLevelOffset -
                                MapVisualSettingsProvider.minContentLevelOffset,
                            label: _formatLevelOffset(contentLevelOffset),
                            onChanged: (value) {
                              mapVisualSettings.setContentLevelOffset(
                                value.round(),
                              );
                            },
                          ),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Text(
                                '${MapVisualSettingsProvider.minContentLevelOffset}',
                                style: theme.textTheme.bodySmall?.copyWith(
                                  color: scheme.onSurfaceVariant,
                                ),
                              ),
                              Text(
                                'Match level',
                                style: theme.textTheme.bodySmall?.copyWith(
                                  color: scheme.onSurfaceVariant,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                              Text(
                                '+${MapVisualSettingsProvider.maxContentLevelOffset}',
                                style: theme.textTheme.bodySmall?.copyWith(
                                  color: scheme.onSurfaceVariant,
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),
                  const Divider(),
                  const SizedBox(height: 12),
                  Text(
                    'Account',
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 10),
                  OutlinedButton.icon(
                    onPressed:
                        (_loading ||
                            _requesting ||
                            _sendingTestPush ||
                            _spawningNearbyContent ||
                            _triggeringTutorial)
                        ? null
                        : _handleLogout,
                    icon: const Icon(Icons.logout_rounded),
                    label: const Text('Log out'),
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
