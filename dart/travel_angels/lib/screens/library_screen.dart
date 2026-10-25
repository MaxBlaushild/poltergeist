import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/screens/documents_screen.dart';
import 'package:travel_angels/screens/library_document_screen.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/document_service.dart';
import 'package:travel_angels/services/friend_service.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/add_guide_flow.dart';
import 'package:travel_angels/widgets/library_document_card.dart';
import 'package:travel_angels/widgets/library_document_content.dart';
import 'package:travel_angels/widgets/library_document_dialogs.dart';

class LibraryScreen extends StatefulWidget {
  const LibraryScreen({
    super.key,
    this.service,
    this.friendService,
    this.calendarService,
    this.userId,
    this.onAddGuide,
    this.onEditDocument,
    this.initialScope = 'network',
  });
  final String initialScope;
  final DocumentService? service;
  final FriendService? friendService;
  final TravelCalendarService? calendarService;
  final String? userId;
  final Future<void> Function(BuildContext)? onAddGuide;
  final Future<bool?> Function(BuildContext, Document)? onEditDocument;
  @override
  State<LibraryScreen> createState() => _LibraryScreenState();
}

class _LibraryScreenState extends State<LibraryScreen> {
  late final DocumentService _service =
      widget.service ?? DocumentService(APIClient(ApiConstants.baseUrl));
  final _search = TextEditingController();
  final _saving = <String>{};
  AuthProvider? _auth;
  String? _identity;
  List<Document> _documents = [];
  late String _scope =
      const ['network', 'saved', 'yours'].contains(widget.initialScope)
      ? widget.initialScope
      : 'network';
  String _format = 'all';
  String? _error;
  bool _loading = true;
  bool _adding = false;
  int _generation = 0;
  @override
  void initState() {
    super.initState();
    _load();
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
  void didUpdateWidget(covariant LibraryScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.userId != widget.userId) _authChanged();
  }

  void _authChanged() {
    final id = _auth != null ? _auth!.user?.id : widget.userId;
    if (id == _identity || !mounted) return;
    _identity = id;
    _generation++;
    setState(() {
      _documents = [];
      _saving.clear();
      _loading = false;
      _error = id == null ? 'Sign in to see your travel library.' : null;
    });
    if (id != null) _load();
  }

  @override
  void dispose() {
    _auth?.removeListener(_authChanged);
    _search.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    final generation = ++_generation;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final data = await _service.getLibraryDocuments(_scope);
      if (!mounted || generation != _generation) return;
      setState(() => _documents = data.map(Document.fromJson).toList());
    } catch (error) {
      if (!mounted || generation != _generation) return;
      setState(() {
        _documents = [];
        _error = libraryError(
          error,
          'Your library could not be loaded. Please try again.',
        );
      });
    } finally {
      if (mounted && generation == _generation) {
        setState(() => _loading = false);
      }
    }
  }

  Future<void> _save(Document document) async {
    final generation = _generation;
    setState(() => _saving.add(document.id));
    try {
      final data = await _service.setDocumentSaved(
        document.id,
        !document.isSaved,
      );
      if (!mounted || generation != _generation) return;
      final updated = Document.fromJson(data);
      setState(() {
        _documents = _documents
            .map((item) => item.id == updated.id ? updated : item)
            .where((item) => _scope != 'saved' || item.isSaved)
            .toList();
      });
    } catch (error) {
      if (!mounted || generation != _generation) return;
      if (libraryAccessDenied(error)) {
        setState(
          () => _documents.removeWhere((item) => item.id == document.id),
        );
      }
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            libraryError(
              error,
              'The guide could not be saved. Please try again.',
            ),
          ),
        ),
      );
    } finally {
      if (mounted) setState(() => _saving.remove(document.id));
    }
  }

  Future<void> _open(Document document) async {
    await Navigator.push<bool>(
      context,
      MaterialPageRoute(
        builder: (_) => LibraryDocumentScreen(
          document: document,
          service: _service,
          friendService: widget.friendService,
          calendarService: widget.calendarService,
          userId: widget.userId,
          onEditDocument: widget.onEditDocument,
        ),
      ),
    );
    if (mounted) await _load();
  }

  Future<void> _add() async {
    setState(() => _adding = true);
    try {
      if (widget.onAddGuide != null) {
        await widget.onAddGuide!(context);
      } else {
        final document = await showAddGuideFlow(context);
        if (document != null && mounted) {
          setState(() {
            _scope = 'yours';
            _search.clear();
            _format = 'all';
            _documents = [];
          });
          await _open(document);
        }
      }
      if (mounted) await _load();
    } finally {
      if (mounted) setState(() => _adding = false);
    }
  }

  List<Document> get _filtered {
    final query = _search.text.trim().toLowerCase();
    return _documents
        .where(
          (document) =>
              (_format == 'all' || libraryFormat(document) == _format) &&
              (query.isEmpty ||
                  '${document.title} ${libraryDestination(document)} ${libraryAuthor(document)}'
                      .toLowerCase()
                      .contains(query)),
        )
        .toList();
  }

  void _changeScope(String scope) {
    if (_scope == scope) return;
    setState(() {
      _scope = scope;
      _documents = [];
    });
    _load();
  }

  @override
  Widget build(BuildContext context) => SafeArea(
    child: RefreshIndicator(
      onRefresh: _load,
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
                  final documents = _filtered;
                  final columns = constraints.maxWidth >= 1020
                      ? 3
                      : constraints.maxWidth >= 680
                      ? 2
                      : 1;
                  final cardWidth =
                      (constraints.maxWidth - 20 * (columns - 1)) / columns;
                  final heading = Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Travel library',
                        style: Theme.of(context).textTheme.headlineLarge,
                      ),
                      const SizedBox(height: 6),
                      Text(
                        'Good plans, passed from friend to friend.',
                        style: Theme.of(
                          context,
                        ).textTheme.bodyMedium?.copyWith(color: AppTheme.muted),
                      ),
                    ],
                  );
                  final add = FilledButton.icon(
                    key: const ValueKey('library-add-guide'),
                    onPressed: _adding ? null : _add,
                    icon: const Icon(Icons.add, size: 18),
                    label: const Text('Add a guide'),
                  );
                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      if (constraints.maxWidth < 660 ||
                          MediaQuery.textScalerOf(context).scale(14) > 21)
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [heading, const SizedBox(height: 16), add],
                        )
                      else
                        Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Expanded(child: heading),
                            const SizedBox(width: 20),
                            add,
                          ],
                        ),
                      const SizedBox(height: 24),
                      Container(
                        decoration: const BoxDecoration(
                          border: Border(
                            bottom: BorderSide(color: AppTheme.outline),
                          ),
                        ),
                        child: SingleChildScrollView(
                          scrollDirection: Axis.horizontal,
                          child: Row(
                            children: [
                              for (final tab in [
                                ('network', 'From your network'),
                                ('saved', 'Saved'),
                                ('yours', 'Yours'),
                              ])
                                Semantics(
                                  selected: _scope == tab.$1,
                                  button: true,
                                  child: Container(
                                    decoration: BoxDecoration(
                                      border: Border(
                                        bottom: BorderSide(
                                          color: _scope == tab.$1
                                              ? AppTheme.rust
                                              : Colors.transparent,
                                          width: 2,
                                        ),
                                      ),
                                    ),
                                    child: TextButton(
                                      key: ValueKey('library-tab-${tab.$1}'),
                                      onPressed: () => _changeScope(tab.$1),
                                      child: Padding(
                                        padding: const EdgeInsets.symmetric(
                                          horizontal: 4,
                                          vertical: 6,
                                        ),
                                        child: Text(tab.$2),
                                      ),
                                    ),
                                  ),
                                ),
                            ],
                          ),
                        ),
                      ),
                      const SizedBox(height: 20),
                      if (constraints.maxWidth < 600)
                        Column(
                          children: [
                            _searchField(),
                            const SizedBox(height: 12),
                            _formatField(),
                          ],
                        )
                      else
                        Row(
                          children: [
                            Expanded(child: _searchField()),
                            const SizedBox(width: 16),
                            SizedBox(width: 220, child: _formatField()),
                          ],
                        ),
                      const SizedBox(height: 22),
                      Wrap(
                        alignment: WrapAlignment.spaceBetween,
                        crossAxisAlignment: WrapCrossAlignment.center,
                        spacing: 12,
                        runSpacing: 8,
                        children: [
                          Text(switch (_scope) {
                            'saved' => 'Ready for your next trip',
                            'yours' => 'Your guides & works in progress',
                            _ => 'Recently shared',
                          }, style: Theme.of(context).textTheme.titleMedium),
                          Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              if (!_loading && _error == null)
                                Semantics(
                                  liveRegion: true,
                                  child: Text(
                                    '${documents.length} ${documents.length == 1 ? 'guide' : 'guides'}',
                                    style: Theme.of(
                                      context,
                                    ).textTheme.labelSmall,
                                  ),
                                ),
                              if (_scope == 'yours')
                                PopupMenuButton<String>(
                                  tooltip: 'Library options',
                                  onSelected: (_) async {
                                    await Navigator.push<void>(
                                      context,
                                      MaterialPageRoute(
                                        builder: (_) => Scaffold(
                                          appBar: AppBar(
                                            title: const Text('Manage files'),
                                          ),
                                          body: const DocumentsScreen(),
                                        ),
                                      ),
                                    );
                                    if (mounted) _load();
                                  },
                                  itemBuilder: (_) => [
                                    const PopupMenuItem(
                                      value: 'manage',
                                      child: Text('Manage files'),
                                    ),
                                  ],
                                ),
                            ],
                          ),
                        ],
                      ),
                      const SizedBox(height: 14),
                      if (_loading)
                        const Padding(
                          padding: EdgeInsets.all(32),
                          child: Center(child: CircularProgressIndicator()),
                        )
                      else if (_error != null)
                        _empty(_error!, retry: true)
                      else if (documents.isEmpty)
                        _empty(
                          _search.text.isNotEmpty || _format != 'all'
                              ? 'No guides match your search and filters.'
                              : switch (_scope) {
                                  'saved' =>
                                    'Save a guide to keep it close for your next trip.',
                                  'yours' =>
                                    'Bring in a favorite itinerary, spreadsheet, or a guide of your own.',
                                  _ =>
                                    'Guides shared by your friends will appear here.',
                                },
                        )
                      else
                        Wrap(
                          spacing: 20,
                          runSpacing: 20,
                          children: [
                            for (final document in documents)
                              SizedBox(
                                width: cardWidth,
                                child: LibraryDocumentCard(
                                  key: ValueKey('library-card-${document.id}'),
                                  document: document,
                                  isOwner:
                                      _identity?.isNotEmpty == true &&
                                      document.userId == _identity,
                                  saving: _saving.contains(document.id),
                                  onOpen: () => _open(document),
                                  onSave: () => _save(document),
                                ),
                              ),
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
  Widget _searchField() => TextField(
    key: const ValueKey('library-search'),
    controller: _search,
    decoration: const InputDecoration(
      prefixIcon: Icon(Icons.search),
      hintText: 'Where are you headed?',
      labelText: 'Search title, destination, or person',
    ),
    onChanged: (_) => setState(() {}),
  );
  Widget _formatField() => KeyedSubtree(
    key: ValueKey('library-format-state-$_format'),
    child: DropdownButtonFormField<String>(
      key: const ValueKey('library-format-filter'),
      initialValue: _format,
      isExpanded: true,
      decoration: const InputDecoration(labelText: 'Format'),
      items: const [
        DropdownMenuItem(value: 'all', child: Text('All formats')),
        DropdownMenuItem(value: 'spreadsheet', child: Text('Spreadsheets')),
        DropdownMenuItem(value: 'itinerary', child: Text('Itineraries')),
        DropdownMenuItem(value: 'guide', child: Text('Guides')),
        DropdownMenuItem(value: 'video', child: Text('Videos')),
      ],
      onChanged: (value) => setState(() => _format = value ?? 'all'),
    ),
  );
  Widget _empty(String message, {bool retry = false}) => Container(
    padding: const EdgeInsets.all(28),
    decoration: BoxDecoration(
      border: Border.all(color: AppTheme.outline),
      color: AppTheme.cream.withValues(alpha: .25),
    ),
    child: Column(
      children: [
        Text(message, textAlign: TextAlign.center),
        if (retry)
          Padding(
            padding: const EdgeInsets.only(top: 12),
            child: OutlinedButton(onPressed: _load, child: const Text('Retry')),
          ),
      ],
    ),
  );
}
