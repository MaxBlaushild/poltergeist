import 'dart:typed_data';
import 'dart:ui' as ui;

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_trip_design.dart';

class TripProtectedHero extends StatefulWidget {
  const TripProtectedHero({
    super.key,
    required this.hero,
    required this.service,
    this.accessQuery = const {},
  });
  final Map<String, dynamic> hero;
  final TravelCalendarService service;
  final Map<String, dynamic> accessQuery;

  @override
  State<TripProtectedHero> createState() => _TripProtectedHeroState();
}

class _TripProtectedHeroState extends State<TripProtectedHero> {
  Uint8List? _bytes;
  int _generation = 0;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void didUpdateWidget(covariant TripProtectedHero oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.hero['mediaId'] != widget.hero['mediaId'] ||
        oldWidget.hero['url'] != widget.hero['url']) {
      _load();
    }
  }

  Future<void> _load() async {
    final generation = ++_generation;
    setState(() => _bytes = null);
    final url = widget.hero['url'] as String?;
    if (url == null || url.isEmpty) return;
    try {
      final bytes = await widget.service.loadHero(
        url,
        query: widget.accessQuery,
      );
      if (mounted && generation == _generation) setState(() => _bytes = bytes);
    } catch (_) {
      // Never keep a previously accessible image after a denied request.
      if (mounted && generation == _generation) setState(() => _bytes = null);
    }
  }

  @override
  Widget build(BuildContext context) =>
      TripHeroCanvas(bytes: _bytes, hero: widget.hero);
}

class TripHeroCanvas extends StatefulWidget {
  const TripHeroCanvas({super.key, this.bytes, this.hero = const {}});
  final Uint8List? bytes;
  final Map<String, dynamic> hero;

  @override
  State<TripHeroCanvas> createState() => _TripHeroCanvasState();
}

class _TripHeroCanvasState extends State<TripHeroCanvas> {
  ui.Image? _image;
  int _generation = 0;

  @override
  void initState() {
    super.initState();
    _decode();
  }

  @override
  void didUpdateWidget(covariant TripHeroCanvas oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.bytes != widget.bytes) _decode();
  }

  Future<void> _decode() async {
    final generation = ++_generation;
    _image?.dispose();
    _image = null;
    if (widget.bytes == null) return;
    try {
      final buffer = await ui.ImmutableBuffer.fromUint8List(widget.bytes!);
      final descriptor = await ui.ImageDescriptor.encoded(buffer);
      if (descriptor.width * descriptor.height > 20000000) {
        descriptor.dispose();
        buffer.dispose();
        return;
      }
      final codec = await descriptor.instantiateCodec(
        targetWidth: descriptor.width > 1800 ? 1800 : null,
      );
      final frame = await codec.getNextFrame();
      codec.dispose();
      descriptor.dispose();
      buffer.dispose();
      if (!mounted || generation != _generation) {
        frame.image.dispose();
        return;
      }
      setState(() => _image = frame.image);
    } catch (_) {
      /* Keep the designed fallback if decoding fails. */
    }
  }

  @override
  void dispose() {
    _image?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Semantics(
    image: true,
    label: _image == null
        ? 'Travel illustration with warm hills and a winding route'
        : widget.hero['altText'] as String? ?? 'Trip cover image',
    child: ExcludeSemantics(
      child: ClipRect(
        child: CustomPaint(
          painter: _TripHeroPainter(_image, widget.hero),
          child: const SizedBox.expand(),
        ),
      ),
    ),
  );
}

class _TripHeroPainter extends CustomPainter {
  _TripHeroPainter(this.image, this.hero);
  final ui.Image? image;
  final Map<String, dynamic> hero;

  double _value(String key, double fallback) =>
      (hero[key] as num?)?.toDouble() ?? fallback;

  @override
  void paint(Canvas canvas, Size size) {
    if (image != null) {
      final crop =
          Rect.fromLTWH(
            _value('cropX', 0) * image!.width,
            _value('cropY', 0) * image!.height,
            _value('cropWidth', 1) * image!.width,
            _value('cropHeight', 1) * image!.height,
          ).intersect(
            Rect.fromLTWH(
              0,
              0,
              image!.width.toDouble(),
              image!.height.toDouble(),
            ),
          );
      if (crop.isEmpty) return;
      final fitted = applyBoxFit(BoxFit.cover, crop.size, size);
      final source = Alignment(
        _value('focalX', .5) * 2 - 1,
        _value('focalY', .5) * 2 - 1,
      ).inscribe(fitted.source, crop);
      canvas.drawImageRect(
        image!,
        source,
        Offset.zero & size,
        Paint()..filterQuality = FilterQuality.medium,
      );
      return;
    }
    final rect = Offset.zero & size;
    canvas.drawRect(
      rect,
      Paint()
        ..shader = const LinearGradient(
          colors: [tripPeach, tripCream],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ).createShader(rect),
    );
    canvas.drawCircle(
      Offset(size.width * .76, size.height * .31),
      size.height * .19,
      Paint()..color = tripOrange,
    );
    for (var i = 0; i < 5; i++) {
      final path = Path()
        ..moveTo(-30, size.height * (.5 + i * .12))
        ..cubicTo(
          size.width * .25,
          size.height * (.08 + i * .15),
          size.width * .55,
          size.height * (1.14 - i * .025),
          size.width + 30,
          size.height * (.48 + i * .16),
        )
        ..lineTo(size.width + 30, size.height + 10)
        ..lineTo(-30, size.height + 10)
        ..close();
      canvas.drawPath(
        path,
        Paint()
          ..color = [
            Color.lerp(tripPeach, tripInk, .2)!,
            Color.lerp(tripPeach, tripInk, .4)!,
            Color.lerp(tripPeach, tripInk, .6)!,
            Color.lerp(tripPeach, tripInk, .8)!,
            Color.lerp(tripPeach, tripInk, 1)!,
          ][i],
      );
    }
    final route = Path()
      ..moveTo(size.width * .2, size.height * .98)
      ..cubicTo(
        size.width * .45,
        size.height * .7,
        size.width * .12,
        size.height * .7,
        size.width * .38,
        size.height * .48,
      )
      ..quadraticBezierTo(
        size.width * .48,
        size.height * .4,
        size.width * .62,
        size.height * .55,
      );
    canvas.drawPath(
      route,
      Paint()
        ..color = tripCream
        ..strokeWidth = 2.5
        ..style = PaintingStyle.stroke,
    );
    canvas.drawCircle(
      Offset(size.width * .62, size.height * .55),
      5,
      Paint()..color = tripCream,
    );
  }

  @override
  bool shouldRepaint(covariant _TripHeroPainter old) =>
      old.image != image || old.hero != hero;
}

class TravelTripHeroEditor extends StatefulWidget {
  const TravelTripHeroEditor({
    super.key,
    required this.stop,
    required this.service,
    this.accessQuery = const {},
  });
  final Map<String, dynamic> stop;
  final TravelCalendarService service;
  final Map<String, dynamic> accessQuery;

  @override
  State<TravelTripHeroEditor> createState() => _TravelTripHeroEditorState();
}

class _TravelTripHeroEditorState extends State<TravelTripHeroEditor> {
  late final Map<String, dynamic> _original;
  late final TextEditingController _alt;
  Uint8List? _bytes;
  String? _filename;
  double _focalX = .5;
  double _focalY = .5;
  double _zoom = 1;
  bool _cropChanged = false;
  bool _remove = false;
  bool _busy = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _original = tripMap(tripDraft(widget.stop)['hero']);
    _alt = TextEditingController(text: _original['altText'] as String? ?? '');
    _focalX = (_original['focalX'] as num?)?.toDouble() ?? .5;
    _focalY = (_original['focalY'] as num?)?.toDouble() ?? .5;
    _zoom = (1 / ((_original['cropWidth'] as num?)?.toDouble() ?? 1)).clamp(
      1,
      4,
    );
    _load();
  }

  @override
  void dispose() {
    _alt.dispose();
    super.dispose();
  }

  Map<String, dynamic> get _metadata {
    final width = 1 / _zoom;
    return {
      ..._original,
      'altText': _alt.text.trim(),
      'focalX': _focalX,
      'focalY': _focalY,
      'cropX': _cropChanged
          ? (_focalX - width / 2).clamp(0, 1 - width)
          : _original['cropX'] ?? 0,
      'cropY': _cropChanged
          ? (_focalY - width / 2).clamp(0, 1 - width)
          : _original['cropY'] ?? 0,
      'cropWidth': _cropChanged ? width : _original['cropWidth'] ?? 1,
      'cropHeight': _cropChanged ? width : _original['cropHeight'] ?? 1,
    };
  }

  Future<void> _load() async {
    final url = _original['url'] as String?;
    if (url == null) return;
    try {
      final bytes = await widget.service.loadHero(
        url,
        query: widget.accessQuery,
      );
      if (mounted) setState(() => _bytes = bytes);
    } catch (_) {}
  }

  Future<void> _pick() async {
    final selection = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['jpg', 'jpeg', 'png'],
      withData: true,
    );
    if (selection == null || !mounted) return;
    final file = selection.files.single;
    if (file.bytes == null || file.size > 8 * 1024 * 1024) {
      setState(() => _error = 'Choose a JPEG or PNG image up to 8 MB.');
      return;
    }
    setState(() {
      _bytes = file.bytes;
      _filename = file.name;
      _remove = false;
      _error = null;
      _zoom = 1;
      _focalX = .5;
      _focalY = .5;
      _cropChanged = true;
    });
  }

  Future<void> _save() async {
    if (!_remove && _alt.text.trim().isEmpty) {
      setState(
        () => _error =
            'Describe the cover image for people using a screen reader.',
      );
      return;
    }
    if (!_remove && _bytes == null && _original.isEmpty) {
      setState(() => _error = 'Choose an image first.');
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      if (_remove) {
        await saveTripDraft(widget.service, widget.stop, {'hero': null});
      } else if (_filename != null) {
        final metadata = _metadata;
        await widget.service.uploadHero(
          widget.stop['id'] as String,
          _bytes!,
          _filename!,
          altText: _alt.text.trim(),
          focalX: _focalX,
          focalY: _focalY,
          revision: widget.stop['revision'] as int,
          cropX: (metadata['cropX'] as num).toDouble(),
          cropY: (metadata['cropY'] as num).toDouble(),
          cropWidth: (metadata['cropWidth'] as num).toDouble(),
          cropHeight: (metadata['cropHeight'] as num).toDouble(),
        );
      } else {
        await saveTripDraft(widget.service, widget.stop, {'hero': _metadata});
      }
      if (mounted) Navigator.pop(context, true);
    } catch (error) {
      if (mounted) setState(() => _error = tripError(error));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) => TripFormDialog(
    title: 'Make it feel like your trip',
    busy: _busy,
    error: _error,
    onSave: _save,
    saveLabel: 'Save cover draft',
    children: [
      const Text(
        'Your published cover stays in place until you publish the draft. Location metadata is removed from uploads.',
      ),
      const SizedBox(height: 16),
      ClipRRect(
        borderRadius: BorderRadius.circular(3),
        child: AspectRatio(
          aspectRatio: 16 / 9,
          child: TripHeroCanvas(
            bytes: _remove ? null : _bytes,
            hero: _metadata,
          ),
        ),
      ),
      const SizedBox(height: 12),
      Wrap(
        spacing: 8,
        children: [
          OutlinedButton.icon(
            onPressed: _busy ? null : _pick,
            icon: const Icon(Icons.add_photo_alternate_outlined),
            label: Text(_bytes == null ? 'Choose image' : 'Replace image'),
          ),
          if (_original.isNotEmpty || _bytes != null)
            TextButton(
              onPressed: _busy
                  ? null
                  : () => setState(() => _remove = !_remove),
              child: Text(_remove ? 'Keep image' : 'Remove image'),
            ),
        ],
      ),
      if (!_remove) ...[
        TextField(
          controller: _alt,
          maxLength: 300,
          decoration: const InputDecoration(
            labelText: 'Image description',
            hintText: 'Sunset over the rooftops of Lisbon',
          ),
        ),
        const Text('Crop zoom'),
        Slider(
          value: _zoom,
          min: 1,
          max: 4,
          label: '${_zoom.toStringAsFixed(1)}×',
          onChanged: (value) => setState(() {
            _zoom = value;
            _cropChanged = true;
          }),
        ),
        const Text('Focal point · left to right'),
        Slider(
          value: _focalX,
          onChanged: (value) => setState(() {
            _focalX = value;
            _cropChanged = true;
          }),
        ),
        const Text('Focal point · top to bottom'),
        Slider(
          value: _focalY,
          onChanged: (value) => setState(() {
            _focalY = value;
            _cropChanged = true;
          }),
        ),
      ],
    ],
  );
}
