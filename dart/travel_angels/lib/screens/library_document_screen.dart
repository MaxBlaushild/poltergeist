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
        title: const Text('Travel library'),
        actions: [
          IconButton(
            tooltip: 'Refresh guide',
            onPressed: _loading || _busy ? null : () => _load(),
            icon: const Icon(Icons.refresh),
          ),
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
                        MediaQuery.sizeOf(context).width < 600 ? 16 : 28,
                      ),
                      child: LayoutBuilder(
                        builder: (context, constraints) {
                          final wide =
                              constraints.maxWidth >= 850 &&
                              MediaQuery.textScalerOf(context).scale(14) <= 21;
                          final destination = libraryDestination(document);
                          return Column(
                            crossAxisAlignment: CrossAxisAlignment.stretch,
                            children: [
                              if (destination.isNotEmpty)
                                Text(
                                  destination,
                                  style: Theme.of(context).textTheme.labelLarge
                                      ?.copyWith(
                                        color: AppTheme.rust,
                                        letterSpacing: 1,
                                      ),
                                ),
                              const SizedBox(height: 8),
                              Text(
                                document.title,
                                style: Theme.of(
                                  context,
                                ).textTheme.headlineLarge,
                              ),
                              const SizedBox(height: 12),
                              Wrap(
                                spacing: 18,
                                runSpacing: 8,
                                crossAxisAlignment: WrapCrossAlignment.center,
                                children: [
                                  Text(
                                    libraryFormatLabel(libraryFormat(document)),
                                    style: Theme.of(
                                      context,
                                    ).textTheme.bodySmall,
                                  ),
                                  if (document.updatedAt != null)
                                    Text(
                                      'Updated ${DateFormat.yMMMd().format(document.updatedAt!.toLocal())}',
                                      style: Theme.of(context)
                                          .textTheme
                                          .bodySmall
                                          ?.copyWith(color: AppTheme.muted),
                                    ),
                                  if (_owner) ...[
                                    TextButton.icon(
                                      key: const ValueKey(
                                        'library-reader-edit',
                                      ),
                                      onPressed: _busy ? null : _edit,
                                      icon: const Icon(
                                        Icons.edit_outlined,
                                        size: 18,
                                      ),
                                      label: const Text('Edit'),
                                    ),
                                    TextButton.icon(
                                      key: const ValueKey(
                                        'library-reader-share',
                                      ),
                                      onPressed: _busy ? null : _share,
                                      icon: const Icon(
                                        Icons.people_outline,
                                        size: 18,
                                      ),
                                      label: const Text('Share'),
                                    ),
                                    TextButton.icon(
                                      key: const ValueKey(
                                        'library-reader-delete',
                                      ),
                                      onPressed: _busy ? null : _delete,
                                      icon: const Icon(
                                        Icons.delete_outline,
                                        size: 18,
                                      ),
                                      label: const Text('Delete'),
                                    ),
                                  ],
                                ],
                              ),
                              const SizedBox(height: 24),
                              if (wide)
                                Row(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Expanded(flex: 7, child: _reader(document)),
                                    const SizedBox(width: 28),
                                    Expanded(
                                      flex: 3,
                                      child: _sidebar(document),
                                    ),
                                  ],
                                )
                              else
                                Column(
                                  crossAxisAlignment:
                                      CrossAxisAlignment.stretch,
                                  children: [
                                    _sidebar(document),
                                    const SizedBox(height: 24),
                                    _reader(document),
                                  ],
                                ),
                              const SizedBox(height: 32),
                            ],
                          );
                        },
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
            padding: const EdgeInsets.all(14),
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
            padding: const EdgeInsets.all(20),
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
          Container(
            decoration: const BoxDecoration(
              border: Border(top: BorderSide(color: AppTheme.outline)),
            ),
            padding: const EdgeInsets.all(14),
            child: Text(
              linked
                  ? 'This guide links to the original file. Its owner controls access there.'
                  : imported
                  ? 'Imported copy. Changes to the original file do not sync automatically. Its access permissions are kept separately.'
                  : 'A guide saved in Travel Angels.',
              style: Theme.of(context).textTheme.labelSmall,
            ),
          ),
        ],
      ),
    );
  }

  Widget _sidebar(Document document) {
    final author = libraryAuthor(document);
    final source = librarySafeUri(document.link);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          children: [
            CircleAvatar(
              backgroundColor: AppTheme.peach,
              foregroundColor: AppTheme.ink,
              child: Text(author.characters.first.toUpperCase()),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                '$author${_owner ? ' · owner' : ''}',
                style: Theme.of(context).textTheme.bodyMedium,
              ),
            ),
          ],
        ),
        if (document.personalNote.trim().isNotEmpty) ...[
          const SizedBox(height: 22),
          Text(
            _owner ? 'Your note' : 'A note from $author',
            style: Theme.of(
              context,
            ).textTheme.labelLarge?.copyWith(color: AppTheme.rust),
          ),
          const SizedBox(height: 8),
          Text(
            document.personalNote,
            style: Theme.of(context).textTheme.bodyLarge?.copyWith(height: 1.7),
          ),
        ],
        const SizedBox(height: 22),
        FilledButton.icon(
          key: const ValueKey('library-reader-save'),
          onPressed: _busy ? null : _save,
          icon: Icon(document.isSaved ? Icons.bookmark : Icons.bookmark_border),
          label: Text(document.isSaved ? 'Saved' : 'Save guide'),
        ),
        const SizedBox(height: 10),
        OutlinedButton.icon(
          key: const ValueKey('library-reader-trip'),
          onPressed: _busy ? null : _trips,
          icon: const Icon(Icons.calendar_today_outlined),
          label: const Text('Save to a trip'),
        ),
        if (source != null) ...[
          const SizedBox(height: 10),
          TextButton.icon(
            key: const ValueKey('library-reader-source'),
            onPressed: () => openLibrarySource(context, source.toString()),
            icon: const Icon(Icons.open_in_new, size: 18),
            label: const Text('Open source'),
          ),
        ],
        const SizedBox(height: 22),
        Container(
          padding: const EdgeInsets.only(top: 18),
          decoration: const BoxDecoration(
            border: Border(top: BorderSide(color: AppTheme.outline)),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(
                    _owner && document.visibility == 'private'
                        ? Icons.lock_outline
                        : Icons.people_outline,
                    size: 18,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      _owner
                          ? libraryAudience(document.visibility)
                          : 'Shared with you',
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              Text(
                _owner
                    ? 'You choose who can read this guide.'
                    : 'Saving keeps the author and original sharing permissions. Access can change if its owner stops sharing.',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  height: 1.6,
                  color: AppTheme.muted,
                ),
              ),
              if (_tripNames.isNotEmpty)
                Padding(
                  padding: const EdgeInsets.only(top: 14),
                  child: Text(
                    'Your trip references: ${_tripNames.join(' · ')}',
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                ),
            ],
          ),
        ),
      ],
    );
  }
}
