import 'dart:async';
import 'package:flutter/material.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/screens/library_document_screen.dart';
import 'package:travel_angels/screens/library_screen.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/document_service.dart';

/// Personal references never expose a guide to another traveler on the trip.
class TravelTripLibrary extends StatefulWidget {
  const TravelTripLibrary({
    super.key,
    required this.tripId,
    required this.userId,
    this.service,
  });
  final String tripId;
  final String userId;
  final DocumentService? service;

  @override
  State<TravelTripLibrary> createState() => _TravelTripLibraryState();
}

class _TravelTripLibraryState extends State<TravelTripLibrary>
    with WidgetsBindingObserver {
  late final _service =
      widget.service ?? DocumentService(APIClient(ApiConstants.baseUrl));
  List<Document> _documents = [];
  bool _open = false;
  bool _loading = false;
  String? _error;
  int _generation = 0;
  Timer? _timer;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _timer = Timer.periodic(const Duration(seconds: 30), (_) {
      if (_open && ModalRoute.of(context)?.isCurrent == true) _load();
    });
  }

  @override
  void dispose() {
    _generation++;
    _timer?.cancel();
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed && _open) _load();
  }

  Future<void> _load() async {
    final generation = ++_generation;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final rows = await _service.getTripDocuments(widget.tripId);
      if (!mounted || generation != _generation) return;
      setState(() {
        _documents = rows.map(Document.fromJson).toList();
        _loading = false;
      });
    } catch (_) {
      if (!mounted || generation != _generation) return;
      setState(() {
        _documents = [];
        _loading = false;
        _error = 'Your trip guides could not be loaded.';
      });
    }
  }

  Future<void> _read(Document document) async {
    await Navigator.push<void>(
      context,
      MaterialPageRoute(
        builder: (_) => LibraryDocumentScreen(
          document: document,
          userId: widget.userId,
          service: _service,
        ),
      ),
    );
    if (mounted && _open) await _load();
  }

  Future<void> _browse() async {
    await Navigator.push<void>(
      context,
      MaterialPageRoute(
        builder: (_) => Scaffold(
          appBar: AppBar(title: const Text('Library')),
          body: LibraryScreen(userId: widget.userId, service: _service),
        ),
      ),
    );
    if (mounted && _open) await _load();
  }

  @override
  Widget build(BuildContext context) => ExpansionTile(
    key: const ValueKey('trip-library'),
    tilePadding: EdgeInsets.zero,
    childrenPadding: const EdgeInsets.only(bottom: 24),
    title: Text(
      'Your trip library',
      style: Theme.of(context).textTheme.headlineSmall,
    ),
    subtitle: const Text('Guides you’ve saved for this trip'),
    onExpansionChanged: (open) {
      setState(() {
        _open = open;
        if (!open) {
          _generation++;
          _documents = [];
        }
      });
      if (open) _load();
    },
    children: [
      const Align(
        alignment: Alignment.centerLeft,
        child: Text(
          'This collection is just for you. Each guide keeps the audience chosen by its owner.',
        ),
      ),
      const SizedBox(height: 12),
      if (_loading) const LinearProgressIndicator(minHeight: 2),
      if (_error != null)
        Row(
          children: [
            Expanded(child: Text(_error!)),
            TextButton(onPressed: _load, child: const Text('Retry')),
          ],
        ),
      if (!_loading && _error == null && _documents.isEmpty)
        const Padding(
          padding: EdgeInsets.symmetric(vertical: 16),
          child: Align(
            alignment: Alignment.centerLeft,
            child: Text(
              'Open a guide in the Library and choose “Save to a trip” to keep it here.',
            ),
          ),
        ),
      for (final document in _documents)
        ListTile(
          contentPadding: EdgeInsets.zero,
          leading: Icon(
            document.format == 'spreadsheet'
                ? Icons.table_chart_outlined
                : Icons.menu_book_outlined,
          ),
          title: Text(document.title),
          subtitle: Text(
            document.destination.isNotEmpty
                ? document.destination
                : document.user?.name ?? 'Travel guide',
          ),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => _read(document),
        ),
      Align(
        alignment: Alignment.centerLeft,
        child: TextButton.icon(
          onPressed: _browse,
          icon: const Icon(Icons.auto_stories_outlined, size: 18),
          label: const Text('Browse library'),
        ),
      ),
    ],
  );
}
