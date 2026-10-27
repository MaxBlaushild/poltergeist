import 'dart:async';

import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/document_service.dart';
import 'package:travel_angels/services/friend_service.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/edit_guide_flow.dart';
import 'package:travel_angels/widgets/library_document_content.dart';
import 'package:travel_angels/widgets/library_document_dialogs.dart';
import 'package:travel_angels/widgets/video_preview_dialog.dart';

class LibraryDocumentScreen extends StatefulWidget {
  const LibraryDocumentScreen({
    super.key,
    required this.document,
    this.service,
    this.friendService,
    this.calendarService,
    this.userId,
    this.onEditDocument,
  });
  final Document document;
  final DocumentService? service;
  final FriendService? friendService;
  final TravelCalendarService? calendarService;
  final String? userId;
  final Future<bool?> Function(BuildContext, Document)? onEditDocument;
  @override
  State<LibraryDocumentScreen> createState() => _LibraryDocumentScreenState();
}

class _LibraryDocumentScreenState extends State<LibraryDocumentScreen>
    with WidgetsBindingObserver {
  late final DocumentService _service =
      widget.service ?? DocumentService(APIClient(ApiConstants.baseUrl));
  Document? _document;
  AuthProvider? _auth;
  String? _identity;
  String? _error;
  Timer? _timer;
  bool _loading = true;
  bool _busy = false;
  bool _sensitiveDialog = false;
  Route<dynamic>? _sensitiveRoute;
  int _generation = 0;
  List<String> _tripNames = [];
  bool get _owner =>
      _identity?.isNotEmpty == true && _document?.userId == _identity;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _load();
    _timer = Timer.periodic(const Duration(seconds: 30), (_) {
      if (!_busy && !_loading) _load(foreground: false);
    });
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final auth = context.read<AuthProvider?>();
    if (auth != _auth) {
      _auth?.removeListener(_authChanged);
      _auth = auth;
      _auth?.addListener(_authChanged);
    }
    _identity = _auth != null ? _auth!.user?.id : widget.userId;
  }

  @override
  void didUpdateWidget(covariant LibraryDocumentScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.userId != widget.userId) _authChanged();
  }

  void _authChanged() {
    final identity = _auth != null ? _auth!.user?.id : widget.userId;
    if (!mounted || identity == _identity) return;
    _identity = identity;
    _generation++;
    _closeSensitiveDialog();
    setState(() {
      _document = null;
      _tripNames = [];
      _loading = false;
      _error = identity == null ? 'Sign in to read this guide.' : null;
    });
    if (identity != null) _load();
  }

  void _closeSensitiveDialog() {
    if (!_sensitiveDialog || !mounted) return;
    _sensitiveDialog = false;
    final route = _sensitiveRoute;
    _sensitiveRoute = null;
    // Access loss must dismiss even a dialog whose save temporarily blocks back.
    if (route?.isActive == true) Navigator.of(context).removeRoute(route!);
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed && !_busy) _load(foreground: false);
  }

  @override
  void dispose() {
    _timer?.cancel();
    _auth?.removeListener(_authChanged);
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  Future<void> _load({bool foreground = true}) async {
    final generation = ++_generation;
    if (foreground) {
      setState(() {
        _loading = true;
        _error = null;
      });
    }
    try {
      final data = await _service.getDocument(widget.document.id);
      if (!mounted || generation != _generation) return;
      final document = Document.fromJson(data);
      final lostOwnership = _owner && document.userId != _identity;
      if (lostOwnership) _closeSensitiveDialog();
      setState(() {
        _document = document;
        _error = null;
        _loading = false;
      });
      _loadReferences(document, generation);
    } catch (error) {
      if (!mounted || generation != _generation) return;
      _closeSensitiveDialog();
      setState(() {
        _document = null;
        _tripNames = [];
        _error = libraryError(
          error,
          'This guide could not be loaded. Please try again.',
        );
        _loading = false;
      });
    }
  }

  Future<void> _loadReferences(Document document, int generation) async {
    if (document.savedTripIds.isEmpty) {
      if (mounted) setState(() => _tripNames = []);
      return;
    }
    try {
      final trips = await _service.getDocumentTrips(document.id);
      if (mounted && generation == _generation) {
        setState(
          () => _tripNames = trips
              .where((trip) => trip['isSaved'] == true)
              .map((trip) => trip['title']?.toString() ?? 'Trip')
              .toList(),
        );
      }
    } catch (error) {
      if (mounted && generation == _generation && libraryAccessDenied(error)) {
        _closeSensitiveDialog();
        setState(() {
          _document = null;
          _tripNames = [];
          _error = libraryError(error);
        });
      }
    }
  }

  Future<void> _save() async {
    final document = _document;
    if (document == null) return;
    final identity = _identity;
    setState(() => _busy = true);
    try {
      final data = await _service.setDocumentSaved(
        document.id,
        !document.isSaved,
      );
      if (mounted && identity == _identity && _document != null) {
        setState(() => _document = Document.fromJson(data));
      }
    } catch (error) {
      if (identity == _identity) {
        _handleError(error, 'Your saved guides could not be updated.');
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  void _handleError(Object error, String fallback) {
    if (!mounted) return;
    if (libraryAccessDenied(error)) {
      _closeSensitiveDialog();
      setState(() {
        _document = null;
        _tripNames = [];
        _error = libraryError(error);
      });
    } else {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(libraryError(error, fallback))));
    }
  }

  Future<void> _share() async {
    final document = _document;
    if (document == null || !_owner) return;
    await _dialog(
      LibraryShareDialog(
        document: document,
        service: _service,
        friendService: widget.friendService,
      ),
    );
  }

  Future<void> _trips() async {
    final document = _document;
    if (document == null) return;
    await _dialog(
      LibraryTripPickerDialog(document: document, service: _service),
    );
  }

  Future<void> _dialog(Widget dialog) async {
    final identity = _identity;
    _sensitiveDialog = true;
    final updated = await showDialog<Document>(
      context: context,
      builder: (dialogContext) {
        _sensitiveRoute = ModalRoute.of(dialogContext);
        return dialog;
      },
    );
    _sensitiveDialog = false;
    _sensitiveRoute = null;
    if (!mounted || identity != _identity || _document == null) return;
    if (updated != null) {
      setState(() => _document = updated);
    }
    // A multi-trip save can partially succeed before the dialog is cancelled.
    await _load(foreground: false);
  }

  Future<void> _edit() async {
    final document = _document;
    if (document == null || !_owner) return;
    await (widget.onEditDocument ?? editGuide)(context, document);
    if (mounted) await _load();
  }

  Future<void> _delete() async {
    final document = _document;
    if (document == null || !_owner) return;
    final identity = _identity;
    _sensitiveDialog = true;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) {
        _sensitiveRoute = ModalRoute.of(dialogContext);
        return AlertDialog(
          title: const Text('Delete this guide?'),
          content: const Text(
            'This removes it from Travel Angels and from saved libraries and trip references. The original source file is kept.',
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(dialogContext, false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(dialogContext, true),
              child: const Text('Delete guide'),
            ),
          ],
        );
      },
    );
    _sensitiveDialog = false;
    _sensitiveRoute = null;
    if (!mounted ||
        confirmed != true ||
        !_owner ||
        _document?.id != document.id) {
      return;
    }
    setState(() => _busy = true);
    try {
      await _service.deleteDocument(document.id);
      if (mounted && identity == _identity) Navigator.pop(context, true);
    } catch (error) {
      if (identity == _identity) {
        _handleError(error, 'This guide could not be deleted.');
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final document = _document;
    return Scaffold(
      appBar: AppBar(
        title: Text(
          'Library',
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
          style: Theme.of(context).textTheme.titleMedium,
        ),
        backgroundColor: AppTheme.paper,
        actions: [
          if (!_loading && document != null) ...[
            IconButton(
              key: const ValueKey('library-reader-save'),
              tooltip: document.isSaved
                  ? 'Remove from saved guides'
                  : 'Save guide',
              isSelected: document.isSaved,
              onPressed: _busy ? null : _save,
              iconSize: 20,
              icon: const Icon(Icons.bookmark_border),
              selectedIcon: const Icon(Icons.bookmark, color: AppTheme.rust),
            ),
            IconButton(
              key: const ValueKey('library-reader-trip'),
              tooltip: 'Save to a trip',
              onPressed: _busy ? null : _trips,
              iconSize: 20,
              icon: const Icon(Icons.calendar_today_outlined),
            ),
            if (librarySafeUri(document.link) case final source?)
              IconButton(
                key: const ValueKey('library-reader-source'),
                tooltip: 'Open source',
                onPressed: () => openLibrarySource(context, source.toString()),
                iconSize: 20,
                icon: const Icon(Icons.open_in_new),
              ),
            _options(),
          ] else
            IconButton(
              tooltip: 'Refresh guide',
              onPressed: _loading || _busy ? null : () => _load(),
              iconSize: 20,
              icon: const Icon(Icons.refresh),
            ),
          const SizedBox(width: 4),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : document == null
          ? Center(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      _error ?? 'This guide is unavailable.',
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 16),
                    OutlinedButton(
                      onPressed: () => _load(),
                      child: const Text('Retry'),
                    ),
                  ],
                ),
              ),
            )
          : RefreshIndicator(
              onRefresh: () => _load(),
              child: SingleChildScrollView(
                physics: const AlwaysScrollableScrollPhysics(),
                child: Center(
                  child: ConstrainedBox(
                    constraints: const BoxConstraints(maxWidth: 1200),
                    child: Padding(
                      padding: EdgeInsets.all(
                        MediaQuery.sizeOf(context).width < 600 ? 16 : 24,
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          if (libraryDestination(document).isNotEmpty) ...[
                            Text(
                              libraryDestination(document),
                              style: Theme.of(context).textTheme.labelMedium
                                  ?.copyWith(color: AppTheme.rust),
                            ),
                            const SizedBox(height: 6),
                          ],
                          Text(
                            document.title,
                            style: Theme.of(context).textTheme.headlineLarge,
                          ),
                          const SizedBox(height: 8),
                          Wrap(
                            spacing: 12,
                            runSpacing: 4,
                            children: [
                              Text(
                                libraryAuthor(document),
                                style: Theme.of(context).textTheme.bodySmall,
                              ),
                              if (document.updatedAt != null)
                                Text(
                                  'Updated ${DateFormat.yMMMd().format(document.updatedAt!.toLocal())}',
                                  style: Theme.of(context).textTheme.bodySmall,
                                ),
                            ],
                          ),
                          const SizedBox(height: 16),
                          _reader(document),
                          if (document.personalNote.trim().isNotEmpty) ...[
                            const SizedBox(height: 20),
                            Text(
                              _owner
                                  ? 'Your note'
                                  : 'A note from ${libraryAuthor(document)}',
                              style: Theme.of(context).textTheme.labelMedium
                                  ?.copyWith(color: AppTheme.rust),
                            ),
                            const SizedBox(height: 6),
                            Text(
                              document.personalNote,
                              style: Theme.of(context).textTheme.bodyMedium,
                            ),
                          ],
                          if (_tripNames.isNotEmpty) ...[
                            const SizedBox(height: 16),
                            Text(
                              'Saved to ${_tripNames.join(' · ')}',
                              style: Theme.of(context).textTheme.bodySmall,
                            ),
                          ],
                          const SizedBox(height: 24),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
            ),
    );
  }

  Widget _reader(Document document) {
    final format = libraryFormat(document);
    final content = document.content?.trim() ?? '';
    final source = librarySafeUri(document.link);
    final imported =
        document.provider != CloudDocumentProvider.internal &&
        content.isNotEmpty;
    final linked = content.isEmpty && source != null;
    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: AppTheme.outline),
        borderRadius: BorderRadius.circular(3),
      ),
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Container(
            color: libraryFormatColor(format),
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            child: Wrap(
              spacing: 16,
              runSpacing: 8,
              alignment: WrapAlignment.spaceBetween,
              children: [
                Text(
                  librarySource(document),
                  style: Theme.of(context).textTheme.bodySmall,
                ),
                Text(
                  linked
                      ? 'Linked original'
                      : imported
                      ? 'Saved copy'
                      : 'Guide',
                  style: Theme.of(context).textTheme.labelSmall,
                ),
              ],
            ),
          ),
          Padding(
            padding: EdgeInsets.all(
              MediaQuery.sizeOf(context).width < 600 ? 16 : 24,
            ),
            child: SelectionArea(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  if (format == 'video' && source != null)
                    OutlinedButton.icon(
                      onPressed: () async {
                        _sensitiveDialog = true;
                        await showDialog<void>(
                          context: context,
                          builder: (dialogContext) {
                            _sensitiveRoute = ModalRoute.of(dialogContext);
                            return VideoPreviewDialog(
                              videoUrl: source.toString(),
                            );
                          },
                        );
                        _sensitiveDialog = false;
                        _sensitiveRoute = null;
                      },
                      icon: const Icon(Icons.play_circle_outline),
                      label: const Text('Play video'),
                    ),
                  if (content.isNotEmpty)
                    LibraryDocumentContent(
                      key: const ValueKey('library-reader-content'),
                      content: content,
                      format: format,
                    )
                  else if (format != 'video')
                    Text(
                      source != null
                          ? 'No text preview is available. Open the source to read this guide.'
                          : 'This guide does not have content yet.',
                      style: Theme.of(
                        context,
                      ).textTheme.bodyMedium?.copyWith(color: AppTheme.muted),
                    ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _options() => PopupMenuButton<String>(
    key: const ValueKey('library-reader-more'),
    tooltip: 'Guide options',
    enabled: !_busy,
    iconSize: 20,
    onSelected: (action) async {
      switch (action) {
        case 'edit':
          await _edit();
        case 'share':
          await _share();
        case 'delete':
          await _delete();
        case 'refresh':
          await _load();
      }
    },
    itemBuilder: (_) => [
      if (_owner) ...[
        _menuItem('edit', 'Edit', Icons.edit_outlined),
        _menuItem('share', 'Share', Icons.people_outline),
        _menuItem('delete', 'Delete', Icons.delete_outline),
        const PopupMenuDivider(),
      ],
      _menuItem('refresh', 'Refresh guide', Icons.refresh),
    ],
  );

  PopupMenuItem<String> _menuItem(String action, String label, IconData icon) =>
      PopupMenuItem(
        key: ValueKey('library-reader-$action'),
        value: action,
        child: Row(
          children: [
            Icon(icon, size: 18),
            const SizedBox(width: 12),
            Expanded(child: Text(label)),
          ],
        ),
      );
}
