import 'dart:math' as math;
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:url_launcher/url_launcher.dart';

import '../constants/api_constants.dart';
import '../models/character.dart';
import '../models/inventory_item.dart';
import '../models/point_of_interest.dart';
import '../models/quest.dart';
import '../models/quest_node.dart';
import '../providers/auth_provider.dart';
import '../providers/discoveries_provider.dart';
import '../providers/location_provider.dart';
import '../providers/quest_log_provider.dart';
import '../services/inventory_service.dart';
import '../services/media_service.dart';
import '../services/poi_service.dart';
import '../utils/camera_capture.dart';
import '../widgets/paper_texture.dart';

const _placeholderImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp';

enum QuestSubmissionOverlayPhase { hidden, loading, success, failure }

typedef QuestSubmissionOverlayCallback = void Function(
  QuestSubmissionOverlayPhase phase, {
  String? message,
});

/// Unlock radius in meters. Must match backend (POST /sonar/pointOfInterest/unlock).
const _unlockRadiusMeters = 200.0;

/// Bottom-sheet content for a tapped point of interest.
/// When [hasDiscovered] is false: tags, how to unlock, distance, and Unlock button (disabled when too far).
/// When true: full panel with name, image, description, tags, etc.
class PointOfInterestPanel extends StatefulWidget {
  const PointOfInterestPanel({
    super.key,
    required this.pointOfInterest,
    required this.hasDiscovered,
    this.quest,
    this.questNode,
    required this.onClose,
    this.onUnlocked,
    this.onCharacterTap,
    this.onQuestSubmissionState,
  });

  final PointOfInterest pointOfInterest;
  final bool hasDiscovered;
  final Quest? quest;
  final QuestNode? questNode;
  final VoidCallback onClose;
  /// Called after successful unlock (e.g. refresh discoveries and POI markers). Optional.
  final Future<void> Function()? onUnlocked;
  final void Function(Character character)? onCharacterTap;
  final QuestSubmissionOverlayCallback? onQuestSubmissionState;

  @override
  State<PointOfInterestPanel> createState() => _PointOfInterestPanelState();
}

class _PointOfInterestPanelState extends State<PointOfInterestPanel> {
  bool _loading = false;
  bool _justUnlocked = false;
  String? _error;
  bool _isDescriptionExpanded = false;
  bool _loadingTelescope = false;
  bool _usingTelescope = false;
  bool _telescopeChecked = false;
  String? _telescopeError;
  InventoryItem? _telescopeItem;
  OwnedInventoryItem? _ownedTelescope;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      _maybeLoadTelescope();
    });
  }

  @override
  void didUpdateWidget(covariant PointOfInterestPanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.hasDiscovered != widget.hasDiscovered ||
        oldWidget.pointOfInterest.id != widget.pointOfInterest.id) {
      _telescopeChecked = false;
      _ownedTelescope = null;
      _telescopeItem = null;
      _telescopeError = null;
      _maybeLoadTelescope();
    }
  }

  Future<void> _maybeLoadTelescope() async {
    if (_telescopeChecked || widget.hasDiscovered) return;
    _telescopeChecked = true;
    await _loadTelescope();
  }

  Future<void> _loadTelescope() async {
    if (_loadingTelescope) return;
    setState(() {
      _loadingTelescope = true;
      _telescopeError = null;
    });
    try {
      final svc = context.read<InventoryService>();
      final itemsFuture = svc.getInventoryItems();
      final ownedFuture = svc.getOwnedInventoryItems();
      final items = await itemsFuture;
      final owned = await ownedFuture;
      if (!mounted) return;
      final item = items.firstWhere(
        (i) => i.name.trim().toLowerCase() == 'golden telescope',
        orElse: () => const InventoryItem(
          id: 0,
          name: '',
          imageUrl: '',
          flavorText: '',
          effectText: '',
        ),
      );
      InventoryItem? telescopeItem =
          item.id == 0 ? null : item;
      OwnedInventoryItem? telescopeOwned;
      if (telescopeItem != null) {
        telescopeOwned = owned.firstWhere(
          (o) =>
              o.inventoryItemId == telescopeItem.id &&
              o.quantity > 0,
          orElse: () => const OwnedInventoryItem(
            id: '',
            inventoryItemId: 0,
            quantity: 0,
          ),
        );
        if (telescopeOwned.id.isEmpty || telescopeOwned.quantity <= 0) {
          telescopeOwned = null;
        }
      }
      setState(() {
        _telescopeItem = telescopeItem;
        _ownedTelescope = telescopeOwned;
        _loadingTelescope = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loadingTelescope = false;
        _telescopeError = _errorMessage(e);
      });
    }
  }

  static String _formatTagName(String name) {
    final parts = name.split('_');
    return parts
        .asMap()
        .entries
        .map((e) {
          final s = e.value;
          if (s.isEmpty) return '';
          return e.key == 0
              ? (s[0].toUpperCase() + s.substring(1).toLowerCase())
              : s.toLowerCase();
        })
        .join(' ');
  }

  double _haversineMeters(double lat1, double lon1, double lat2, double lon2) {
    const R = 6371e3;
    final phi1 = lat1 * math.pi / 180;
    final phi2 = lat2 * math.pi / 180;
    final dPhi = (lat2 - lat1) * math.pi / 180;
    final dLambda = (lon2 - lon1) * math.pi / 180;
    final a = math.sin(dPhi / 2) * math.sin(dPhi / 2) +
        math.cos(phi1) * math.cos(phi2) *
            math.sin(dLambda / 2) * math.sin(dLambda / 2);
    final c = 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a));
    return R * c;
  }

  String _errorMessage(Object e) {
    if (e is DioException && e.response?.data is Map) {
      final d = e.response!.data as Map<String, dynamic>;
      final msg = d['error'] ?? d['message'];
      if (msg != null && msg.toString().isNotEmpty) return msg.toString();
    }
    return e.toString();
  }

  bool _isDiscoveryDuplicateError(Object e) {
    final msg = _errorMessage(e).toLowerCase();
    final mentionsDiscovery = msg.contains('discover') || msg.contains('point_of_interest');
    final mentionsDuplicate = msg.contains('duplicate') ||
        msg.contains('already') ||
        msg.contains('unique') ||
        msg.contains('constraint');
    return mentionsDiscovery && mentionsDuplicate;
  }

  bool _isAlreadyDiscovered() {
    if (widget.hasDiscovered || _justUnlocked) return true;
    try {
      return context
          .read<DiscoveriesProvider>()
          .hasDiscovered(widget.pointOfInterest.id);
    } catch (_) {
      return false;
    }
  }

  Future<void> _handleUnlock() async {
    if (_loading) return;
    if (_isAlreadyDiscovered()) {
      await widget.onUnlocked?.call();
      if (!mounted) return;
      setState(() {
        _justUnlocked = true;
        _error = null;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Already discovered.')),
      );
      return;
    }
    final loc = context.read<LocationProvider>().location;
    if (loc == null) {
      setState(() => _error = 'Location not available. Enable location access.');
      return;
    }
    final userId = context.read<AuthProvider>().user?.id;
    if (userId == null || userId.isEmpty) {
      setState(() => _error = 'Please log in to unlock.');
      return;
    }
    final poi = widget.pointOfInterest;
    final plat = double.tryParse(poi.lat) ?? 0.0;
    final plng = double.tryParse(poi.lng) ?? 0.0;
    final dist = _haversineMeters(loc.latitude, loc.longitude, plat, plng);
    if (dist > _unlockRadiusMeters) {
      setState(() => _error = 'Too far away (${dist.round()} m). Get within ${_unlockRadiusMeters.round()} m to unlock.');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await context.read<PoiService>().unlockPointOfInterest(
            poi.id,
            loc.latitude,
            loc.longitude,
            userId: userId,
          );
      if (!mounted) return;
      await widget.onUnlocked?.call();
      if (!mounted) return;
      setState(() {
        _justUnlocked = true;
        _loading = false;
        _error = null;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Discovered!')),
      );
    } catch (e) {
      if (_isDiscoveryDuplicateError(e)) {
        if (!mounted) return;
        await widget.onUnlocked?.call();
        if (!mounted) return;
        setState(() {
          _justUnlocked = true;
          _loading = false;
          _error = null;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Already discovered.')),
        );
        return;
      }
      if (mounted) {
        setState(() {
          _loading = false;
          _error = _errorMessage(e);
        });
      }
    }
  }

  Future<void> _handleTelescopeUnlock() async {
    if (_usingTelescope) return;
    if (_isAlreadyDiscovered()) {
      await widget.onUnlocked?.call();
      if (!mounted) return;
      setState(() {
        _justUnlocked = true;
        _usingTelescope = false;
        _telescopeError = null;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Already discovered.')),
      );
      return;
    }
    final owned = _ownedTelescope;
    if (owned == null || owned.quantity <= 0) {
      setState(() => _telescopeError = 'No Golden Telescope available.');
      return;
    }
    final userId = context.read<AuthProvider>().user?.id;
    if (userId == null || userId.isEmpty) {
      setState(() => _telescopeError = 'Please log in to use the Golden Telescope.');
      return;
    }
    final poi = widget.pointOfInterest;
    final plat = double.tryParse(poi.lat) ?? 0.0;
    final plng = double.tryParse(poi.lng) ?? 0.0;
    setState(() {
      _usingTelescope = true;
      _telescopeError = null;
    });
    try {
      await context.read<PoiService>().unlockPointOfInterest(
            poi.id,
            plat,
            plng,
            userId: userId,
          );
      String? consumeWarning;
      try {
        await context.read<InventoryService>().useItem(owned.id);
      } catch (e) {
        consumeWarning =
            'Discovered, but we could not consume the Golden Telescope. Please check your inventory.';
        debugPrint('Golden Telescope consumption failed: $e');
      }
      if (!mounted) return;
      await widget.onUnlocked?.call();
      if (!mounted) return;
      setState(() {
        _justUnlocked = true;
        _usingTelescope = false;
      });
      await _loadTelescope();
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            consumeWarning ?? 'Discovered with the Golden Telescope!',
          ),
        ),
      );
    } catch (e) {
      if (_isDiscoveryDuplicateError(e)) {
        if (!mounted) return;
        await widget.onUnlocked?.call();
        if (!mounted) return;
        setState(() {
          _justUnlocked = true;
          _usingTelescope = false;
          _telescopeError = null;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Already discovered.')),
        );
        return;
      }
      if (!mounted) return;
      setState(() {
        _usingTelescope = false;
        _telescopeError = _errorMessage(e);
      });
    }
  }

  Future<void> _showQuestSubmissionModal() async {
    final quest = widget.quest;
    final node = widget.questNode;
    if (quest == null || node == null) return;

    final textController = TextEditingController();
    CapturedImage? capturedImage;
    bool uploadingImage = false;
    String? selectedChallengeId = node.challenges.isNotEmpty
        ? node.challenges.first.id
        : null;

    await showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (context) {
        return Padding(
          padding: EdgeInsets.only(
            left: 16,
            right: 16,
            bottom: MediaQuery.viewInsetsOf(context).bottom + 24,
            top: 16,
          ),
          child: StatefulBuilder(
            builder: (context, setModalState) {
              final canUseCamera = kIsWeb ||
                  defaultTargetPlatform == TargetPlatform.iOS ||
                  defaultTargetPlatform == TargetPlatform.android;
              return Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    quest.name,
                    style: Theme.of(context).textTheme.titleLarge?.copyWith(
                          fontWeight: FontWeight.bold,
                        ),
                  ),
                  const SizedBox(height: 12),
                  if (node.challenges.length > 1)
                    DropdownButtonFormField<String>(
                      value: selectedChallengeId,
                      items: node.challenges
                          .map(
                            (c) => DropdownMenuItem(
                              value: c.id,
                              child: Text(c.question),
                            ),
                          )
                          .toList(),
                      onChanged: (value) {
                        setModalState(() => selectedChallengeId = value);
                      },
                      decoration: const InputDecoration(
                        labelText: 'Challenge',
                        border: OutlineInputBorder(),
                      ),
                    )
                  else if (node.challenges.isNotEmpty)
                    Text(
                      node.challenges.first.question,
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: textController,
                    decoration: const InputDecoration(
                      labelText: 'Answer',
                      border: OutlineInputBorder(),
                    ),
                    maxLines: 3,
                  ),
                  const SizedBox(height: 12),
                  if (canUseCamera)
                    Row(
                      children: [
                        Expanded(
                          child: OutlinedButton.icon(
                            onPressed: uploadingImage
                                ? null
                                : () async {
                                    final result = await captureImageFromCamera();
                                    if (!mounted) return;
                                    if (result == null || result.bytes.isEmpty) {
                                      ScaffoldMessenger.of(context).showSnackBar(
                                        const SnackBar(
                                          content: Text('No photo captured.'),
                                        ),
                                      );
                                      return;
                                    }
                                    setModalState(() => capturedImage = result);
                                  },
                            icon: const Icon(Icons.photo_camera),
                            label: const Text('Take photo'),
                          ),
                        ),
                        if (capturedImage != null) ...[
                          const SizedBox(width: 12),
                          TextButton(
                            onPressed: () => setModalState(() => capturedImage = null),
                            child: const Text('Clear'),
                          ),
                        ],
                      ],
                    ),
                  if (capturedImage != null) ...[
                    const SizedBox(height: 12),
                    ClipRRect(
                      borderRadius: BorderRadius.circular(8),
                      child: Image.memory(
                        capturedImage!.bytes,
                        height: 160,
                        fit: BoxFit.cover,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Captured photo will be uploaded on submit.',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ],
                  const SizedBox(height: 16),
                  FilledButton(
                    onPressed: uploadingImage
                        ? null
                        : () async {
                            final mediaService = context.read<MediaService>();
                            final questLogProvider = context.read<QuestLogProvider>();
                            final userId =
                                context.read<AuthProvider>().user?.id ?? 'anonymous';
                            final startedAt = DateTime.now();
                            setModalState(() => uploadingImage = true);
                            Navigator.of(context).pop();
                            widget.onClose();
                            widget.onQuestSubmissionState?.call(
                              QuestSubmissionOverlayPhase.loading,
                            );
                            String? imageSubmissionUrl;
                            if (capturedImage != null) {
                              final ext = _extensionFromMime(
                                    capturedImage!.mimeType,
                                    capturedImage!.name,
                                  ) ??
                                  'jpg';
                              final key =
                                  'quest-submissions/$userId/${DateTime.now().millisecondsSinceEpoch}.$ext';
                              final url = await mediaService.getPresignedUploadUrl(
                                ApiConstants.crewPointsOfInterestBucket,
                                key,
                              );
                              if (url == null) {
                                final elapsed = DateTime.now().difference(startedAt);
                                if (elapsed < const Duration(milliseconds: 700)) {
                                  await Future<void>.delayed(
                                    const Duration(milliseconds: 700),
                                  );
                                }
                                widget.onQuestSubmissionState?.call(
                                  QuestSubmissionOverlayPhase.failure,
                                  message: 'Failed to prepare image upload.',
                                );
                                return;
                              }
                              final ok = await mediaService.uploadToPresigned(
                                url,
                                Uint8List.fromList(capturedImage!.bytes),
                                capturedImage!.mimeType ?? 'image/jpeg',
                              );
                              if (!ok) {
                                final elapsed = DateTime.now().difference(startedAt);
                                if (elapsed < const Duration(milliseconds: 700)) {
                                  await Future<void>.delayed(
                                    const Duration(milliseconds: 700),
                                  );
                                }
                                widget.onQuestSubmissionState?.call(
                                  QuestSubmissionOverlayPhase.failure,
                                  message: 'Failed to upload photo.',
                                );
                                return;
                              }
                              imageSubmissionUrl = url.split('?').first;
                            }
                            final resp = await questLogProvider.submitQuestNodeChallenge(
                              node.id,
                              questNodeChallengeId: selectedChallengeId,
                              textSubmission: textController.text.trim(),
                              imageSubmissionUrl: imageSubmissionUrl,
                            );
                            final elapsed = DateTime.now().difference(startedAt);
                            if (elapsed < const Duration(milliseconds: 700)) {
                              await Future<void>.delayed(
                                const Duration(milliseconds: 700),
                              );
                            }
                            final success = resp['successful'] == true;
                            final reason = resp['reason']?.toString() ?? '';
                            widget.onQuestSubmissionState?.call(
                              success
                                  ? QuestSubmissionOverlayPhase.success
                                  : QuestSubmissionOverlayPhase.failure,
                              message: success
                                  ? (reason.isNotEmpty
                                      ? reason
                                      : 'Challenge completed!')
                                  : (reason.isNotEmpty ? reason : 'Submission failed'),
                            );
                          },
                    child: const Text('Submit'),
                  ),
                ],
              );
            },
          ),
        );
      },
    );
  }

  String? _extensionFromMime(String? mimeType, String? filename) {
    final name = filename ?? '';
    final dot = name.lastIndexOf('.');
    if (dot != -1 && dot < name.length - 1) {
      return name.substring(dot + 1).toLowerCase();
    }
    switch (mimeType) {
      case 'image/png':
        return 'png';
      case 'image/gif':
        return 'gif';
      case 'image/webp':
        return 'webp';
      case 'image/jpeg':
      case 'image/jpg':
        return 'jpg';
      default:
        return null;
    }
  }

  @override
  Widget build(BuildContext context) {
    final poi = widget.pointOfInterest;
    final showFull = widget.hasDiscovered || _justUnlocked;
    if (!showFull) {
      return _buildUndiscovered(context, poi);
    }
    return _buildFullPanel(context, poi);
  }

  Widget _buildUndiscovered(BuildContext context, PointOfInterest poi) {
    final loc = context.watch<LocationProvider>().location;
    final plat = double.tryParse(poi.lat) ?? 0.0;
    final plng = double.tryParse(poi.lng) ?? 0.0;
    final distance = loc != null
        ? _haversineMeters(loc.latitude, loc.longitude, plat, plng)
        : null;
    final withinRange = distance != null && distance <= _unlockRadiusMeters;
    final tags = poi.tags;

    return DraggableScrollableSheet(
      initialChildSize: 0.9,
      minChildSize: 0.4,
      maxChildSize: 0.95,
      builder: (_, scrollController) => PaperSheet(
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Row(
                    children: [
                      Icon(
                        Icons.lock_outline,
                        size: 28,
                        color: Theme.of(context).colorScheme.primary,
                      ),
                      const SizedBox(width: 10),
                      Text(
                        'Undiscovered',
                        style: Theme.of(context).textTheme.titleLarge?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                      ),
                    ],
                  ),
                  IconButton(
                    onPressed: widget.onClose,
                    icon: const Icon(Icons.close),
                  ),
                ],
              ),
            ),
            Expanded(
              child: ListView(
                controller: scrollController,
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
                children: [
                  // Tags
                  if (tags.isNotEmpty) ...[
                    Text(
                      'Tags',
                      style: Theme.of(context).textTheme.titleSmall?.copyWith(
                            fontWeight: FontWeight.w600,
                            color: Theme.of(context)
                                .colorScheme
                                .onSurface
                                .withValues(alpha: 0.8),
                          ),
                    ),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 8,
                      runSpacing: 6,
                      children: tags
                          .map(
                            (t) => Chip(
                              label: Text(
                                _formatTagName(t.name),
                                style: const TextStyle(fontSize: 12),
                              ),
                              materialTapTargetSize:
                                  MaterialTapTargetSize.shrinkWrap,
                            ),
                          )
                          .toList(),
                    ),
                    const SizedBox(height: 20),
                  ],
                  // How to unlock
                  Text(
                    'Visit this location to unlock this point of interest. You must be within ${_unlockRadiusMeters.round()} meters to discover it.',
                    style: Theme.of(context).textTheme.bodyLarge,
                  ),
                  const SizedBox(height: 16),
                  // Distance
                  if (distance != null)
                    Text(
                      withinRange
                          ? 'Within range! Tap Unlock to discover.'
                          : 'You are ${distance.round()} m away.',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: withinRange
                                ? Theme.of(context).colorScheme.primary
                                : Theme.of(context)
                                    .colorScheme
                                    .onSurface
                                    .withValues(alpha: 0.7),
                            fontWeight: withinRange ? FontWeight.w600 : null,
                          ),
                    )
                  else
                    Text(
                      'Enable location to see distance.',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: Theme.of(context)
                                .colorScheme
                                .onSurface
                                .withValues(alpha: 0.6),
                          ),
                    ),
                  if (_error != null) ...[
                    const SizedBox(height: 12),
                    Text(
                      _error!,
                      style: TextStyle(
                        color: Theme.of(context).colorScheme.error,
                        fontSize: 14,
                      ),
                    ),
                  ],
                  if (_loadingTelescope) ...[
                    const SizedBox(height: 16),
                    Row(
                      children: [
                        const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        ),
                        const SizedBox(width: 8),
                        Text(
                          'Checking Golden Telescope…',
                          style: Theme.of(context).textTheme.bodyMedium,
                        ),
                      ],
                    ),
                  ],
                  if (!_loadingTelescope && _ownedTelescope != null) ...[
                    const SizedBox(height: 20),
                    Container(
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: Theme.of(context)
                            .colorScheme
                            .surfaceContainerHighest,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(
                          color: Theme.of(context).dividerColor,
                        ),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          Row(
                            children: [
                              Icon(
                                Icons.remove_red_eye_outlined,
                                color: Theme.of(context).colorScheme.primary,
                              ),
                              const SizedBox(width: 8),
                              Text(
                                'Golden Telescope',
                                style: Theme.of(context)
                                    .textTheme
                                    .titleMedium
                                    ?.copyWith(fontWeight: FontWeight.w600),
                              ),
                              const Spacer(),
                              Text(
                                'x${_ownedTelescope!.quantity}',
                                style: Theme.of(context).textTheme.labelLarge,
                              ),
                            ],
                          ),
                          const SizedBox(height: 8),
                          Text(
                            'Reveal this hidden point of interest from anywhere. Consumes one Golden Telescope.',
                            style: Theme.of(context).textTheme.bodyMedium,
                          ),
                          if (_telescopeError != null) ...[
                            const SizedBox(height: 8),
                            Text(
                              _telescopeError!,
                              style: TextStyle(
                                color: Theme.of(context).colorScheme.error,
                              ),
                            ),
                          ],
                          const SizedBox(height: 12),
                          FilledButton(
                            onPressed: _usingTelescope ? null : _handleTelescopeUnlock,
                            child: Text(
                              _usingTelescope
                                  ? 'Revealing…'
                                  : 'Use Golden Telescope',
                            ),
                          ),
                        ],
                      ),
                    ),
                  ] else if (_telescopeError != null) ...[
                    const SizedBox(height: 16),
                    Text(
                      _telescopeError!,
                      style: TextStyle(
                        color: Theme.of(context).colorScheme.error,
                      ),
                    ),
                  ],
                  const SizedBox(height: 24),
                  FilledButton(
                    onPressed: (_loading || !withinRange) ? null : _handleUnlock,
                    child: Text(
                      _loading
                          ? 'Unlocking…'
                          : !withinRange
                              ? 'Too far to unlock'
                              : 'Unlock',
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFullPanel(BuildContext context, PointOfInterest poi) {
    final imageUrl = (poi.imageURL != null && poi.imageURL!.isNotEmpty)
        ? poi.imageURL!
        : _placeholderImageUrl;
    final tags = poi.tags;
    final characters = poi.characters;

    return DraggableScrollableSheet(
      initialChildSize: 0.9,
      minChildSize: 0.4,
      maxChildSize: 0.95,
      builder: (_, scrollController) => PaperSheet(
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          poi.name,
                          style: Theme.of(context).textTheme.titleLarge?.copyWith(
                                fontWeight: FontWeight.bold,
                              ),
                        ),
                        if (poi.originalName != null &&
                            poi.originalName!.isNotEmpty) ...[
                          const SizedBox(height: 4),
                          Text(
                            poi.originalName!,
                            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                  color: Theme.of(context)
                                      .colorScheme
                                      .onSurface
                                      .withValues(alpha: 0.7),
                                ),
                          ),
                        ],
                        if (poi.googleMapsPlaceId != null &&
                            poi.googleMapsPlaceId!.isNotEmpty) ...[
                          const SizedBox(height: 4),
                          GestureDetector(
                            onTap: () async {
                              final uri = Uri.parse(
                                'https://www.google.com/maps/place/?q=place_id:${poi.googleMapsPlaceId}',
                              );
                              try {
                                await launchUrl(uri, mode: LaunchMode.externalApplication);
                              } catch (_) {}
                            },
                            child: Text(
                              'View on Google Maps',
                              style: TextStyle(
                                color: Theme.of(context).colorScheme.primary,
                                decoration: TextDecoration.underline,
                                fontSize: 14,
                              ),
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),
                  IconButton(
                    onPressed: widget.onClose,
                    icon: const Icon(Icons.close),
                  ),
                ],
              ),
            ),
            Expanded(
              child: ListView(
                controller: scrollController,
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
                children: [
                  ClipRRect(
                    borderRadius: BorderRadius.circular(12),
                    child: Image.network(
                      imageUrl,
                      height: 200,
                      width: double.infinity,
                      fit: BoxFit.cover,
                      errorBuilder: (_, __, ___) => Container(
                        height: 200,
                        color: Colors.grey.shade300,
                        child: const Icon(Icons.image_not_supported, size: 48),
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),
                  if (widget.quest != null && widget.questNode != null) ...[
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.amber.shade50,
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(color: Colors.amber.shade200),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Quest Objective',
                            style: Theme.of(context).textTheme.titleSmall?.copyWith(
                                  fontWeight: FontWeight.bold,
                                ),
                          ),
                          const SizedBox(height: 6),
                          ...widget.questNode!.challenges.map(
                                (c) => Padding(
                                  padding: const EdgeInsets.only(bottom: 4),
                                  child: Text(
                                    '• ${c.question}',
                                    style: Theme.of(context).textTheme.bodySmall,
                                  ),
                                ),
                              ),
                          const SizedBox(height: 8),
                          FilledButton(
                            onPressed: _showQuestSubmissionModal,
                            child: Text('Quest: ${widget.quest!.name}'),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 16),
                  ],
                  if (tags.isNotEmpty) ...[
                    Text(
                      'Tags',
                      style: Theme.of(context).textTheme.titleSmall?.copyWith(
                            fontWeight: FontWeight.w600,
                            color: Theme.of(context)
                                .colorScheme
                                .onSurface
                                .withValues(alpha: 0.8),
                          ),
                    ),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 8,
                      runSpacing: 6,
                      children: tags
                          .map(
                            (t) => Chip(
                              label: Text(
                                _formatTagName(t.name),
                                style: const TextStyle(fontSize: 12),
                              ),
                              materialTapTargetSize:
                                  MaterialTapTargetSize.shrinkWrap,
                            ),
                          )
                          .toList(),
                    ),
                    const SizedBox(height: 16),
                  ],
                  if (poi.description != null && poi.description!.isNotEmpty) ...[
                    GestureDetector(
                      onTap: () => setState(() => _isDescriptionExpanded = !_isDescriptionExpanded),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            'Description',
                            style: Theme.of(context).textTheme.titleSmall?.copyWith(
                                  fontWeight: FontWeight.w600,
                                  color: Theme.of(context)
                                      .colorScheme
                                      .onSurface
                                      .withValues(alpha: 0.8),
                                ),
                          ),
                          Icon(
                            _isDescriptionExpanded
                                ? Icons.keyboard_arrow_up
                                : Icons.keyboard_arrow_down,
                            color: Theme.of(context)
                                .colorScheme
                                .onSurface
                                .withValues(alpha: 0.6),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 8),
                    AnimatedCrossFade(
                      firstChild: Text(
                        poi.description!,
                        maxLines: 3,
                        overflow: TextOverflow.ellipsis,
                        style: Theme.of(context).textTheme.bodyLarge,
                      ),
                      secondChild: Text(
                        poi.description!,
                        style: Theme.of(context).textTheme.bodyLarge,
                      ),
                      crossFadeState: _isDescriptionExpanded
                          ? CrossFadeState.showSecond
                          : CrossFadeState.showFirst,
                      duration: const Duration(milliseconds: 200),
                    ),
                    const SizedBox(height: 12),
                  ],
                  if (characters.isNotEmpty) ...[
                    const SizedBox(height: 8),
                    Text(
                      'Characters',
                      style: Theme.of(context).textTheme.titleSmall?.copyWith(
                            fontWeight: FontWeight.w600,
                            color: Theme.of(context)
                                .colorScheme
                                .onSurface
                                .withValues(alpha: 0.8),
                          ),
                    ),
                    const SizedBox(height: 8),
                    SizedBox(
                      height: 120,
                      child: ListView.separated(
                        scrollDirection: Axis.horizontal,
                        itemCount: characters.length,
                        separatorBuilder: (_, __) => const SizedBox(width: 12),
                        itemBuilder: (_, i) {
                          final ch = characters[i];
                          final imageUrl = ch.dialogueImageUrl ?? ch.mapIconUrl;
                          return InkWell(
                            onTap: widget.onCharacterTap == null
                                ? null
                                : () => widget.onCharacterTap!(ch),
                            borderRadius: BorderRadius.circular(12),
                            child: Container(
                              width: 120,
                              padding: const EdgeInsets.all(8),
                              decoration: BoxDecoration(
                                color: Theme.of(context).colorScheme.surfaceVariant,
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Column(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  SizedBox(
                                    width: 56,
                                    height: 56,
                                    child: Stack(
                                      clipBehavior: Clip.none,
                                      children: [
                                        CircleAvatar(
                                          radius: 28,
                                          backgroundColor: Colors.grey.shade300,
                                          backgroundImage:
                                              imageUrl != null ? NetworkImage(imageUrl) : null,
                                          child: imageUrl == null
                                              ? const Icon(Icons.person)
                                              : null,
                                        ),
                                        if (ch.hasAvailableQuest)
                                          Positioned(
                                            right: -2,
                                            top: -2,
                                            child: Container(
                                              width: 20,
                                              height: 20,
                                              decoration: BoxDecoration(
                                                color: const Color(0xFFF5C542),
                                                shape: BoxShape.circle,
                                                border: Border.all(
                                                  color: Colors.white,
                                                  width: 2,
                                                ),
                                                boxShadow: const [
                                                  BoxShadow(
                                                    color: Colors.black26,
                                                    blurRadius: 4,
                                                    offset: Offset(0, 2),
                                                  ),
                                                ],
                                              ),
                                              child: const Center(
                                                child: Text(
                                                  '!',
                                                  style: TextStyle(
                                                    fontSize: 12,
                                                    fontWeight: FontWeight.w800,
                                                    color: Color(0xFF3A2400),
                                                  ),
                                                ),
                                              ),
                                            ),
                                          ),
                                      ],
                                    ),
                                  ),
                                  const SizedBox(height: 8),
                                  Text(
                                    ch.name,
                                    maxLines: 2,
                                    overflow: TextOverflow.ellipsis,
                                    textAlign: TextAlign.center,
                                    style: Theme.of(context).textTheme.bodySmall,
                                  ),
                                ],
                              ),
                            ),
                          );
                        },
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
