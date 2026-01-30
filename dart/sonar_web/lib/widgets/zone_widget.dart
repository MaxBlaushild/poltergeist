import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../models/user_zone_reputation.dart';
import '../providers/location_provider.dart';
import '../providers/zone_provider.dart';
import '../services/poi_service.dart';

class ZoneWidget extends StatefulWidget {
  final VoidCallback? onWidgetOpen;
  final VoidCallback? onWidgetClose;

  const ZoneWidget({
    super.key,
    this.onWidgetOpen,
    this.onWidgetClose,
  });

  @override
  State<ZoneWidget> createState() => _ZoneWidgetState();
}

class _ZoneWidgetState extends State<ZoneWidget> {
  bool _isOpen = false;
  bool _showContent = false;
  UserZoneReputation? _reputation;
  bool _loadingReputation = false;

  @override
  void initState() {
    super.initState();
    _updateSelectedZoneFromLocation();
  }

  void _updateSelectedZoneFromLocation() {
    final location = context.read<LocationProvider>().location;
    if (location == null) return;

    final zoneProvider = context.read<ZoneProvider>();
    final zone = zoneProvider.findZoneAtCoordinate(location.latitude, location.longitude);
    if (zone != null) {
      zoneProvider.setSelectedZone(zone);
      _loadReputation(zone.id);
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

        // Update zone when location changes
        WidgetsBinding.instance.addPostFrameCallback((_) {
          _updateSelectedZoneFromLocation();
        });

        // Load reputation when zone changes
        if (selectedZone != null && _reputation?.zoneId != selectedZone.id) {
          _loadReputation(selectedZone.id);
        }

        return Positioned(
          top: 80,
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
                  color: Colors.white.withOpacity(0.8),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.black, width: 2),
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
                            style: const TextStyle(fontSize: 14),
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        Icon(
                          _isOpen ? Icons.keyboard_arrow_up : Icons.keyboard_arrow_down,
                          size: 16,
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
                                    style: const TextStyle(fontSize: 14),
                                  ),
                                ],
                                if (_reputation != null) ...[
                                  const SizedBox(height: 8),
                                  Row(
                                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                    children: [
                                      Text(
                                        'Reputation: ${_reputation!.name.name}',
                                        style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
                                      ),
                                      Text(
                                        '${_reputation!.reputationOnLevel} / ${_reputation!.reputationToNextLevel}',
                                        style: const TextStyle(fontSize: 12),
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
