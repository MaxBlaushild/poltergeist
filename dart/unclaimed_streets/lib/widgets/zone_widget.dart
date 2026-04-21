import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../constants/zone_kind_visuals.dart';
import '../models/zone.dart';
import '../models/user_zone_reputation.dart';
import '../providers/location_provider.dart';
import '../providers/zone_provider.dart';
import '../services/poi_service.dart';

class ZoneWidget extends StatefulWidget {
  final VoidCallback? onWidgetOpen;
  final VoidCallback? onWidgetClose;
  final ZoneWidgetController? controller;
  final bool expandUpwards;
  final double expandedHeight;

  const ZoneWidget({
    super.key,
    this.onWidgetOpen,
    this.onWidgetClose,
    this.controller,
    this.expandUpwards = false,
    this.expandedHeight = 260,
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
    widget.controller?._attach(_setOpen, () => _isOpen);
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      _updateSelectedZoneFromLocation();
    });
  }

  @override
  void didUpdateWidget(covariant ZoneWidget oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.controller != widget.controller) {
      oldWidget.controller?._detach();
      widget.controller?._attach(_setOpen, () => _isOpen);
    }
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
    final zone = zoneProvider.findZoneAtCoordinate(
      location.latitude,
      location.longitude,
    );
    zoneProvider.setSelectedZone(zone?.discovered == true ? zone : null);
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

  void _setOpen(bool value) {
    if (_isOpen == value) return;
    setState(() {
      _isOpen = value;
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
  }

  String _capitalize(String value) {
    if (value.isEmpty) return value;
    return value[0].toUpperCase() + value.substring(1);
  }

  Zone? _zoneAtCurrentLocation(ZoneProvider zoneProvider) {
    final location = context.watch<LocationProvider>().location;
    if (location == null) return null;
    return zoneProvider.findZoneAtCoordinate(
      location.latitude,
      location.longitude,
    );
  }

  List<_GenrePreviewEntry> _genreScoresPreview() {
    final selectedZone = context.read<ZoneProvider>().selectedZone;
    final scores = selectedZone?.genreScores ?? const [];
    if (scores.isEmpty) {
      return const <_GenrePreviewEntry>[];
    }
    final sorted =
        scores
            .map((entry) => _GenrePreviewEntry(entry.genre.name, entry.score))
            .toList(growable: false)
          ..sort((left, right) {
            final scoreCompare = right.score.compareTo(left.score);
            if (scoreCompare != 0) {
              return scoreCompare;
            }
            return left.name.toLowerCase().compareTo(right.name.toLowerCase());
          });
    final hasNonZero = sorted.any((entry) => entry.score > 0);
    return hasNonZero
        ? sorted
              .where((entry) => entry.score > 0)
              .take(4)
              .toList(growable: false)
        : sorted.take(4).toList(growable: false);
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<ZoneProvider>(
      builder: (context, zoneProvider, _) {
        final selectedZone = zoneProvider.selectedZone;
        final locationZone = _zoneAtCurrentLocation(zoneProvider);
        final undiscoveredZone =
            selectedZone == null &&
                locationZone != null &&
                !locationZone.discovered
            ? locationZone
            : null;
        final displayedZone = selectedZone ?? undiscoveredZone;
        final showingUndiscovered = undiscoveredZone != null;
        final theme = Theme.of(context);
        final zoneVisuals = displayedZone != null
            ? zoneKindVisualProfileForSlug(displayedZone.kind)
            : null;
        final surfaceColor = zoneVisuals != null
            ? zoneVisuals
                  .surfaceColor(undiscovered: showingUndiscovered)
                  .withValues(alpha: showingUndiscovered ? 0.97 : 0.95)
            : (showingUndiscovered
                  ? const Color(0xFF1C2430).withValues(alpha: 0.96)
                  : theme.colorScheme.surface.withValues(alpha: 0.95));
        final borderColor = zoneVisuals != null
            ? zoneVisuals.borderColor(undiscovered: showingUndiscovered)
            : (showingUndiscovered
                  ? const Color(0xFFD3BF88)
                  : theme.colorScheme.outlineVariant);
        final accentColor = zoneVisuals != null
            ? zoneVisuals.accentColor(undiscovered: showingUndiscovered)
            : borderColor;
        final chipBackgroundColor = zoneVisuals != null
            ? zoneVisuals.chipBackgroundColor(undiscovered: showingUndiscovered)
            : theme.colorScheme.surfaceContainerHighest;
        final chipTextColor = zoneVisuals != null
            ? zoneVisuals.chipTextColor(undiscovered: showingUndiscovered)
            : theme.colorScheme.onSurface;
        final primaryTextColor =
            zoneVisuals?.panelText ??
            (showingUndiscovered
                ? const Color(0xFFF6E7B8)
                : theme.colorScheme.onSurface);
        final secondaryTextColor =
            zoneVisuals?.panelSubtext ??
            (showingUndiscovered
                ? const Color(0xFFD7DDE8)
                : theme.colorScheme.onSurface);
        final textStyle = theme.textTheme.bodyMedium?.copyWith(
          color: primaryTextColor,
          fontWeight: FontWeight.w600,
        );
        final subTextStyle = theme.textTheme.bodySmall?.copyWith(
          color: secondaryTextColor,
        );
        final zoneKindLabel =
            zoneVisuals?.label ?? humanizeZoneKindSlug(displayedZone?.kind);
        final genreScoresPreview = _genreScoresPreview();
        final expandUpwards = widget.expandUpwards;
        final expandedHeight = widget.expandedHeight;
        const collapsedHeight = 40.0;
        final arrowIcon = _isOpen
            ? (expandUpwards
                  ? Icons.keyboard_arrow_down
                  : Icons.keyboard_arrow_up)
            : (expandUpwards
                  ? Icons.keyboard_arrow_up
                  : Icons.keyboard_arrow_down);
        final headerLeading = showingUndiscovered
            ? Icon(Icons.explore_off, size: 16, color: primaryTextColor)
            : zoneVisuals != null
            ? Container(
                width: 22,
                height: 22,
                decoration: BoxDecoration(
                  color: chipBackgroundColor,
                  borderRadius: BorderRadius.circular(999),
                  border: Border.all(
                    color: accentColor.withValues(alpha: 0.28),
                    width: 1,
                  ),
                ),
                child: Icon(zoneVisuals.icon, size: 13, color: chipTextColor),
              )
            : null;
        final header = Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            if (headerLeading != null) ...[
              headerLeading,
              const SizedBox(width: 8),
            ],
            Expanded(
              child: Text(
                showingUndiscovered
                    ? 'Uncharted Territory'
                    : displayedZone?.name ?? 'Hinterlands',
                style: textStyle,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Icon(arrowIcon, size: 16, color: primaryTextColor),
          ],
        );
        final content = _showContent && _isOpen
            ? SingleChildScrollView(
                padding: EdgeInsets.only(
                  top: expandUpwards ? 0 : 8,
                  bottom: expandUpwards ? 8 : 0,
                ),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    if (showingUndiscovered) ...[
                      Container(
                        padding: const EdgeInsets.all(10),
                        decoration: BoxDecoration(
                          color: chipBackgroundColor.withValues(alpha: 0.78),
                          borderRadius: BorderRadius.circular(10),
                          border: Border.all(
                            color: accentColor.withValues(alpha: 0.44),
                            width: 1,
                          ),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Unknown lands ahead',
                              style: textStyle?.copyWith(fontSize: 14),
                            ),
                            const SizedBox(height: 6),
                            Text(
                              'Step inside this shrouded zone to uncover its true name and earn a small discovery reward.',
                              style: subTextStyle,
                            ),
                          ],
                        ),
                      ),
                    ] else ...[
                      if (zoneVisuals != null) ...[
                        Wrap(
                          spacing: 6,
                          runSpacing: 6,
                          children: [
                            _ZoneMetaChip(
                              icon: zoneVisuals.icon,
                              label: zoneKindLabel,
                              backgroundColor: chipBackgroundColor,
                              textColor: chipTextColor,
                              borderColor: accentColor.withValues(alpha: 0.32),
                            ),
                            _ZoneMetaChip(
                              label: zoneVisuals.focusLabel,
                              backgroundColor: chipBackgroundColor.withValues(
                                alpha: 0.9,
                              ),
                              textColor: chipTextColor,
                              borderColor: accentColor.withValues(alpha: 0.18),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Text(
                          zoneVisuals.atmosphereLabel,
                          style: subTextStyle?.copyWith(
                            color: secondaryTextColor.withValues(alpha: 0.9),
                            fontStyle: FontStyle.italic,
                          ),
                        ),
                        const SizedBox(height: 10),
                      ],
                      if (_reputation != null) ...[
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              'Reputation: ${_capitalize(_reputation!.name.name)}',
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
                                ? _reputation!.reputationOnLevel /
                                      _reputation!.reputationToNextLevel
                                : 0.0,
                            backgroundColor: chipBackgroundColor.withValues(
                              alpha: 0.8,
                            ),
                            valueColor: AlwaysStoppedAnimation<Color>(
                              accentColor,
                            ),
                          ),
                        ),
                      ],
                    ],
                    if (!showingUndiscovered &&
                        _reputation != null &&
                        displayedZone?.description != null)
                      const SizedBox(height: 8),
                    if (!showingUndiscovered &&
                        displayedZone?.description != null) ...[
                      Text(displayedZone!.description!, style: subTextStyle),
                    ],
                    if (!showingUndiscovered &&
                        displayedZone?.genreScores.isNotEmpty == true) ...[
                      const SizedBox(height: 10),
                      Text('Genres', style: textStyle?.copyWith(fontSize: 13)),
                      const SizedBox(height: 6),
                      Wrap(
                        spacing: 6,
                        runSpacing: 6,
                        children: [
                          ...genreScoresPreview.map(
                            (entry) => Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 4,
                              ),
                              decoration: BoxDecoration(
                                color: chipBackgroundColor,
                                borderRadius: BorderRadius.circular(999),
                              ),
                              child: Text(
                                '${entry.name} ${entry.score}',
                                style: subTextStyle?.copyWith(
                                  color: chipTextColor,
                                ),
                              ),
                            ),
                          ),
                          if ((displayedZone?.genreScores.length ?? 0) >
                              genreScoresPreview.length)
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 4,
                              ),
                              decoration: BoxDecoration(
                                color: chipBackgroundColor,
                                borderRadius: BorderRadius.circular(999),
                              ),
                              child: Text(
                                '+${(displayedZone?.genreScores.length ?? 0) - genreScoresPreview.length} more',
                                style: subTextStyle?.copyWith(
                                  color: chipTextColor,
                                ),
                              ),
                            ),
                        ],
                      ),
                    ],
                  ],
                ),
              )
            : const SizedBox.shrink();

        return AnimatedContainer(
          duration: const Duration(milliseconds: 300),
          width: _isOpen ? 256 : 144,
          height: _isOpen ? null : collapsedHeight,
          constraints: _isOpen
              ? BoxConstraints(
                  minHeight: collapsedHeight,
                  maxHeight: expandedHeight,
                )
              : const BoxConstraints.tightFor(height: collapsedHeight),
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: surfaceColor,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: borderColor, width: 1.6),
            boxShadow: [
              BoxShadow(
                color:
                    zoneVisuals?.shadowColor(
                      undiscovered: showingUndiscovered,
                    ) ??
                    (showingUndiscovered
                        ? const Color(0x44101723)
                        : const Color(0x332D2416)),
                blurRadius: showingUndiscovered ? 18 : 12,
                offset: const Offset(0, 4),
              ),
            ],
          ),
          child: Material(
            color: Colors.transparent,
            child: InkWell(
              borderRadius: BorderRadius.circular(12),
              onTap: () {
                _setOpen(!_isOpen);
              },
              child: _isOpen
                  ? Column(
                      mainAxisSize: MainAxisSize.min,
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        if (expandUpwards) ...[
                          Flexible(fit: FlexFit.loose, child: content),
                          header,
                        ] else ...[
                          header,
                          Flexible(fit: FlexFit.loose, child: content),
                        ],
                      ],
                    )
                  : Center(child: header),
            ),
          ),
        );
      },
    );
  }
}

class ZoneWidgetController {
  void Function(bool isOpen)? _setOpen;
  bool Function()? _isOpen;

  void _attach(void Function(bool) setOpen, bool Function() isOpen) {
    _setOpen = setOpen;
    _isOpen = isOpen;
  }

  void _detach() {
    _setOpen = null;
    _isOpen = null;
  }

  bool get isOpen => _isOpen?.call() ?? false;

  void open() => _setOpen?.call(true);
  void close() => _setOpen?.call(false);
  void toggle() => _setOpen?.call(!isOpen);
}

class _ZoneMetaChip extends StatelessWidget {
  final IconData? icon;
  final String label;
  final Color backgroundColor;
  final Color textColor;
  final Color borderColor;

  const _ZoneMetaChip({
    required this.label,
    required this.backgroundColor,
    required this.textColor,
    required this.borderColor,
    this.icon,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 5),
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: borderColor, width: 1),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (icon != null) ...[
            Icon(icon, size: 12, color: textColor),
            const SizedBox(width: 4),
          ],
          Text(
            label,
            style: Theme.of(context).textTheme.labelSmall?.copyWith(
              color: textColor,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }
}

class _GenrePreviewEntry {
  const _GenrePreviewEntry(this.name, this.score);

  final String name;
  final int score;
}
