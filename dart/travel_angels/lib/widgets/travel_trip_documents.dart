import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/screens/travel_trip_document_screen.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/document_service.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/library_document_content.dart';
import 'package:travel_angels/widgets/travel_trip_design.dart';
import 'package:travel_angels/widgets/trip_document_access.dart';

class TravelTripDocuments extends StatefulWidget {
  const TravelTripDocuments({
    super.key,
    required this.stopId,
    required this.userId,
    required this.service,
    required this.canView,
    required this.canContribute,
    this.accessQuery = const {},
    this.refreshToken = 0,
    this.documentService,
    this.onFlowChanged,
    this.onJoin,
  });

  final String stopId;
  final String? userId;
  final TravelCalendarService service;
  final bool canView;
  final bool canContribute;
  final Map<String, dynamic> accessQuery;
  final int refreshToken;
  final DocumentService? documentService;
  final ValueChanged<bool>? onFlowChanged;
  final Future<void> Function()? onJoin;

  @override
  State<TravelTripDocuments> createState() => _TravelTripDocumentsState();
}

class _TravelTripDocumentsState extends State<TravelTripDocuments> {
  late TripDocumentAccess _access;
  AuthProvider? _auth;
  List<Map<String, dynamic>> _documents = [];
  String? _error;
  bool _loading = false;
  int _request = 0;
  Route<dynamic>? _pageRoute;
  bool _flowOpen = false;

  @override
  void initState() {
    super.initState();
    _newAccess();
    if (widget.canView && widget.userId != null) _load();
  }

  void _newAccess() {
    _access = TripDocumentAccess(
      service: widget.service,
      stopId: widget.stopId,
      userId: widget.userId ?? '',
      query: Map.unmodifiable(widget.accessQuery),
      canContribute: widget.canContribute,
    );
    _access.addListener(_accessChanged);
  }

  void _accessChanged() {
    if (_access.active || !mounted) return;
    _request++;
    setState(() {
      _documents = [];
      _loading = false;
      _error = 'These trip documents are no longer available to your account.';
    });
    _closeFlows();
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final auth = context.read<AuthProvider?>();
    if (auth != _auth) {
      _auth?.removeListener(_identityChanged);
      _auth = auth;
      auth?.addListener(_identityChanged);
    }
  }

  void _identityChanged() {
    if (_auth?.user?.id == _access.userId) return;
    _access.invalidate();
    _request++;
    if (mounted) setState(() => _documents = []);
    _closeFlows();
  }

  void _closeFlows() {
    final route = _pageRoute;
    if (!_flowOpen || route == null) return;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted && _flowOpen) {
        Navigator.of(context).popUntil((candidate) => candidate == route);
      }
    });
  }

  @override
  void didUpdateWidget(covariant TravelTripDocuments oldWidget) {
    super.didUpdateWidget(oldWidget);
    final identityChanged =
        oldWidget.userId != widget.userId ||
        oldWidget.stopId != widget.stopId ||
        oldWidget.accessQuery.toString() != widget.accessQuery.toString();
    if (identityChanged || (oldWidget.canView && !widget.canView)) {
      _access.invalidate();
      _request++;
      _documents = [];
      _closeFlows();
      _access.removeListener(_accessChanged);
      _newAccess();
    }
    _access.updateContribution(widget.canContribute);
    if (widget.canView &&
        widget.userId != null &&
        (identityChanged ||
            !oldWidget.canView ||
            widget.refreshToken != oldWidget.refreshToken)) {
      _load();
    }
  }

  @override
  void dispose() {
    _request++;
    _access.active = false;
    _access.generation++;
    _access.removeListener(_accessChanged);
    _auth?.removeListener(_identityChanged);
    super.dispose();
  }

  Future<void> _load() async {
    if (!widget.canView ||
        widget.userId == null ||
        (_auth != null && _auth!.user?.id != widget.userId)) {
      return;
    }
    final request = ++_request;
    final access = _access;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final result = await widget.service.get(access.path, query: access.query);
      if (!mounted || request != _request || access != _access) return;
      if (result['canView'] == false) {
        access.invalidate();
        _closeFlows();
        setState(() {
          _documents = [];
          _loading = false;
          _error =
              'Trip documents are available to hosts and joined travelers.';
        });
        return;
      }
      if (!access.active) {
        access.removeListener(_accessChanged);
        _newAccess();
      }
      _access.updateContribution(result['canContribute'] == true);
      setState(() {
        _documents = tripMaps(result['documents']);
        _loading = false;
      });
    } catch (error) {
      if (!mounted || request != _request || access != _access) return;
      if (tripDocumentDenied(error)) {
        access.invalidate();
        _closeFlows();
      }
      setState(() {
        _documents = [];
        _loading = false;
        _error = tripDocumentDenied(error)
            ? 'These trip documents are no longer available to your account.'
            : 'The trip documents could not be loaded.';
      });
    }
  }

  Future<void> _flow(Future<void> Function() action) async {
    if (_flowOpen || !_access.active) return;
    _pageRoute = ModalRoute.of(context);
    _flowOpen = true;
    widget.onFlowChanged?.call(true);
    try {
      await action();
    } finally {
      _flowOpen = false;
      if (mounted) {
        widget.onFlowChanged?.call(false);
        if (_access.active) await _load();
      }
    }
  }

  Future<void> _read(Map<String, dynamic> document) => _flow(() async {
    await Navigator.of(context).push<void>(
      MaterialPageRoute(
        builder: (_) => TravelTripDocumentScreen(
          access: _access,
          documentId: document['id'] as String,
        ),
      ),
    );
  });

  Future<void> _add() => _flow(() async {
    final access = _access;
    final generation = access.generation;
    final kind = await showModalBottomSheet<String>(
      context: context,
      showDragHandle: true,
      isScrollControlled: true,
      builder: (context) => SafeArea(
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              ListTile(
                leading: const Icon(Icons.upload_file_outlined),
                title: const Text('Upload a file'),
                subtitle: const Text(
                  'Documents, spreadsheets, or images · up to 20 MB',
                ),
                onTap: () => Navigator.pop(context, 'file'),
              ),
              ListTile(
                leading: const Icon(Icons.link_outlined),
                title: const Text('Add a link'),
                onTap: () => Navigator.pop(context, 'link'),
              ),
              ListTile(
                leading: const Icon(Icons.auto_stories_outlined),
                title: const Text('Copy a guide from your library'),
                onTap: () => Navigator.pop(context, 'guide'),
              ),
              const SizedBox(height: 16),
            ],
          ),
        ),
      ),
    );
    if (!mounted ||
        kind == null ||
        !access.current(generation) ||
        !access.canContribute) {
      return;
    }
    await showDialog<bool>(
      context: context,
      builder: (_) => TripDocumentAddDialog(
        access: access,
        kind: kind,
        documentService: widget.documentService,
      ),
    );
  });

  @override
  Widget build(BuildContext context) {
    final available = widget.canView && widget.userId != null;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: Text(
                'Documents',
                style: Theme.of(context).textTheme.headlineSmall,
              ),
            ),
            if (available && _access.active && _access.canContribute)
              IconButton(
                tooltip: 'Add document',
                onPressed: _flowOpen ? null : _add,
                icon: const Icon(Icons.add, size: 22),
              ),
          ],
        ),
        const SizedBox(height: 12),
        if (!available)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Shared documents are for hosts and joined travelers.',
                  style: TextStyle(color: tripMuted),
                ),
                if (widget.onJoin != null)
                  TextButton(
                    onPressed: widget.onJoin,
                    child: const Text('Join the trip'),
                  ),
              ],
            ),
          )
        else ...[
          if (_loading && _documents.isEmpty)
            const LinearProgressIndicator(minHeight: 2),
          if (_error != null)
            Row(
              children: [
                Expanded(child: Text(_error!)),
                TextButton(
                  onPressed: _loading ? null : _load,
                  child: const Text('Retry'),
                ),
              ],
            ),
          if (!_loading && _error == null && _documents.isEmpty)
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 20),
              child: Text(
                'A place for the itinerary, reservations, and everything you’ll need together.',
                style: TextStyle(color: tripMuted, height: 1.5),
              ),
            ),
          for (final document in _documents)
            Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Material(
                color: tripCream,
                borderRadius: BorderRadius.circular(12),
                child: InkWell(
                  borderRadius: BorderRadius.circular(12),
                  onTap: () => _read(document),
                  child: Padding(
                    padding: const EdgeInsets.all(14),
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Icon(
                          document['kind'] == 'link'
                              ? Icons.link
                              : document['format'] == 'spreadsheet'
                              ? Icons.table_chart_outlined
                              : Icons.description_outlined,
                          color: tripRust,
                          size: 22,
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                document['title'] as String? ??
                                    'Untitled document',
                                style: const TextStyle(
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                              const SizedBox(height: 4),
                              Text(
                                [
                                  tripDocumentKind(document),
                                  tripMap(document['addedBy'])['displayName']
                                          as String? ??
                                      '',
                                  if ((document['commentCount'] as num? ?? 0) >
                                      0)
                                    '${document['commentCount']} ${document['commentCount'] == 1 ? 'comment' : 'comments'}',
                                ].where((s) => s.isNotEmpty).join(' · '),
                                style: Theme.of(context).textTheme.bodySmall
                                    ?.copyWith(color: tripMuted),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(width: 4),
                        const Icon(
                          Icons.chevron_right,
                          size: 18,
                          color: tripMuted,
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
        ],
      ],
    );
  }
}

class TripDocumentAddDialog extends StatefulWidget {
  const TripDocumentAddDialog({
    super.key,
    required this.access,
    required this.kind,
    this.documentService,
  });
  final TripDocumentAccess access;
  final String kind;
  final DocumentService? documentService;

  @override
  State<TripDocumentAddDialog> createState() => _TripDocumentAddDialogState();
}

class _TripDocumentAddDialogState extends State<TripDocumentAddDialog> {
  final _title = TextEditingController();
  final _link = TextEditingController();
  List<Map<String, dynamic>> _guides = [];
  String? _guideId;
  Uint8List? _bytes;
  String? _filename;
  String? _error;
  bool _busy = false;
  bool _loading = false;
  late final int _generation = widget.access.generation;
  bool get _current =>
      mounted &&
      widget.access.current(_generation) &&
      widget.access.canContribute;

  @override
  void initState() {
    super.initState();
    widget.access.addListener(_accessChanged);
    if (widget.kind == 'guide') _loadGuides();
  }

  void _accessChanged() {
    if (_current) return;
    _bytes = null;
    _guides = [];
    _title.clear();
    _link.clear();
    if (mounted) setState(() {});
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted && ModalRoute.of(context)?.isCurrent == true) {
        Navigator.pop(context);
      }
    });
  }

  @override
  void dispose() {
    widget.access.removeListener(_accessChanged);
    _title.dispose();
    _link.dispose();
    _bytes = null;
    super.dispose();
  }

  Future<void> _loadGuides() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final service =
          widget.documentService ??
          DocumentService(APIClient(ApiConstants.baseUrl));
      final guides = await service.getLibraryDocuments('yours');
      if (!_current) return;
      setState(() {
        _guides = guides
            .where((guide) => guide['userId'] == widget.access.userId)
            .toList();
        _loading = false;
      });
    } catch (_) {
      if (!_current) return;
      setState(() {
        _loading = false;
        _error = 'Your guides could not be loaded.';
      });
    }
  }

  Future<void> _pickFile() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: const [
          'pdf',
          'docx',
          'xlsx',
          'csv',
          'txt',
          'jpeg',
          'jpg',
          'png',
        ],
        withData: true,
      );
      if (!_current || result == null) return;
      final file = result.files.single;
      if (file.bytes == null ||
          file.bytes!.isEmpty ||
          file.size > 20 * 1024 * 1024) {
        setState(() => _error = 'Choose a nonempty file up to 20 MB.');
        return;
      }
      setState(() {
        _bytes = file.bytes;
        _filename = file.name;
        if (_title.text.trim().isEmpty) _title.text = file.name;
      });
    } catch (_) {
      if (_current) {
        setState(
          () => _error = 'The file could not be opened. Please try again.',
        );
      }
    } finally {
      if (_current) setState(() => _busy = false);
    }
  }

  Future<void> _save() async {
    if (!_current || _busy) return;
    final title = _title.text.trim();
    if (title.isEmpty ||
        (widget.kind == 'file' && _bytes == null) ||
        (widget.kind == 'guide' && _guideId == null)) {
      setState(() => _error = 'Add a title and choose what to share.');
      return;
    }
    if (widget.kind == 'link' && librarySafeUri(_link.text.trim()) == null) {
      setState(() => _error = 'Enter a complete http or https link.');
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      if (widget.kind == 'file') {
        await widget.access.service.uploadTripDocument(
          widget.access.stopId,
          _bytes!,
          _filename!,
          title: title,
          query: widget.access.query,
        );
      } else {
        await widget.access.service.post(
          widget.access.path,
          query: widget.access.query,
          data: {
            'title': title,
            if (widget.kind == 'link') 'link': _link.text.trim(),
            if (widget.kind == 'guide') 'sourceDocumentId': _guideId,
          },
        );
      }
      if (mounted && _current) Navigator.pop(context, true);
    } catch (error) {
      if (!_current) return;
      if (tripDocumentDenied(error)) {
        widget.access.invalidate();
        return;
      }
      setState(() {
        _busy = false;
        _error = tripError(error);
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    if (!_current) return const SizedBox.shrink();
    return AlertDialog(
      title: Text(
        widget.kind == 'guide'
            ? 'Share a guide'
            : widget.kind == 'link'
            ? 'Add a link'
            : 'Upload a file',
      ),
      content: SizedBox(
        width: 460,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                widget.kind == 'guide'
                    ? 'A copy is shared with this trip. Your library guide stays unchanged.'
                    : 'Shared with hosts and joined travelers.',
                style: Theme.of(
                  context,
                ).textTheme.bodySmall?.copyWith(color: tripMuted),
              ),
              const SizedBox(height: 16),
              if (widget.kind == 'file') ...[
                OutlinedButton.icon(
                  onPressed: _busy ? null : _pickFile,
                  icon: const Icon(Icons.attach_file, size: 18),
                  label: Text(
                    _filename ?? 'Choose file',
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                const Padding(
                  padding: EdgeInsets.only(top: 6, bottom: 16),
                  child: Text(
                    'PDF, DOCX, XLSX, CSV, TXT, JPEG, PNG · 20 MB max',
                    style: TextStyle(fontSize: 12, color: tripMuted),
                  ),
                ),
              ],
              if (widget.kind == 'guide') ...[
                if (_loading) const LinearProgressIndicator(minHeight: 2),
                if (!_loading && _guides.isEmpty && _error == null)
                  const Padding(
                    padding: EdgeInsets.symmetric(vertical: 12),
                    child: Text('Your library has no guides to copy yet.'),
                  ),
                for (final guide in _guides)
                  ListTile(
                    contentPadding: EdgeInsets.zero,
                    dense: true,
                    title: Text(guide['title'] as String? ?? 'Untitled guide'),
                    trailing: Icon(
                      _guideId == guide['id']
                          ? Icons.check_circle
                          : Icons.circle_outlined,
                      size: 20,
                    ),
                    onTap: _busy
                        ? null
                        : () => setState(() {
                            _guideId = guide['id'] as String;
                            _title.text = guide['title'] as String? ?? '';
                          }),
                  ),
                const SizedBox(height: 12),
              ],
              TextField(
                controller: _title,
                enabled: !_busy,
                maxLength: 200,
                textCapitalization: TextCapitalization.sentences,
                decoration: const InputDecoration(
                  labelText: 'Title',
                  counterText: '',
                ),
              ),
              if (widget.kind == 'link') ...[
                const SizedBox(height: 12),
                TextField(
                  controller: _link,
                  enabled: !_busy,
                  keyboardType: TextInputType.url,
                  autocorrect: false,
                  decoration: const InputDecoration(
                    labelText: 'Link',
                    hintText: 'https://',
                  ),
                ),
              ],
              if (_error != null)
                Padding(
                  padding: const EdgeInsets.only(top: 12),
                  child: Text(
                    _error!,
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                ),
              if (_error != null && widget.kind == 'guide' && _guides.isEmpty)
                TextButton(
                  onPressed: _loading ? null : _loadGuides,
                  child: const Text('Retry'),
                ),
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: _busy ? null : () => Navigator.pop(context),
          child: const Text('Cancel'),
        ),
        FilledButton(
          onPressed: _busy || _loading ? null : _save,
          child: Text(_busy ? 'Sharing…' : 'Share with trip'),
        ),
      ],
    );
  }
}
