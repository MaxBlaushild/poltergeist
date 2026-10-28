import 'dart:async';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:travel_angels/utils/trip_document_download.dart';
import 'package:travel_angels/widgets/library_document_content.dart';
import 'package:travel_angels/widgets/travel_trip_design.dart';
import 'package:travel_angels/widgets/trip_document_access.dart';

class TravelTripDocumentScreen extends StatefulWidget {
  const TravelTripDocumentScreen({
    super.key,
    required this.access,
    required this.documentId,
  });
  final TripDocumentAccess access;
  final String documentId;

  @override
  State<TravelTripDocumentScreen> createState() =>
      _TravelTripDocumentScreenState();
}

class _TravelTripDocumentScreenState extends State<TravelTripDocumentScreen>
    with WidgetsBindingObserver {
  Map<String, dynamic>? _document;
  List<Map<String, dynamic>> _comments = [];
  final _composer = TextEditingController();
  final _composerFocus = FocusNode();
  final _composerKey = GlobalKey();
  String? _editingId;
  String? _error;
  String? _commentError;
  String? _nextBefore;
  bool _hasMore = false;
  bool _loading = true;
  bool _busy = false;
  bool _downloading = false;
  bool _loadingOlder = false;
  bool _canContribute = false;
  int _request = 0;
  Uint8List? _image;
  Timer? _timer;

  String get _path => widget.access.documentPath(widget.documentId);
  bool get _canComment =>
      widget.access.active && widget.access.canContribute && _canContribute;

  @override
  void initState() {
    super.initState();
    widget.access.addListener(_accessChanged);
    WidgetsBinding.instance.addObserver(this);
    _timer = Timer.periodic(
      const Duration(seconds: 30),
      (_) => _load(quiet: true),
    );
    _load();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) _load(quiet: true);
  }

  void _accessChanged() {
    _request++;
    _composer.clear();
    _editingId = null;
    _image = null;
    if (!widget.access.active) {
      if (mounted) {
        setState(() {
          _document = null;
          _comments = [];
          _loading = false;
        });
      }
    } else {
      if (mounted) setState(() => _canContribute = false);
      _load(quiet: true);
    }
  }

  @override
  void dispose() {
    _request++;
    _timer?.cancel();
    widget.access.removeListener(_accessChanged);
    WidgetsBinding.instance.removeObserver(this);
    _composer.dispose();
    _composerFocus.dispose();
    _image = null;
    super.dispose();
  }

  bool _current(int generation, [int? request]) =>
      mounted &&
      widget.access.current(generation) &&
      (request == null || request == _request);

  void _failAccess() {
    widget.access.invalidate();
    if (mounted) {
      setState(() {
        _document = null;
        _comments = [];
        _image = null;
        _loading = false;
        _error = 'This document is no longer available to your account.';
      });
    }
  }

  Future<void> _load({bool quiet = false}) async {
    if (!widget.access.active) return;
    final generation = widget.access.generation;
    final request = ++_request;
    if (!quiet) {
      setState(() {
        _loading = true;
        _error = null;
      });
    }
    try {
      final responses = await Future.wait([
        widget.access.service.get(_path, query: widget.access.query),
        widget.access.service.get(
          '$_path/comments',
          query: widget.access.query,
        ),
      ]);
      if (!_current(generation, request)) return;
      final document = tripMap(responses[0]['document']);
      final canContribute =
          document['canContribute'] == true &&
          responses[1]['canContribute'] == true;
      setState(() {
        _document = document;
        _comments = tripMaps(responses[1]['comments']);
        _nextBefore = responses[1]['nextBefore'] as String?;
        _hasMore = responses[1]['hasMore'] == true;
        _canContribute = canContribute;
        _loading = false;
        _error = null;
        if (!canContribute) {
          _composer.clear();
          _editingId = null;
        }
      });
      if (_image == null &&
          ['image/jpeg', 'image/png'].contains(document['contentType']) &&
          document['fileUrl'] is String) {
        try {
          final image = await widget.access.service.loadTripDocumentFile(
            document['fileUrl'] as String,
            query: widget.access.query,
          );
          if (_current(generation, request)) setState(() => _image = image);
        } catch (error) {
          if (_current(generation, request) && tripDocumentDenied(error)) {
            _failAccess();
          }
        }
      }
    } catch (error) {
      if (!_current(generation, request)) return;
      if (tripDocumentDenied(error)) {
        _failAccess();
        return;
      }
      // A failed recheck does not leave a cached protected reader on screen.
      setState(() {
        _document = null;
        _comments = [];
        _image = null;
        _loading = false;
        _error = 'The document could not be loaded. Please try again.';
      });
    }
  }

  Future<void> _olderComments() async {
    if (_loadingOlder || !_hasMore || _nextBefore == null) return;
    final generation = widget.access.generation;
    final request = _request;
    setState(() {
      _loadingOlder = true;
      _commentError = null;
    });
    try {
      final response = await widget.access.service.get(
        '$_path/comments',
        query: {...widget.access.query, 'before': _nextBefore},
      );
      if (!_current(generation, request)) return;
      setState(() {
        final known = _comments.map((comment) => comment['id']).toSet();
        _comments = [
          ...tripMaps(
            response['comments'],
          ).where((comment) => !known.contains(comment['id'])),
          ..._comments,
        ];
        _hasMore = response['hasMore'] == true;
        _nextBefore = response['nextBefore'] as String?;
      });
    } catch (error) {
      if (!_current(generation, request)) return;
      if (tripDocumentDenied(error)) {
        _failAccess();
        return;
      }
      setState(() => _commentError = 'Older comments could not be loaded.');
    } finally {
      if (_current(generation)) setState(() => _loadingOlder = false);
    }
  }

  Future<void> _download() async {
    final document = _document;
    if (_downloading || document == null || document['fileUrl'] is! String) {
      return;
    }
    final generation = widget.access.generation;
    setState(() {
      _downloading = true;
      _error = null;
    });
    try {
      final bytes = await widget.access.service.loadTripDocumentFile(
        document['fileUrl'] as String,
        query: widget.access.query,
      );
      if (!_current(generation)) return;
      await saveTripDocumentBytes(
        bytes,
        document['filename'] as String? ?? 'trip-document',
        isCurrent: () => _current(generation),
      );
    } catch (error) {
      if (!_current(generation)) return;
      if (tripDocumentDenied(error)) {
        _failAccess();
        return;
      }
      setState(() => _error = 'The file could not be saved. Please try again.');
    } finally {
      if (_current(generation)) setState(() => _downloading = false);
    }
  }

  Future<bool> _confirm(String title, String message) async =>
      await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: Text(title),
          content: Text(message),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Remove'),
            ),
          ],
        ),
      ) ==
      true;

  Future<void> _removeDocument() async {
    final generation = widget.access.generation;
    if (!await _confirm(
      'Remove document?',
      'This removes it and its comments from the trip. Any original library guide stays unchanged.',
    )) {
      return;
    }
    if (!_current(generation)) return;
    setState(() => _busy = true);
    try {
      await widget.access.service.delete(_path, query: widget.access.query);
      if (mounted && _current(generation)) Navigator.pop(context);
    } catch (error) {
      if (!_current(generation)) return;
      if (tripDocumentDenied(error)) {
        _failAccess();
        return;
      }
      setState(() {
        _busy = false;
        _error = tripError(error);
      });
    }
  }

  Future<void> _saveComment() async {
    final body = _composer.text.trim();
    if (!_canComment || _busy || body.isEmpty) return;
    final generation = widget.access.generation;
    final editing = _editingId;
    setState(() {
      _busy = true;
      _commentError = null;
    });
    try {
      if (editing == null) {
        await widget.access.service.post(
          '$_path/comments',
          query: widget.access.query,
          data: {'body': body},
        );
      } else {
        await widget.access.service.put(
          '$_path/comments/${Uri.encodeComponent(editing)}',
          query: widget.access.query,
          data: {'body': body},
        );
      }
      if (!_current(generation)) return;
      _composer.clear();
      setState(() {
        _editingId = null;
        _busy = false;
      });
      await _load(quiet: true);
    } catch (error) {
      if (!_current(generation)) return;
      if (tripDocumentDenied(error)) {
        _failAccess();
        return;
      }
      setState(() {
        _busy = false;
        _commentError = tripError(error);
      });
    }
  }

  Future<void> _removeComment(Map<String, dynamic> comment) async {
    final generation = widget.access.generation;
    if (!await _confirm(
      'Remove comment?',
      'This comment will be removed from the conversation.',
    )) {
      return;
    }
    if (!_current(generation)) return;
    setState(() {
      _busy = true;
      _commentError = null;
    });
    try {
      await widget.access.service.delete(
        '$_path/comments/${Uri.encodeComponent(comment['id'] as String)}',
        query: widget.access.query,
      );
      if (!_current(generation)) return;
      setState(() {
        _busy = false;
        if (_editingId == comment['id']) {
          _editingId = null;
          _composer.clear();
        }
      });
      await _load(quiet: true);
    } catch (error) {
      if (!_current(generation)) return;
      if (tripDocumentDenied(error)) {
        _failAccess();
        return;
      }
      setState(() {
        _busy = false;
        _commentError = tripError(error);
      });
    }
  }

  String _date(dynamic value) {
    final date = DateTime.tryParse(value as String? ?? '');
    return date == null ? '' : DateFormat.MMMd().format(date.toLocal());
  }

  @override
  Widget build(BuildContext context) {
    final document = widget.access.active ? _document : null;
    return Scaffold(
      backgroundColor: tripPaper,
      appBar: AppBar(
        title: const Text('Documents'),
        actions: [
          if (document?['canDelete'] == true)
            PopupMenuButton<String>(
              tooltip: 'Document options',
              enabled: !_busy,
              onSelected: (_) => _removeDocument(),
              itemBuilder: (_) => [
                const PopupMenuItem(
                  value: 'remove',
                  child: Text('Remove from trip'),
                ),
              ],
            ),
        ],
      ),
      body: _loading && document == null
          ? const Center(child: CircularProgressIndicator())
          : document == null
          ? Center(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(_error ?? 'This document is no longer available.'),
                    if (widget.access.active)
                      TextButton(onPressed: _load, child: const Text('Retry')),
                  ],
                ),
              ),
            )
          : SingleChildScrollView(
              child: Center(
                child: ConstrainedBox(
                  constraints: const BoxConstraints(maxWidth: 820),
                  child: Padding(
                    padding: const EdgeInsets.fromLTRB(20, 24, 20, 40),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Text(
                          tripDocumentKind(document).toUpperCase(),
                          style: const TextStyle(
                            fontSize: 11,
                            letterSpacing: 1.5,
                            color: tripRust,
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          document['title'] as String? ?? 'Untitled document',
                          style: Theme.of(context).textTheme.headlineMedium,
                        ),
                        const SizedBox(height: 8),
                        Text(
                          [
                            tripMap(document['addedBy'])['displayName']
                                    as String? ??
                                '',
                            _date(document['createdAt']),
                          ].where((item) => item.isNotEmpty).join(' · '),
                          style: const TextStyle(
                            color: tripMuted,
                            fontSize: 13,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Wrap(
                          spacing: 8,
                          runSpacing: 4,
                          children: [
                            if (document['fileUrl'] is String)
                              TextButton.icon(
                                onPressed: _downloading ? null : _download,
                                icon: const Icon(
                                  Icons.download_outlined,
                                  size: 18,
                                ),
                                label: Text(
                                  _downloading ? 'Saving…' : 'Download',
                                ),
                              ),
                            if (librarySafeUri(document['link'] as String?) !=
                                null)
                              TextButton.icon(
                                onPressed: () => openLibrarySource(
                                  context,
                                  document['link'] as String,
                                ),
                                icon: const Icon(Icons.open_in_new, size: 16),
                                label: const Text('Open source'),
                              ),
                          ],
                        ),
                        if (_error != null)
                          Padding(
                            padding: const EdgeInsets.only(top: 8),
                            child: Text(
                              _error!,
                              style: TextStyle(
                                color: Theme.of(context).colorScheme.error,
                              ),
                            ),
                          ),
                        const SizedBox(height: 16),
                        if (_image != null)
                          ClipRRect(
                            borderRadius: BorderRadius.circular(12),
                            child: Image.memory(
                              _image!,
                              fit: BoxFit.contain,
                              errorBuilder: (_, _, _) =>
                                  const Text('Download this image to open it.'),
                            ),
                          )
                        else if ((document['content'] as String? ?? '')
                            .trim()
                            .isNotEmpty)
                          DecoratedBox(
                            decoration: BoxDecoration(
                              color: tripCream,
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: Padding(
                              padding: const EdgeInsets.all(20),
                              child: SelectionArea(
                                child: LibraryDocumentContent(
                                  content: document['content'] as String,
                                  format:
                                      document['format'] as String? ?? 'guide',
                                ),
                              ),
                            ),
                          )
                        else
                          Container(
                            padding: const EdgeInsets.all(24),
                            decoration: BoxDecoration(
                              color: tripCream,
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: Column(
                              children: [
                                Icon(
                                  document['kind'] == 'link'
                                      ? Icons.link
                                      : Icons.insert_drive_file_outlined,
                                  color: tripRust,
                                  size: 32,
                                ),
                                const SizedBox(height: 12),
                                Text(
                                  document['kind'] == 'link'
                                      ? 'Open the source to view this document.'
                                      : 'Download to open the original file.',
                                  textAlign: TextAlign.center,
                                ),
                                if (document['filename'] is String) ...[
                                  const SizedBox(height: 6),
                                  Text(
                                    document['filename'] as String,
                                    textAlign: TextAlign.center,
                                    style: const TextStyle(
                                      color: tripMuted,
                                      fontSize: 12,
                                    ),
                                  ),
                                ],
                              ],
                            ),
                          ),
                        const SizedBox(height: 32),
                        const Divider(color: tripOutline),
                        const SizedBox(height: 12),
                        Text(
                          'Conversation',
                          style: Theme.of(context).textTheme.titleLarge,
                        ),
                        if (_hasMore)
                          Align(
                            alignment: Alignment.centerLeft,
                            child: TextButton(
                              onPressed: _loadingOlder ? null : _olderComments,
                              child: Text(
                                _loadingOlder ? 'Loading…' : 'Earlier comments',
                              ),
                            ),
                          ),
                        if (_comments.isEmpty)
                          const Padding(
                            padding: EdgeInsets.symmetric(vertical: 16),
                            child: Text(
                              'Questions, ideas, and little details for the group.',
                              style: TextStyle(color: tripMuted),
                            ),
                          ),
                        for (final comment in _comments) _comment(comment),
                        if (_commentError != null)
                          Padding(
                            padding: const EdgeInsets.symmetric(vertical: 8),
                            child: Text(
                              _commentError!,
                              style: TextStyle(
                                color: Theme.of(context).colorScheme.error,
                              ),
                            ),
                          ),
                        if (_canComment) ...[
                          const SizedBox(height: 12),
                          if (_editingId != null)
                            Row(
                              children: [
                                const Expanded(
                                  child: Text(
                                    'Editing comment',
                                    style: TextStyle(color: tripMuted),
                                  ),
                                ),
                                TextButton(
                                  onPressed: _busy
                                      ? null
                                      : () => setState(() {
                                          _editingId = null;
                                          _composer.clear();
                                        }),
                                  child: const Text('Cancel'),
                                ),
                              ],
                            ),
                          Row(
                            key: _composerKey,
                            crossAxisAlignment: CrossAxisAlignment.end,
                            children: [
                              Expanded(
                                child: TextField(
                                  controller: _composer,
                                  focusNode: _composerFocus,
                                  enabled: !_busy,
                                  minLines: 1,
                                  maxLines: 5,
                                  maxLength: 4000,
                                  textCapitalization:
                                      TextCapitalization.sentences,
                                  decoration: const InputDecoration(
                                    hintText: 'Add a comment…',
                                    counterText: '',
                                  ),
                                ),
                              ),
                              const SizedBox(width: 4),
                              IconButton(
                                tooltip: _editingId == null
                                    ? 'Post comment'
                                    : 'Save comment',
                                onPressed: _busy ? null : _saveComment,
                                icon: Icon(
                                  _editingId == null
                                      ? Icons.arrow_upward
                                      : Icons.check,
                                ),
                              ),
                            ],
                          ),
                        ],
                      ],
                    ),
                  ),
                ),
              ),
            ),
    );
  }

  Widget _comment(Map<String, dynamic> comment) => Padding(
    padding: const EdgeInsets.only(top: 16),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: Wrap(
                spacing: 8,
                runSpacing: 3,
                children: [
                  Text(
                    tripMap(comment['author'])['displayName'] as String? ??
                        'Traveler',
                    style: const TextStyle(fontWeight: FontWeight.w600),
                  ),
                  Text(
                    _date(comment['createdAt']),
                    style: const TextStyle(color: tripMuted, fontSize: 12),
                  ),
                ],
              ),
            ),
            if (comment['canDelete'] == true ||
                (comment['canEdit'] == true && _canComment))
              SizedBox(
                width: 36,
                height: 28,
                child: PopupMenuButton<String>(
                  tooltip: 'Comment options',
                  padding: EdgeInsets.zero,
                  enabled: !_busy,
                  iconSize: 18,
                  onSelected: (action) {
                    if (action == 'remove') {
                      _removeComment(comment);
                      return;
                    }
                    setState(() {
                      _editingId = comment['id'] as String;
                      _composer.text = comment['body'] as String? ?? '';
                    });
                    WidgetsBinding.instance.addPostFrameCallback((_) {
                      if (!mounted || !_canComment) return;
                      final composerContext = _composerKey.currentContext;
                      if (composerContext != null) {
                        Scrollable.ensureVisible(composerContext);
                      }
                      _composerFocus.requestFocus();
                    });
                  },
                  itemBuilder: (_) => [
                    if (comment['canEdit'] == true && _canComment)
                      const PopupMenuItem(
                        value: 'edit',
                        child: Text('Edit comment'),
                      ),
                    if (comment['canDelete'] == true)
                      const PopupMenuItem(
                        value: 'remove',
                        child: Text('Remove comment'),
                      ),
                  ],
                ),
              ),
          ],
        ),
        const SizedBox(height: 5),
        SelectableText(
          comment['body'] as String? ?? '',
          style: const TextStyle(height: 1.5),
        ),
      ],
    ),
  );
}
