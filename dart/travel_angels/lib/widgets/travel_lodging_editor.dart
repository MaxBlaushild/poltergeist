import 'dart:async';
import 'dart:ui' as ui;

import 'package:file_picker/file_picker.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_lodging_card.dart';
import 'package:travel_angels/widgets/travel_trip_design.dart';

class PendingLodgingPhoto {
  PendingLodgingPhoto({required this.bytes, required this.filename});
  final Uint8List bytes;
  final String filename;
  String altText = '';
}

/// Provider content lives only in this editor. Only the selected place ID is
/// returned to the trip draft; uploaded photos are the host's own content.
class TravelLodgingSourceEditor extends StatefulWidget {
  const TravelLodgingSourceEditor({
    super.key,
    required this.stopId,
    required this.service,
    required this.property,
    required this.photos,
    required this.pendingPhotos,
    required this.onPropertyChanged,
    required this.onPhotosChanged,
    required this.onPendingChanged,
    required this.onPickingChanged,
    this.busy = false,
  });
  final String stopId;
  final TravelCalendarService service;
  final Map<String, dynamic> property;
  final List<Map<String, dynamic>> photos;
  final List<PendingLodgingPhoto> pendingPhotos;
  final ValueChanged<Map<String, dynamic>?> onPropertyChanged;
  final ValueChanged<List<Map<String, dynamic>>> onPhotosChanged;
  final ValueChanged<List<PendingLodgingPhoto>> onPendingChanged;
  final ValueChanged<bool> onPickingChanged;
  final bool busy;

  @override
  State<TravelLodgingSourceEditor> createState() =>
      _TravelLodgingSourceEditorState();
}

class _TravelLodgingSourceEditorState extends State<TravelLodgingSourceEditor> {
  final _search = TextEditingController();
  Timer? _timer;
  int _generation = 0;
  bool _loading = false;
  bool _picking = false;
  bool _searched = false;
  bool _searchOpen = false;
  String? _error;
  String? _photoError;
  List<Map<String, dynamic>> _results = [];
  Map<String, dynamic>? _preview;
  Map<String, dynamic>? _selected;

  String get _base =>
      '/calendar/stops/${Uri.encodeComponent(widget.stopId)}/lodging-properties';
  bool get _disabled => widget.busy || _picking;
  String get _placeId => widget.property['placeId'] as String? ?? '';

  @override
  void initState() {
    super.initState();
    if (_placeId.isNotEmpty) _loadSelected();
  }

  @override
  void didUpdateWidget(covariant TravelLodgingSourceEditor oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.property['placeId'] != widget.property['placeId'] ||
        oldWidget.stopId != widget.stopId ||
        oldWidget.service != widget.service) {
      _generation++;
      _timer?.cancel();
      _results = [];
      _preview = null;
      if (_selected?['placeId'] != _placeId ||
          oldWidget.service != widget.service ||
          oldWidget.stopId != widget.stopId) {
        _selected = null;
        if (_placeId.isNotEmpty) _loadSelected();
      }
    }
  }

  @override
  void dispose() {
    _generation++;
    _timer?.cancel();
    _search.dispose();
    super.dispose();
  }

  Future<void> _loadSelected() async {
    final generation = ++_generation;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final response = await widget.service.get(
        '$_base/${Uri.encodeComponent(_placeId)}',
      );
      if (!mounted || generation != _generation) return;
      setState(() => _selected = tripMap(response['property']));
    } catch (_) {
      if (mounted && generation == _generation) {
        setState(
          () => _error =
              'Property details are unavailable. Your linked property is still saved.',
        );
      }
    } finally {
      if (mounted && generation == _generation) {
        setState(() => _loading = false);
      }
    }
  }

  void _searchChanged(String value) {
    _timer?.cancel();
    final generation = ++_generation;
    final query = value.trim();
    setState(() {
      _preview = null;
      _results = [];
      _error = null;
      _searched = false;
      _loading = query.length >= 2;
    });
    if (query.length < 2) return;
    _timer = Timer(const Duration(milliseconds: 350), () async {
      try {
        final response = await widget.service.get(_base, query: {'q': query});
        if (!mounted || generation != _generation) return;
        setState(() {
          _results = tripMaps(response['properties']);
          _searched = true;
        });
      } catch (_) {
        if (mounted && generation == _generation) {
          setState(
            () => _error =
                'Could not search properties. Try again or add your stay manually.',
          );
        }
      } finally {
        if (mounted && generation == _generation) {
          setState(() => _loading = false);
        }
      }
    });
  }

  Future<void> _review(Map<String, dynamic> candidate) async {
    if (_disabled) return;
    _timer?.cancel();
    final generation = ++_generation;
    setState(() {
      _loading = true;
      _error = null;
      _preview = candidate;
    });
    try {
      final response = await widget.service.get(
        '$_base/${Uri.encodeComponent(candidate['placeId'] as String)}',
      );
      if (!mounted || generation != _generation) return;
      setState(() => _preview = tripMap(response['property']));
    } catch (_) {
      if (mounted && generation == _generation) {
        setState(() {
          _preview = null;
          _error = 'Could not load this property. Try another match.';
        });
      }
    } finally {
      if (mounted && generation == _generation) {
        setState(() => _loading = false);
      }
    }
  }

  void _attach() {
    if (_disabled || _loading || _preview == null) return;
    final selected = _preview!;
    setState(() {
      _selected = selected;
      _preview = null;
      _results = [];
      _searchOpen = false;
      _search.clear();
      _error = null;
    });
    widget.onPropertyChanged({
      'provider': 'google',
      'placeId': selected['placeId'],
    });
    FocusScope.of(context).unfocus();
  }

  void _unlink() {
    if (_disabled) return;
    _generation++;
    _timer?.cancel();
    setState(() {
      _selected = null;
      _preview = null;
      _results = [];
      _search.clear();
      _searchOpen = false;
      _searched = false;
      _loading = false;
      _error = null;
    });
    widget.onPropertyChanged(null);
  }

  List<Map<String, dynamic>> get _searchAttributions {
    final sources = <String, Map<String, dynamic>>{};
    for (final property in _results.take(5)) {
      for (final source in tripMaps(property['attributions'])) {
        sources['${source['displayName']}|${source['uri']}'] = source;
      }
    }
    return sources.values.toList();
  }

  Future<void> _pickPhoto() async {
    if (_disabled || widget.photos.length + widget.pendingPhotos.length >= 6) {
      return;
    }
    setState(() {
      _picking = true;
      _photoError = null;
    });
    widget.onPickingChanged(true);
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['jpg', 'jpeg', 'png'],
        allowMultiple: false,
        withData: true,
      );
      if (!mounted || result == null) return;
      final file = result.files.single;
      final bytes = file.bytes;
      if (bytes == null || bytes.isEmpty || bytes.length > 8 * 1024 * 1024) {
        throw const FormatException('Choose a JPEG or PNG image up to 8 MB.');
      }
      final jpeg =
          bytes.length >= 3 &&
          bytes[0] == 255 &&
          bytes[1] == 216 &&
          bytes[2] == 255;
      final png =
          bytes.length >= 8 &&
          listEquals(bytes.sublist(0, 8), [137, 80, 78, 71, 13, 10, 26, 10]);
      if (!jpeg && !png) {
        throw const FormatException('Choose a JPEG or PNG image.');
      }
      final buffer = await ui.ImmutableBuffer.fromUint8List(bytes);
      try {
        final descriptor = await ui.ImageDescriptor.encoded(buffer);
        try {
          if (descriptor.width * descriptor.height > 20000000) {
            throw const FormatException(
              'Choose an image smaller than 20 megapixels.',
            );
          }
          final codec = await descriptor.instantiateCodec(targetWidth: 264);
          try {
            final frame = await codec.getNextFrame();
            frame.image.dispose();
          } finally {
            codec.dispose();
          }
        } finally {
          descriptor.dispose();
        }
      } finally {
        buffer.dispose();
      }
      if (!mounted) return;
      widget.onPendingChanged([
        ...widget.pendingPhotos,
        PendingLodgingPhoto(bytes: bytes, filename: file.name),
      ]);
    } catch (error) {
      if (mounted) {
        setState(
          () => _photoError = error is FormatException
              ? error.message
              : 'Could not open that image. Try another photo.',
        );
      }
    } finally {
      if (mounted) {
        setState(() => _picking = false);
        widget.onPickingChanged(false);
      }
    }
  }

  Widget _propertyPreview(
    Map<String, dynamic> property, {
    required bool candidate,
  }) {
    final photos = tripMaps(property['photos']);
    return Container(
      margin: const EdgeInsets.only(top: 12),
      decoration: BoxDecoration(
        color: tripCream,
        border: Border.all(color: tripOutline),
        borderRadius: BorderRadius.circular(4),
      ),
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (photos.isNotEmpty)
            AspectRatio(
              aspectRatio: 16 / 9,
              child: LodgingPhoto(photo: photos.first, service: widget.service),
            ),
          Padding(
            padding: const EdgeInsets.all(14),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  property['name'] as String? ?? 'Linked property',
                  style: const TextStyle(
                    fontFamily: tripDisplayFont,
                    fontSize: 24,
                    color: tripInk,
                  ),
                ),
                if (property['address'] is String)
                  Padding(
                    padding: const EdgeInsets.only(top: 6),
                    child: Text(
                      property['address'] as String,
                      style: const TextStyle(color: tripMuted),
                    ),
                  ),
                const SizedBox(height: 10),
                LodgingGoogleAttribution(
                  attributions: [
                    ...tripMaps(property['attributions']),
                    if (photos.isNotEmpty)
                      ...tripMaps(photos.first['authorAttributions']),
                  ],
                  googleMapsUri: photos.isEmpty
                      ? null
                      : photos.first['googleMapsUri'] as String?,
                ),
                const SizedBox(height: 8),
                if (candidate)
                  FilledButton.tonalIcon(
                    onPressed: _disabled || _loading ? null : _attach,
                    icon: const Icon(Icons.check, size: 18),
                    label: const Text('Use this property'),
                  )
                else
                  Wrap(
                    spacing: 8,
                    children: [
                      TextButton(
                        onPressed: _disabled
                            ? null
                            : () => setState(() => _searchOpen = true),
                        child: const Text('Change property'),
                      ),
                      TextButton(
                        onPressed: _disabled ? null : _unlink,
                        child: const Text('Unlink'),
                      ),
                    ],
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _pendingPhoto(PendingLodgingPhoto photo) {
    final thumbnail = ClipRRect(
      borderRadius: BorderRadius.circular(3),
      child: Image.memory(
        photo.bytes,
        width: 88,
        height: 72,
        fit: BoxFit.cover,
        cacheWidth: 264,
        errorBuilder: (context, error, stackTrace) => const SizedBox(
          width: 88,
          height: 72,
          child: Icon(Icons.broken_image_outlined, color: tripMuted),
        ),
      ),
    );
    final description = TextFormField(
      key: ObjectKey(photo),
      initialValue: photo.altText,
      enabled: !_disabled,
      maxLength: 500,
      decoration: const InputDecoration(
        labelText: 'Photo description',
        hintText: 'The garden terrace',
        counterText: '',
      ),
      validator: (text) =>
          text == null || text.trim().isEmpty ? 'Describe this photo.' : null,
      onChanged: (value) => photo.altText = value,
    );
    final remove = IconButton(
      tooltip: 'Remove selected photo',
      onPressed: _disabled
          ? null
          : () => widget.onPendingChanged(
              widget.pendingPhotos.where((p) => p != photo).toList(),
            ),
      icon: const Icon(Icons.close, size: 18),
    );
    // AlertDialog reserves its own inset and content padding on small screens.
    final narrow = MediaQuery.sizeOf(context).width < 520;
    return Padding(
      padding: const EdgeInsets.only(bottom: 14),
      child: narrow
          ? Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Row(
                  children: [
                    thumbnail,
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        photo.filename,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(color: tripMuted, fontSize: 12),
                      ),
                    ),
                    remove,
                  ],
                ),
                const SizedBox(height: 12),
                description,
              ],
            )
          : Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                thumbnail,
                const SizedBox(width: 12),
                Expanded(child: description),
                remove,
              ],
            ),
    );
  }

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    children: [
      if (_selected != null && !_searchOpen)
        _propertyPreview(_selected!, candidate: false)
      else ...[
        TextField(
          key: const ValueKey('lodging-property-search'),
          controller: _search,
          enabled: !_disabled,
          onChanged: _searchChanged,
          textInputAction: TextInputAction.search,
          onSubmitted: _searchChanged,
          decoration: InputDecoration(
            labelText: 'Find a property',
            hintText: 'Hotel, guesthouse, or place name',
            prefixIcon: const Icon(Icons.search, size: 20),
            suffixIcon: _loading
                ? const Padding(
                    padding: EdgeInsets.all(14),
                    child: SizedBox.square(
                      dimension: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                  )
                : null,
          ),
        ),
        if (_placeId.isNotEmpty)
          Wrap(
            children: [
              TextButton(
                onPressed: _disabled
                    ? null
                    : () {
                        _generation++;
                        _timer?.cancel();
                        setState(() {
                          _searchOpen = false;
                          _results = [];
                          _preview = null;
                          _loading = false;
                          _error = null;
                        });
                        if (_selected == null) _loadSelected();
                      },
                child: const Text('Keep linked property'),
              ),
              TextButton(
                onPressed: _disabled ? null : _unlink,
                child: const Text('Unlink'),
              ),
            ],
          ),
        if (_preview != null)
          _propertyPreview(_preview!, candidate: true)
        else if (_results.isNotEmpty)
          Container(
            margin: const EdgeInsets.only(top: 8),
            decoration: BoxDecoration(
              border: Border.all(color: tripOutline),
              borderRadius: BorderRadius.circular(4),
            ),
            child: Column(
              children: [
                for (final property in _results.take(5))
                  ListTile(
                    dense: true,
                    title: Text(property['name'] as String? ?? 'Property'),
                    subtitle: Text(property['address'] as String? ?? ''),
                    trailing: const Icon(Icons.chevron_right, size: 18),
                    onTap: _disabled ? null : () => _review(property),
                  ),
                Padding(
                  padding: const EdgeInsets.all(10),
                  child: LodgingGoogleAttribution(
                    attributions: _searchAttributions,
                  ),
                ),
              ],
            ),
          ),
        if (_searched && _results.isEmpty && !_loading)
          const Padding(
            padding: EdgeInsets.only(top: 8),
            child: Text(
              'No matching properties. Try a city in your search, or add the details below.',
              style: TextStyle(color: tripMuted),
            ),
          ),
      ],
      if (_error != null)
        Padding(
          padding: const EdgeInsets.only(top: 8),
          child: Text(
            _error!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        ),
      const SizedBox(height: 12),
      Wrap(
        crossAxisAlignment: WrapCrossAlignment.center,
        spacing: 8,
        children: [
          TextButton.icon(
            key: const ValueKey('lodging-add-photo'),
            onPressed:
                _disabled ||
                    widget.photos.length + widget.pendingPhotos.length >= 6
                ? null
                : _pickPhoto,
            icon: const Icon(Icons.add_photo_alternate_outlined, size: 18),
            label: Text(_picking ? 'Opening images…' : 'Add your own photos'),
          ),
          if (widget.photos.isNotEmpty || widget.pendingPhotos.isNotEmpty)
            Text(
              '${widget.photos.length + widget.pendingPhotos.length}/6',
              style: const TextStyle(color: tripMuted, fontSize: 12),
            ),
        ],
      ),
      for (final photo in widget.photos)
        Padding(
          padding: const EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              SizedBox(
                width: 88,
                height: 64,
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(3),
                  child: LodgingPhoto(
                    photo: photo,
                    service: widget.service,
                    accessQuery: const {'draft': true},
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  photo['altText'] as String? ?? 'Lodging photo',
                  maxLines: 3,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              IconButton(
                tooltip: 'Remove photo',
                onPressed: _disabled
                    ? null
                    : () => widget.onPhotosChanged(
                        widget.photos
                            .where((p) => p['mediaId'] != photo['mediaId'])
                            .toList(),
                      ),
                icon: const Icon(Icons.close, size: 18),
              ),
            ],
          ),
        ),
      for (final photo in widget.pendingPhotos) _pendingPhoto(photo),
      if (_photoError != null)
        Text(
          _photoError!,
          style: TextStyle(color: Theme.of(context).colorScheme.error),
        ),
      if (_placeId.isNotEmpty ||
          widget.photos.isNotEmpty ||
          widget.pendingPhotos.isNotEmpty)
        const Padding(
          padding: EdgeInsets.only(top: 4, bottom: 16),
          child: Text(
            'The property, map, and photos share this stay’s audience.',
            style: TextStyle(color: tripMuted, fontSize: 12),
          ),
        ),
      const SizedBox(height: 8),
    ],
  );
}
