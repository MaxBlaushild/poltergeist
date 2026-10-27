import 'dart:convert';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/travel_trip_design.dart';

String _text(dynamic value) => value is String ? value.trim() : '';

/// A protected image, decoded only for the lifetime of this widget.
class LodgingPhoto extends StatefulWidget {
  const LodgingPhoto({
    super.key,
    required this.photo,
    required this.service,
    this.accessQuery = const {},
    this.fit = BoxFit.cover,
    this.onTap,
  });

  final Map<String, dynamic> photo;
  final TravelCalendarService service;
  final Map<String, dynamic> accessQuery;
  final BoxFit fit;
  final VoidCallback? onTap;

  @override
  State<LodgingPhoto> createState() => _LodgingPhotoState();
}

class _LodgingPhotoState extends State<LodgingPhoto> {
  ui.Image? _image;
  bool _loading = false;
  int _generation = 0;
  late String _fingerprint;
  AuthProvider? _auth;
  String? _identity;

  String get _currentFingerprint =>
      jsonEncode([widget.photo, widget.accessQuery]);

  @override
  void initState() {
    super.initState();
    _fingerprint = _currentFingerprint;
    _load();
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final auth = context.read<AuthProvider?>();
    if (auth != _auth) {
      _auth?.removeListener(_accountChanged);
      _auth = auth;
      _identity = auth?.user?.id;
      _auth?.addListener(_accountChanged);
    }
  }

  void _accountChanged() {
    if (_identity == _auth?.user?.id) return;
    _identity = _auth?.user?.id;
    _generation++;
    _image?.dispose();
    setState(() {
      _image = null;
      _loading = false;
    });
  }

  @override
  void didUpdateWidget(covariant LodgingPhoto oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (_fingerprint != _currentFingerprint ||
        oldWidget.service != widget.service) {
      _fingerprint = _currentFingerprint;
      _load();
    }
  }

  Future<void> _load() async {
    final generation = ++_generation;
    _image?.dispose();
    _image = null;
    final url = _text(widget.photo['url']);
    setState(() => _loading = url.isNotEmpty);
    if (url.isEmpty) return;
    try {
      final bytes = await widget.service.loadLodgingImage(
        url,
        query: widget.accessQuery,
      );
      if (!mounted || generation != _generation) return;
      final buffer = await ui.ImmutableBuffer.fromUint8List(bytes);
      try {
        final descriptor = await ui.ImageDescriptor.encoded(buffer);
        try {
          if (descriptor.width * descriptor.height > 20000000) return;
          final codec = await descriptor.instantiateCodec(
            targetWidth: descriptor.width > 1600 ? 1600 : null,
          );
          try {
            final frame = await codec.getNextFrame();
            if (!mounted || generation != _generation) {
              frame.image.dispose();
              return;
            }
            setState(() => _image = frame.image);
          } finally {
            codec.dispose();
          }
        } finally {
          descriptor.dispose();
        }
      } finally {
        buffer.dispose();
      }
    } catch (_) {
      // The stay remains useful when its image is unavailable or access expires.
    } finally {
      if (mounted && generation == _generation) {
        setState(() => _loading = false);
      }
    }
  }

  @override
  void dispose() {
    _generation++;
    _auth?.removeListener(_accountChanged);
    _image?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Semantics(
    image: true,
    label: _text(widget.photo['altText']).isEmpty
        ? 'Property photo'
        : _text(widget.photo['altText']),
    child: Material(
      color: tripPeach,
      child: InkWell(
        onTap: widget.onTap,
        child: Stack(
          fit: StackFit.expand,
          children: [
            if (_image != null)
              RawImage(
                image: _image,
                fit: widget.fit,
                filterQuality: FilterQuality.medium,
              )
            else
              const Center(
                child: Icon(Icons.hotel_outlined, size: 42, color: tripRust),
              ),
            if (_loading)
              const Center(
                child: SizedBox(
                  width: 24,
                  height: 24,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
              ),
          ],
        ),
      ),
    ),
  );
}

/// Attribution for live Google Maps content, including linked photo authors.
class LodgingGoogleAttribution extends StatelessWidget {
  const LodgingGoogleAttribution({
    super.key,
    this.attributions = const [],
    this.googleMapsUri,
  });

  final List<Map<String, dynamic>> attributions;
  final String? googleMapsUri;

  @override
  Widget build(BuildContext context) => Wrap(
    spacing: 10,
    runSpacing: 4,
    crossAxisAlignment: WrapCrossAlignment.center,
    children: [
      const Text(
        'Google Maps',
        softWrap: false,
        style: TextStyle(
          fontFamily: 'Roboto',
          fontSize: 12,
          fontWeight: FontWeight.w400,
          color: Color(0xff5e5e5e),
        ),
      ),
      for (final author in attributions)
        if (_text(author['displayName']).isNotEmpty ||
            _text(author['provider']).isNotEmpty)
          _AttributionLink(
            label: _text(author['displayName']).isNotEmpty
                ? _text(author['displayName'])
                : _text(author['provider']),
            url: _text(author['uri']).isNotEmpty
                ? _text(author['uri'])
                : _text(author['providerUri']),
          ),
      if (_text(googleMapsUri).isNotEmpty)
        _AttributionLink(label: 'Photo source', url: googleMapsUri!),
    ],
  );
}

class _AttributionLink extends StatelessWidget {
  const _AttributionLink({required this.label, required this.url});
  final String label;
  final String url;
  @override
  Widget build(BuildContext context) => InkWell(
    onTap: url.isEmpty ? null : () => openTripLink(context, url),
    child: Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Text(
        label,
        style: TextStyle(
          color: tripMuted,
          fontSize: 12,
          decoration: url.isEmpty ? null : TextDecoration.underline,
        ),
      ),
    ),
  );
}

class TravelTripLodgingCard extends StatefulWidget {
  const TravelTripLodgingCard({
    super.key,
    required this.entry,
    this.actions,
    this.service,
    this.stopId,
    this.accessQuery = const {},
    this.draft = false,
  });

  final Map<String, dynamic> entry;
  final Widget? actions;
  final TravelCalendarService? service;
  final String? stopId;
  final Map<String, dynamic> accessQuery;
  final bool draft;

  @override
  State<TravelTripLodgingCard> createState() => _TravelTripLodgingCardState();
}

class _TravelTripLodgingCardState extends State<TravelTripLodgingCard> {
  Map<String, dynamic> _property = {};
  int _generation = 0;
  late String _fingerprint;
  AuthProvider? _auth;
  String? _identity;
  DialogRoute<void>? _detailsRoute;
  ValueNotifier<_StaySnapshot?>? _detailContent;

  String get _currentFingerprint => jsonEncode([
    widget.entry,
    widget.stopId,
    widget.accessQuery,
    widget.draft,
  ]);

  @override
  void initState() {
    super.initState();
    _fingerprint = _currentFingerprint;
    _loadProperty();
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final auth = context.read<AuthProvider?>();
    if (auth != _auth) {
      _auth?.removeListener(_accountChanged);
      _auth = auth;
      _identity = auth?.user?.id;
      _auth?.addListener(_accountChanged);
    }
  }

  void _accountChanged() {
    if (_identity == _auth?.user?.id) return;
    _identity = _auth?.user?.id;
    _generation++;
    _closeDetails();
    setState(() => _property = {});
  }

  @override
  void didUpdateWidget(covariant TravelTripLodgingCard oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (_fingerprint != _currentFingerprint ||
        oldWidget.service != widget.service) {
      _fingerprint = _currentFingerprint;
      _closeDetails();
      _loadProperty();
    }
  }

  Future<void> _loadProperty() async {
    final generation = ++_generation;
    setState(() => _property = {});
    final reference = tripMap(widget.entry['property']);
    if (reference['provider'] != 'google' ||
        _text(reference['placeId']).isEmpty ||
        widget.service == null ||
        _text(widget.stopId).isEmpty ||
        _text(widget.entry['id']).isEmpty) {
      return;
    }
    try {
      final response = await widget.service!.get(
        '/calendar/stops/${Uri.encodeComponent(widget.stopId!)}/lodging/${Uri.encodeComponent(widget.entry['id'] as String)}/property',
        query: {...widget.accessQuery, if (widget.draft) 'draft': true},
      );
      if (!mounted || generation != _generation) return;
      setState(() => _property = tripMap(response['property']));
      _detailContent?.value = _snapshot;
    } catch (_) {
      // Host-entered details remain available if Google Maps is unavailable.
    }
  }

  List<Map<String, dynamic>> get _photos => [
    ...tripMaps(widget.entry['photos']),
    ...tripMaps(
      _property['photos'],
    ).map((photo) => {...photo, '_google': true}),
  ];

  _StaySnapshot get _snapshot =>
      _StaySnapshot(widget.entry, _property, _photos);

  Map<String, dynamic> get _imageQuery => {
    ...widget.accessQuery,
    if (widget.draft) 'draft': true,
  };

  void _closeDetails() {
    final route = _detailsRoute;
    final content = _detailContent;
    _detailsRoute = null;
    _detailContent = null;
    if (route == null) return;
    if (SchedulerBinding.instance.schedulerPhase !=
        SchedulerPhase.persistentCallbacks) {
      content?.value = null;
    }
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (route.isActive) {
        content?.value = null;
        route.navigator?.removeRoute(route);
      }
    });
  }

  void _openDetails({int photoIndex = 0}) {
    if (_detailsRoute != null) return;
    final content = ValueNotifier<_StaySnapshot?>(_snapshot);
    _detailContent = content;
    final route = DialogRoute<void>(
      context: context,
      builder: (_) => ValueListenableBuilder<_StaySnapshot?>(
        valueListenable: content,
        builder: (_, snapshot, _) => snapshot == null
            ? const SizedBox.shrink()
            : _LodgingDetails(
                snapshot: snapshot,
                service: widget.service,
                accessQuery: _imageQuery,
                initialPhoto: photoIndex,
              ),
      ),
    );
    _detailsRoute = route;
    Navigator.of(context, rootNavigator: true).push<void>(route).whenComplete(
      () {
        if (_detailsRoute == route) {
          _detailsRoute = null;
          _detailContent = null;
        }
        content.dispose();
      },
    );
  }

  @override
  void dispose() {
    _generation++;
    _auth?.removeListener(_accountChanged);
    _closeDetails();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final photos = _photos;
    final snapshot = _snapshot;
    return Container(
      margin: const EdgeInsets.only(bottom: 14),
      decoration: BoxDecoration(
        color: tripPaper,
        border: Border.all(color: tripOutline),
        borderRadius: BorderRadius.circular(3),
      ),
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (photos.isNotEmpty && widget.service != null) ...[
            AspectRatio(
              aspectRatio: 16 / 9,
              child: Stack(
                fit: StackFit.expand,
                children: [
                  LodgingPhoto(
                    photo: photos.first,
                    service: widget.service!,
                    accessQuery: _imageQuery,
                    onTap: _openDetails,
                  ),
                  if (photos.length > 1)
                    Positioned(
                      right: 12,
                      bottom: 12,
                      child: FilledButton.tonalIcon(
                        onPressed: _openDetails,
                        style: FilledButton.styleFrom(
                          backgroundColor: tripCream,
                          foregroundColor: tripInk,
                          minimumSize: const Size(0, 34),
                          textStyle: const TextStyle(
                            fontFamily: AppTheme.bodyFont,
                            fontSize: 12,
                          ),
                        ),
                        icon: const Icon(
                          Icons.photo_library_outlined,
                          size: 16,
                        ),
                        label: Text('${photos.length} photos'),
                      ),
                    ),
                ],
              ),
            ),
            if (photos.first['_google'] == true)
              Padding(
                padding: const EdgeInsets.fromLTRB(18, 6, 18, 0),
                child: LodgingGoogleAttribution(
                  attributions: tripMaps(photos.first['authorAttributions']),
                  googleMapsUri: _text(photos.first['googleMapsUri']),
                ),
              ),
          ],
          Padding(
            padding: const EdgeInsets.all(18),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _StayHeading(snapshot: snapshot, onTap: _openDetails),
                const SizedBox(height: 12),
                _StayDates(entry: widget.entry),
                if (_text(widget.entry['description']).isNotEmpty) ...[
                  const SizedBox(height: 10),
                  Text(
                    _text(widget.entry['description']),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(color: tripMuted, height: 1.45),
                  ),
                ],
                const SizedBox(height: 8),
                _StayLinks(snapshot: snapshot, onDetails: _openDetails),
                if (_property.isNotEmpty &&
                    (photos.isEmpty ||
                        photos.first['_google'] != true ||
                        tripMaps(_property['attributions']).isNotEmpty))
                  Padding(
                    padding: const EdgeInsets.only(top: 4),
                    child: LodgingGoogleAttribution(
                      attributions: tripMaps(_property['attributions']),
                    ),
                  ),
                if (widget.actions != null) ...[
                  const Divider(height: 20),
                  widget.actions!,
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _StaySnapshot {
  const _StaySnapshot(this.entry, this.property, this.photos);
  final Map<String, dynamic> entry;
  final Map<String, dynamic> property;
  final List<Map<String, dynamic>> photos;
  String get name => _text(property['name']).isNotEmpty
      ? _text(property['name'])
      : _text(entry['name']).isNotEmpty
      ? _text(entry['name'])
      : 'A place to stay';
  String get hostLabel =>
      _text(entry['name']).isNotEmpty &&
          _text(entry['name']) != name &&
          _text(entry['name']) != 'Our stay'
      ? _text(entry['name'])
      : '';
  String get location => _text(entry['generalLocation']).isNotEmpty
      ? _text(entry['generalLocation'])
      : _text(property['address']);
  String get website => _text(entry['url']).isNotEmpty
      ? _text(entry['url'])
      : _text(property['websiteUrl']);
}

class _StayHeading extends StatelessWidget {
  const _StayHeading({required this.snapshot, this.onTap});
  final _StaySnapshot snapshot;
  final VoidCallback? onTap;
  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Wrap(
        spacing: 8,
        runSpacing: 6,
        children: [
          TripPill(
            snapshot.entry['kind'] == 'recommendation'
                ? 'Recommended stay'
                : 'Group stay',
            icon: Icons.bed_outlined,
          ),
          TripPill(
            _text(snapshot.entry['status']).isEmpty
                ? 'Considering'
                : toBeginningOfSentenceCase(
                    _text(snapshot.entry['status']).replaceAll('_', ' '),
                  ),
            color: tripMuted,
            background: tripPaper,
          ),
        ],
      ),
      const SizedBox(height: 12),
      InkWell(
        onTap: onTap,
        child: Text(
          snapshot.name,
          style: const TextStyle(
            color: tripInk,
            fontSize: 28,
            fontFamily: tripDisplayFont,
            height: 1.15,
          ),
        ),
      ),
      if (snapshot.hostLabel.isNotEmpty)
        Padding(
          padding: const EdgeInsets.only(top: 6),
          child: Text(
            snapshot.hostLabel,
            style: const TextStyle(
              color: tripMuted,
              fontStyle: FontStyle.italic,
            ),
          ),
        ),
      if (snapshot.location.isNotEmpty)
        Padding(
          padding: const EdgeInsets.only(top: 6),
          child: Text(
            snapshot.location,
            style: const TextStyle(color: tripMuted),
          ),
        ),
      if (snapshot.property['rating'] is num)
        Padding(
          padding: const EdgeInsets.only(top: 6),
          child: Row(
            children: [
              const Icon(Icons.star_rounded, color: tripRust, size: 16),
              const SizedBox(width: 4),
              Flexible(
                child: Text(
                  '${snapshot.property['rating']}${snapshot.property['ratingCount'] is num ? ' · ${snapshot.property['ratingCount']} reviews' : ''}',
                  style: const TextStyle(color: tripMuted, fontSize: 12),
                ),
              ),
            ],
          ),
        ),
    ],
  );
}

class _StayDates extends StatelessWidget {
  const _StayDates({required this.entry});
  final Map<String, dynamic> entry;
  @override
  Widget build(BuildContext context) {
    final start = DateTime.tryParse(_text(entry['checkInDate']));
    final end = DateTime.tryParse(_text(entry['checkOutDate']));
    final nights = start != null && end != null
        ? end.difference(start).inDays
        : 0;
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Padding(
          padding: EdgeInsets.only(top: 2),
          child: Icon(Icons.calendar_today_outlined, size: 16, color: tripRust),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            entry['datesTbd'] == true || start == null
                ? 'Stay dates to be decided'
                : '${DateFormat.MMMd().format(start)} – ${end == null ? 'TBD' : DateFormat.MMMd().format(end)}${nights > 0 ? ' · $nights ${nights == 1 ? 'night' : 'nights'}' : ''}',
            style: const TextStyle(color: tripInk, fontWeight: FontWeight.w600),
          ),
        ),
      ],
    );
  }
}

class _StayLinks extends StatelessWidget {
  const _StayLinks({required this.snapshot, this.onDetails});
  final _StaySnapshot snapshot;
  final VoidCallback? onDetails;
  @override
  Widget build(BuildContext context) => Wrap(
    spacing: 4,
    runSpacing: 2,
    children: [
      if (onDetails != null)
        TextButton(onPressed: onDetails, child: const Text('Stay details')),
      if (snapshot.website.isNotEmpty)
        TextButton.icon(
          onPressed: () => openTripLink(context, snapshot.website),
          icon: const Icon(Icons.open_in_new, size: 15),
          label: const Text('View property'),
        ),
      if (_text(snapshot.property['mapsUrl']).isNotEmpty)
        TextButton.icon(
          onPressed: () =>
              openTripLink(context, snapshot.property['mapsUrl'] as String),
          icon: const Icon(Icons.directions_outlined, size: 16),
          label: const Text('Directions'),
        ),
    ],
  );
}

class _LodgingDetails extends StatefulWidget {
  const _LodgingDetails({
    required this.snapshot,
    required this.service,
    required this.accessQuery,
    required this.initialPhoto,
  });
  final _StaySnapshot snapshot;
  final TravelCalendarService? service;
  final Map<String, dynamic> accessQuery;
  final int initialPhoto;
  @override
  State<_LodgingDetails> createState() => _LodgingDetailsState();
}

class _LodgingDetailsState extends State<_LodgingDetails> {
  late final PageController _pages;
  late int _index;
  @override
  void initState() {
    super.initState();
    _index = widget.initialPhoto;
    _pages = PageController(initialPage: _index);
  }

  @override
  void dispose() {
    _pages.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final snapshot = widget.snapshot;
    final entry = snapshot.entry;
    final photos = snapshot.photos;
    final photo = photos.isEmpty
        ? <String, dynamic>{}
        : photos[_index.clamp(0, photos.length - 1)];
    final private = tripMap(entry['privateDetails']);
    final cost = tripMap(entry['cost']);
    return Dialog(
      insetPadding: const EdgeInsets.all(20),
      backgroundColor: tripPaper,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 760),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(20, 10, 8, 8),
              child: Row(
                children: [
                  const Expanded(
                    child: Text(
                      'Your stay',
                      style: TextStyle(
                        fontFamily: tripDisplayFont,
                        fontSize: 24,
                        color: tripInk,
                      ),
                    ),
                  ),
                  IconButton(
                    tooltip: 'Close stay details',
                    onPressed: () => Navigator.pop(context),
                    icon: const Icon(Icons.close),
                  ),
                ],
              ),
            ),
            Flexible(
              child: SingleChildScrollView(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    if (photos.isNotEmpty && widget.service != null) ...[
                      AspectRatio(
                        aspectRatio: 16 / 10,
                        child: PageView.builder(
                          controller: _pages,
                          itemCount: photos.length,
                          onPageChanged: (index) =>
                              setState(() => _index = index),
                          itemBuilder: (_, index) => LodgingPhoto(
                            photo: photos[index],
                            service: widget.service!,
                            accessQuery: widget.accessQuery,
                            fit: BoxFit.contain,
                          ),
                        ),
                      ),
                      if (photos.length > 1)
                        Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            IconButton(
                              tooltip: 'Previous photo',
                              onPressed: _index == 0
                                  ? null
                                  : () => _pages.previousPage(
                                      duration: const Duration(
                                        milliseconds: 200,
                                      ),
                                      curve: Curves.easeOut,
                                    ),
                              icon: const Icon(Icons.chevron_left),
                            ),
                            Text(
                              '${_index + 1} / ${photos.length}',
                              style: const TextStyle(
                                color: tripMuted,
                                fontSize: 12,
                              ),
                            ),
                            IconButton(
                              tooltip: 'Next photo',
                              onPressed: _index == photos.length - 1
                                  ? null
                                  : () => _pages.nextPage(
                                      duration: const Duration(
                                        milliseconds: 200,
                                      ),
                                      curve: Curves.easeOut,
                                    ),
                              icon: const Icon(Icons.chevron_right),
                            ),
                          ],
                        ),
                      if (photo['_google'] == true)
                        Padding(
                          padding: const EdgeInsets.fromLTRB(20, 6, 20, 0),
                          child: LodgingGoogleAttribution(
                            attributions: tripMaps(photo['authorAttributions']),
                            googleMapsUri: _text(photo['googleMapsUri']),
                          ),
                        ),
                    ],
                    Padding(
                      padding: const EdgeInsets.all(20),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          _StayHeading(snapshot: snapshot),
                          const SizedBox(height: 16),
                          _StayDates(entry: entry),
                          if (_text(entry['checkInTime']).isNotEmpty ||
                              _text(entry['checkOutTime']).isNotEmpty)
                            Padding(
                              padding: const EdgeInsets.only(top: 8),
                              child: Text(
                                'Check-in ${_text(entry['checkInTime']).isEmpty ? 'TBD' : entry['checkInTime']} · Checkout ${_text(entry['checkOutTime']).isEmpty ? 'TBD' : entry['checkOutTime']} · ${entry['timeZone'] ?? 'UTC'}',
                                style: const TextStyle(color: tripMuted),
                              ),
                            ),
                          if (_text(entry['description']).isNotEmpty)
                            Padding(
                              padding: const EdgeInsets.only(top: 16),
                              child: TripDescription(
                                entry['description'] as String,
                              ),
                            ),
                          if (_text(snapshot.property['address']).isNotEmpty &&
                              snapshot.property['address'] != snapshot.location)
                            Padding(
                              padding: const EdgeInsets.only(top: 12),
                              child: Text(
                                snapshot.property['address'] as String,
                                style: const TextStyle(color: tripMuted),
                              ),
                            ),
                          if (cost.isNotEmpty)
                            Padding(
                              padding: const EdgeInsets.only(top: 12),
                              child: Text(
                                'Estimate: ${cost['currency']} ${cost['amount']} ${switch (cost['basis']) {
                                  'night' => 'per night',
                                  'person' => 'per person',
                                  _ => 'total',
                                }}',
                                style: const TextStyle(color: tripMuted),
                              ),
                            ),
                          if (_text(entry['roomsOrBeds']).isNotEmpty)
                            Padding(
                              padding: const EdgeInsets.only(top: 12),
                              child: Text(entry['roomsOrBeds'] as String),
                            ),
                          if (_text(private['address']).isNotEmpty ||
                              _text(
                                private['arrivalInstructions'],
                              ).isNotEmpty) ...[
                            const Divider(height: 28),
                            const Text(
                              'Arrival details',
                              style: TextStyle(
                                fontWeight: FontWeight.w600,
                                color: tripInk,
                              ),
                            ),
                            if (_text(private['address']).isNotEmpty)
                              Padding(
                                padding: const EdgeInsets.only(top: 8),
                                child: Text(private['address'] as String),
                              ),
                            if (_text(
                              private['arrivalInstructions'],
                            ).isNotEmpty)
                              Padding(
                                padding: const EdgeInsets.only(top: 8),
                                child: Text(
                                  private['arrivalInstructions'] as String,
                                ),
                              ),
                          ],
                          if (_text(entry['hostNotes']).isNotEmpty)
                            Padding(
                              padding: const EdgeInsets.only(top: 12),
                              child: Text(
                                'Host notes: ${entry['hostNotes']}',
                                style: const TextStyle(color: tripMuted),
                              ),
                            ),
                          const SizedBox(height: 12),
                          _StayLinks(snapshot: snapshot),
                          if (snapshot.property.isNotEmpty)
                            Padding(
                              padding: const EdgeInsets.only(top: 8),
                              child: LodgingGoogleAttribution(
                                attributions: tripMaps(
                                  snapshot.property['attributions'],
                                ),
                              ),
                            ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
