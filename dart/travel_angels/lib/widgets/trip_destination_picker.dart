import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/models/travel_destination.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_trip_design.dart';

class _LocationDrag {
  const _LocationDrag(this.owner, this.identity, this.epoch);
  final Object owner;
  final String identity;
  final int epoch;
}

class _MoveLocationIntent extends Intent {
  const _MoveLocationIntent(this.offset);
  final int offset;
}

class TripDestinationPicker extends StatefulWidget {
  const TripDestinationPicker({
    super.key,
    required this.service,
    required this.value,
    required this.onChanged,
    this.enabled = true,
    this.errorText,
  });

  final TravelCalendarService service;
  final List<TravelDestination> value;
  final ValueChanged<List<TravelDestination>> onChanged;
  final bool enabled;
  final String? errorText;
  static const maxLocations = 12;

  @override
  State<TripDestinationPicker> createState() => _TripDestinationPickerState();
}

class _TripDestinationPickerState extends State<TripDestinationPicker> {
  final _input = TextEditingController();
  final _focus = FocusNode();
  final _dragOwner = Object();
  int _dragEpoch = 0;
  String? _focusedLocation;
  Timer? _debounce;
  int _request = 0;
  bool _loading = false;
  bool _initialized = false;
  bool _invalidated = false;
  String? _error;
  AuthProvider? _auth;
  String? _account;
  List<TravelDestination> _results = [];

  bool get _enabled => widget.enabled && !_invalidated;
  bool get _canAdd =>
      _enabled && widget.value.length < TripDestinationPicker.maxLocations;

  @override
  void initState() {
    super.initState();
    _focus.addListener(_focusChanged);
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final auth = context.read<AuthProvider?>();
    if (_auth != auth) {
      _auth?.removeListener(_accountChanged);
      _auth = auth;
      auth?.addListener(_accountChanged);
    }
    if (!_initialized) {
      _initialized = true;
      _account = auth?.user?.id;
    }
    _accountChanged();
  }

  @override
  void didUpdateWidget(TripDestinationPicker oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.enabled != widget.enabled ||
        oldWidget.service != widget.service) {
      ++_dragEpoch;
    }
    if (oldWidget.service != widget.service) {
      _debounce?.cancel();
      ++_request;
      _results = [];
      _error = null;
      _loading = false;
      if (_focus.hasFocus && _canAdd) _search(_input.text.trim());
    }
  }

  void _accountChanged() {
    if (_invalidated || _auth?.user?.id == _account) return;
    _invalidated = true;
    ++_dragEpoch;
    _debounce?.cancel();
    ++_request;
    _input.clear();
    _results = [];
    _loading = false;
    _error = null;
    if (mounted) setState(() {});
  }

  void _focusChanged() {
    if (!mounted) return;
    setState(() {});
    if (_focus.hasFocus && _canAdd) _search(_input.text.trim());
  }

  @override
  void dispose() {
    ++_request;
    _auth?.removeListener(_accountChanged);
    _debounce?.cancel();
    _focus.removeListener(_focusChanged);
    _focus.dispose();
    _input.dispose();
    super.dispose();
  }

  void _changed(String text) {
    _debounce?.cancel();
    ++_request;
    setState(() {
      _results = [];
      _error = null;
      _loading = _canAdd;
    });
    if (!_canAdd) return;
    _debounce = Timer(
      const Duration(milliseconds: 300),
      () => _search(text.trim()),
    );
  }

  Future<void> _search(String query) async {
    _debounce?.cancel();
    if (!_canAdd) return;
    final request = ++_request;
    final service = widget.service;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final response = await service.get('/destinations', query: {'q': query});
      if (!mounted ||
          !_enabled ||
          request != _request ||
          service != widget.service) {
        return;
      }
      final seen = <String>{};
      setState(() {
        _results =
            (response['destinations'] is List
                    ? response['destinations'] as List
                    : const [])
                .whereType<Map>()
                .map(
                  (value) => TravelDestination.fromJson(
                    Map<String, dynamic>.from(value),
                  ),
                )
                .where(
                  (value) => value.label.isNotEmpty && seen.add(value.identity),
                )
                .toList();
        _loading = false;
        _error = response['lookupUnavailable'] == true
            ? 'Could not look up more places. Try again.'
            : null;
      });
    } catch (error) {
      if (!mounted || !_enabled || request != _request) return;
      setState(() {
        _loading = false;
        _results = [];
        _error = tripError(error);
      });
    }
  }

  bool _selected(TravelDestination destination) =>
      widget.value.any((value) => value.identity == destination.identity);

  void _add(TravelDestination destination) {
    if (!_canAdd || _selected(destination)) return;
    _debounce?.cancel();
    ++_request;
    final updated = [...widget.value];
    final legacyIndex = destination.id == null
        ? -1
        : updated.indexWhere(
            (value) =>
                value.id == null &&
                value.name.trim().toLowerCase() ==
                    destination.name.trim().toLowerCase(),
          );
    if (legacyIndex >= 0) {
      updated[legacyIndex] = destination;
    } else {
      updated.add(destination);
    }
    widget.onChanged(updated);
    _input.clear();
    setState(() {
      _results = [];
      _loading = false;
      _error = null;
    });
    // Keep the add field available, but close the suggestions after a choice.
    _focus.unfocus();
  }

  void _move(String identity, int offset) {
    if (!_enabled) return;
    final index = widget.value.indexWhere((item) => item.identity == identity);
    if (index < 0 ||
        index + offset < 0 ||
        index + offset >= widget.value.length) {
      return;
    }
    final updated = [...widget.value];
    final destination = updated.removeAt(index);
    updated.insert(index + offset, destination);
    widget.onChanged(updated);
  }

  bool _acceptsDrag(_LocationDrag drag, String target) =>
      _enabled &&
      identical(drag.owner, _dragOwner) &&
      drag.epoch == _dragEpoch &&
      drag.identity != target &&
      widget.value.any((item) => item.identity == drag.identity) &&
      widget.value.any((item) => item.identity == target);

  void _drop(_LocationDrag drag, String target) {
    if (!_acceptsDrag(drag, target)) return;
    final from = widget.value.indexWhere(
      (item) => item.identity == drag.identity,
    );
    final to = widget.value.indexWhere((item) => item.identity == target);
    final updated = [...widget.value];
    updated.insert(to, updated.removeAt(from));
    widget.onChanged(updated);
  }

  void _remove(String identity) {
    if (!_enabled) return;
    widget.onChanged([
      for (final destination in widget.value)
        if (destination.identity != identity) destination,
    ]);
  }

  Widget _pill(
    TravelDestination destination,
    double width, {
    bool highlighted = false,
    bool feedback = false,
  }) => Container(
    constraints: BoxConstraints(maxWidth: width),
    decoration: BoxDecoration(
      color: tripPeach,
      border: Border.all(
        color: highlighted || _focusedLocation == destination.identity
            ? tripRust
            : tripOutline,
      ),
      borderRadius: BorderRadius.circular(4),
    ),
    padding: const EdgeInsets.only(left: 12),
    child: Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Flexible(
          child: Text(
            destination.label,
            style: const TextStyle(color: tripInk),
          ),
        ),
        IconButton(
          tooltip: 'Remove ${destination.label}',
          onPressed: _enabled && !feedback
              ? () => _remove(destination.identity)
              : null,
          icon: const Icon(Icons.close, size: 18),
        ),
      ],
    ),
  );

  Widget _chip(TravelDestination destination, int index, double width) {
    final canDrag = _enabled && widget.value.length > 1;
    final touchPlatform = switch (Theme.of(context).platform) {
      TargetPlatform.android || TargetPlatform.iOS => true,
      _ => false,
    };
    return FocusableActionDetector(
      key: ValueKey('trip-location-${destination.identity}'),
      enabled: _enabled,
      mouseCursor: canDrag ? SystemMouseCursors.grab : MouseCursor.defer,
      onShowFocusHighlight: (focused) {
        if (!mounted) return;
        setState(() {
          if (focused) {
            _focusedLocation = destination.identity;
          } else if (_focusedLocation == destination.identity) {
            _focusedLocation = null;
          }
        });
      },
      shortcuts: const {
        SingleActivator(LogicalKeyboardKey.arrowLeft, alt: true):
            _MoveLocationIntent(-1),
        SingleActivator(LogicalKeyboardKey.arrowUp, alt: true):
            _MoveLocationIntent(-1),
        SingleActivator(LogicalKeyboardKey.arrowRight, alt: true):
            _MoveLocationIntent(1),
        SingleActivator(LogicalKeyboardKey.arrowDown, alt: true):
            _MoveLocationIntent(1),
      },
      actions: {
        _MoveLocationIntent: CallbackAction<_MoveLocationIntent>(
          onInvoke: (intent) {
            _move(destination.identity, intent.offset);
            return null;
          },
        ),
      },
      child: Semantics(
        hint: canDrag
            ? touchPlatform
                  ? 'Press and hold to reorder'
                  : 'Drag to reorder'
            : null,
        customSemanticsActions: {
          if (_enabled && index > 0)
            const CustomSemanticsAction(label: 'Move earlier'): () =>
                _move(destination.identity, -1),
          if (_enabled && index < widget.value.length - 1)
            const CustomSemanticsAction(label: 'Move later'): () =>
                _move(destination.identity, 1),
        },
        child: DragTarget<_LocationDrag>(
          onWillAcceptWithDetails: (details) =>
              _acceptsDrag(details.data, destination.identity),
          onAcceptWithDetails: (details) =>
              _drop(details.data, destination.identity),
          builder: (context, candidates, rejected) {
            final child = _pill(
              destination,
              width,
              highlighted: candidates.isNotEmpty,
            );
            final feedback = Material(
              color: Colors.transparent,
              elevation: 6,
              borderRadius: BorderRadius.circular(4),
              child: _pill(
                destination,
                width,
                highlighted: true,
                feedback: true,
              ),
            );
            final data = _LocationDrag(
              _dragOwner,
              destination.identity,
              _dragEpoch,
            );
            if (touchPlatform) {
              return LongPressDraggable<_LocationDrag>(
                data: data,
                maxSimultaneousDrags: canDrag ? 1 : 0,
                feedback: feedback,
                childWhenDragging: Opacity(opacity: 0.35, child: child),
                child: child,
              );
            }
            return Draggable<_LocationDrag>(
              data: data,
              maxSimultaneousDrags: canDrag ? 1 : 0,
              feedback: feedback,
              childWhenDragging: Opacity(opacity: 0.35, child: child),
              child: child,
            );
          },
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final query = _input.text.trim();
    final visible = _results.where((value) => !_selected(value)).toList();
    final custom = TravelDestination(name: query);
    final canCustom =
        query.isNotEmpty &&
        !_selected(custom) &&
        ![
          ...widget.value,
          ...visible,
        ].any((value) => value.name.toLowerCase() == query.toLowerCase());
    return TextFieldTapRegion(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const Text('Locations', style: TextStyle(color: tripInk)),
          if (widget.value.isNotEmpty) ...[
            const SizedBox(height: 8),
            LayoutBuilder(
              builder: (context, constraints) => Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  for (var index = 0; index < widget.value.length; index++)
                    _chip(widget.value[index], index, constraints.maxWidth),
                ],
              ),
            ),
          ],
          const SizedBox(height: 8),
          TextField(
            key: const ValueKey('trip-location-search'),
            controller: _input,
            focusNode: _focus,
            enabled: _canAdd,
            textInputAction: TextInputAction.search,
            maxLength: 200,
            decoration: InputDecoration(
              labelText: 'Add location',
              hintText: 'City, country, or place',
              counterText: '',
              prefixIcon: const Icon(Icons.search),
              errorText: widget.errorText,
              helperText:
                  widget.value.length >= TripDestinationPicker.maxLocations
                  ? 'Remove a location to add another.'
                  : null,
              helperMaxLines: 2,
            ),
            onChanged: _changed,
            onTapOutside: (_) => _focus.unfocus(),
            onSubmitted: (_) {
              final exact = visible
                  .where(
                    (value) => value.name.toLowerCase() == query.toLowerCase(),
                  )
                  .firstOrNull;
              if (exact != null) {
                _add(exact);
              }
            },
          ),
          if (_focus.hasFocus && _canAdd)
            Container(
              key: const ValueKey('trip-location-results'),
              margin: const EdgeInsets.only(top: 8),
              constraints: const BoxConstraints(maxHeight: 224),
              decoration: BoxDecoration(
                color: tripPaper,
                border: Border.all(color: tripOutline),
                borderRadius: BorderRadius.circular(4),
              ),
              child: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    if (_loading)
                      const Padding(
                        padding: EdgeInsets.all(12),
                        child: Row(
                          children: [
                            SizedBox(
                              width: 16,
                              height: 16,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            ),
                            SizedBox(width: 12),
                            Flexible(child: Text('Finding locations…')),
                          ],
                        ),
                      )
                    else if (_error != null)
                      Padding(
                        padding: const EdgeInsets.all(12),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              _error!,
                              style: TextStyle(
                                color: Theme.of(context).colorScheme.error,
                              ),
                            ),
                            TextButton.icon(
                              onPressed: () => _search(query),
                              icon: const Icon(Icons.replay, size: 18),
                              label: const Text('Retry'),
                            ),
                          ],
                        ),
                      )
                    else if (visible.isEmpty)
                      Padding(
                        padding: const EdgeInsets.all(12),
                        child: Text(
                          query.isEmpty
                              ? 'Search for a city, country, or place.'
                              : 'No matching locations.',
                          style: const TextStyle(color: tripMuted),
                        ),
                      )
                    else
                      for (final destination in visible)
                        ListTile(
                          key: ValueKey(
                            'trip-location-result-${destination.identity}',
                          ),
                          dense: true,
                          contentPadding: const EdgeInsets.symmetric(
                            horizontal: 12,
                            vertical: 2,
                          ),
                          title: Text(destination.label),
                          subtitle: Text(destination.contextLabel),
                          trailing: const Icon(Icons.add, size: 18),
                          onTap: () => _add(destination),
                        ),
                    if (canCustom)
                      ListTile(
                        key: const ValueKey('trip-location-custom'),
                        dense: true,
                        contentPadding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 2,
                        ),
                        leading: const Icon(
                          Icons.add_location_alt_outlined,
                          size: 20,
                        ),
                        title: Text('Use “$query” as a custom location'),
                        subtitle: const Text('Only for this trip'),
                        onTap: () => _add(custom),
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
