import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../models/user_zone_reputation.dart';
import '../providers/location_provider.dart';
import '../providers/zone_provider.dart';
import '../services/poi_service.dart';

class ZoneWidget extends StatefulWidget {
  final VoidCallback? onWidgetOpen;
  final VoidCallback? onWidgetClose;
  final double top;

  const ZoneWidget({
    super.key,
    this.onWidgetOpen,
    this.onWidgetClose,
    this.top = 80,
  });

  @override
  State<ZoneWidget> createState() => _ZoneWidgetState();
}

class _ZoneWidgetState extends State<ZoneWidget> {
  bool _isOpen = false;
  bool _showContent = false;
  UserZoneReputation? _reputation;
  bool _loadingReputation = false;
  String? _lastZoneId;
  String? _lastLocationKey;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      _updateSelectedZoneFromLocation();
    });
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final location = context.watch<LocationProvider>().location;
    final locationKey = location == null
        ? null
        : '${location.latitude.toStringAsFixed(5)},${location.longitude.toStringAsFixed(5)}';
    if (locationKey != _lastLocationKey) {
      _lastLocationKey = locationKey;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        _updateSelectedZoneFromLocation();
      });
    }

    final selectedZone = context.watch<ZoneProvider>().selectedZone;
    if (selectedZone?.id != _lastZoneId) {
      _lastZoneId = selectedZone?.id;
      if (selectedZone != null) {
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (!mounted) return;
          _loadReputation(selectedZone.id);
        });
      }
    }
  }

  void _updateSelectedZoneFromLocation() {
    final location = context.read<LocationProvider>().location;
    if (location == null) return;

    final zoneProvider = context.read<ZoneProvider>();
    final zone = zoneProvider.findZoneAtCoordinate(location.latitude, location.longitude);
    if (zone != null) {
      zoneProvider.setSelectedZone(zone);
    }
  }

  Future<void> _loadReputation(String zoneId) async {
    if (_loadingReputation) return;
    setState(() => _loadingReputation = true);
    try {
      final svc = context.read<PoiService>();
      final rep = await svc.getUserZoneReputation(zoneId);
      if (mounted) {
        setState(() {
          _reputation = rep;
          _loadingReputation = false;
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() => _loadingReputation = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<ZoneProvider>(
      builder: (context, zoneProvider, _) {
        final selectedZone = zoneProvider.selectedZone;
        final theme = Theme.of(context);
        final surfaceColor = theme.colorScheme.surface.withValues(alpha: 0.95);
        final borderColor = theme.colorScheme.outlineVariant;
        final textStyle = theme.textTheme.bodyMedium?.copyWith(
          color: theme.colorScheme.onSurface,
          fontWeight: FontWeight.w600,
        );
        final subTextStyle = theme.textTheme.bodySmall?.copyWith(
          color: theme.colorScheme.onSurface,
        );

        return Positioned(
          top: widget.top,
          left: 0,
          right: 0,
          child: Center(
            child: GestureDetector(
              onTap: () {
                setState(() {
                  _isOpen = !_isOpen;
                  if (_isOpen) {
                    widget.onWidgetOpen?.call();
                    Future.delayed(const Duration(milliseconds: 300), () {
                      if (mounted) {
                        setState(() => _showContent = true);
                      }
                    });
                  } else {
                    _showContent = false;
                    widget.onWidgetClose?.call();
                  }
                });
              },
              child: AnimatedContainer(
                duration: const Duration(milliseconds: 300),
                width: _isOpen ? 256 : 144,
                constraints: BoxConstraints(
                  minHeight: _isOpen ? 0 : 40,
                ),
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: surfaceColor,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: borderColor, width: 1.5),
                  boxShadow: const [
                    BoxShadow(
                      color: Color(0x332D2416),
                      blurRadius: 10,
                      offset: Offset(0, 4),
                    ),
                  ],
                ),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Expanded(
                          child: Text(
                            selectedZone?.name ?? 'Hinterlands',
                            style: textStyle,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        Icon(
                          _isOpen ? Icons.keyboard_arrow_up : Icons.keyboard_arrow_down,
                          size: 16,
                          color: theme.colorScheme.onSurface,
                        ),
                      ],
                    ),
                    AnimatedSize(
                      duration: const Duration(milliseconds: 300),
                      child: _showContent && _isOpen
                          ? Column(
                              mainAxisSize: MainAxisSize.min,
                              crossAxisAlignment: CrossAxisAlignment.stretch,
                              children: [
                                if (selectedZone?.description != null) ...[
                                  const SizedBox(height: 8),
                                  Text(
                                    selectedZone!.description!,
                                    style: subTextStyle,
                                  ),
                                ],
                                if (_reputation != null) ...[
                                  const SizedBox(height: 8),
                                  Row(
                                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                    children: [
                                      Text(
                                        'Reputation: ${_reputation!.name.name}',
                                        style: textStyle?.copyWith(fontSize: 14),
                                      ),
                                      Text(
                                        '${_reputation!.reputationOnLevel} / ${_reputation!.reputationToNextLevel}',
                                        style: subTextStyle,
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 4),
                                  ClipRRect(
                                    borderRadius: BorderRadius.circular(4),
                                    child: LinearProgressIndicator(
                                      value: _reputation!.reputationToNextLevel > 0
                                          ? _reputation!.reputationOnLevel / _reputation!.reputationToNextLevel
                                          : 0.0,
                                      backgroundColor: theme.colorScheme.surfaceVariant,
                                      valueColor: AlwaysStoppedAnimation<Color>(
                                        theme.colorScheme.primary,
                                      ),
                                    ),
                                  ),
                                ],
                              ],
                            )
                          : const SizedBox.shrink(),
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      },
    );
  }
}
