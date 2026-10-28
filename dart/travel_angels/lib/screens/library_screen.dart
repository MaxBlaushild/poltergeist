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
  final _searchFocus = FocusNode();
  bool _searchExpanded = false;
  static const _formats = {
    'all': 'All formats',
    'spreadsheet': 'Spreadsheets',
    'itinerary': 'Itineraries',
    'guide': 'Guides',
    'video': 'Videos',
  };
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
    _searchFocus.dispose();
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
            _searchExpanded = false;
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

  void _toggleSearch() {
    setState(() => _searchExpanded = !_searchExpanded);
    if (_searchExpanded) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && _searchExpanded) _searchFocus.requestFocus();
      });
    } else {
      _searchFocus.unfocus();
    }
  }

  @override
  Widget build(BuildContext context) => SafeArea(
    child: SingleChildScrollView(
      child: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 1200),
          child: Padding(
            padding: EdgeInsets.all(
              MediaQuery.sizeOf(context).width < 600 ? 16 : 24,
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
                final compact =
                    constraints.maxWidth < 600 ||
                    MediaQuery.textScalerOf(context).scale(14) > 21;
                return Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: Text(
                            'Travel library',
                            style: Theme.of(context).textTheme.headlineMedium,
                          ),
                        ),
                        if (!compact && !_loading && _error == null) ...[
                          Text(
                            '${documents.length} ${documents.length == 1 ? 'guide' : 'guides'}',
                            style: Theme.of(context).textTheme.labelSmall,
                          ),
                          const SizedBox(width: 16),
                        ],
                        const SizedBox(width: 8),
                        _addButton(compact),
                        if (_scope == 'yours') _libraryOptions(),
                      ],
                    ),
                    const SizedBox(height: 4),
                    Container(
                      decoration: const BoxDecoration(
                        border: Border(
                          bottom: BorderSide(color: AppTheme.outline),
                        ),
                      ),
                      child: Row(
                        children: [
                          Expanded(child: _scopeTabs()),
                          const SizedBox(width: 4),
                          IconButton(
                            key: const ValueKey('library-search-toggle'),
                            tooltip: _searchExpanded
                                ? 'Hide search'
                                : 'Search library',
                            constraints: const BoxConstraints(
                              minWidth: 44,
                              minHeight: 44,
                            ),
                            iconSize: 20,
                            color:
                                _searchExpanded ||
                                    _search.text.trim().isNotEmpty
                                ? AppTheme.rust
                                : AppTheme.ink,
                            icon: Icon(
                              _searchExpanded ? Icons.close : Icons.search,
                            ),
                            onPressed: _toggleSearch,
                          ),
                          _formatMenu(),
                        ],
                      ),
                    ),
                    if (_searchExpanded) ...[
                      const SizedBox(height: 12),
                      _searchField(),
                    ],
                    if (_format != 'all' ||
                        (!_searchExpanded &&
                            _search.text.trim().isNotEmpty)) ...[
                      const SizedBox(height: 8),
                      Wrap(
                        spacing: 8,
                        runSpacing: 4,
                        children: [
                          if (_format != 'all')
                            InputChip(
                              label: Text(_formats[_format]!),
                              deleteButtonTooltipMessage: 'Clear format filter',
                              onDeleted: () => setState(() => _format = 'all'),
                            ),
                          if (!_searchExpanded &&
                              _search.text.trim().isNotEmpty)
                            ConstrainedBox(
                              constraints: BoxConstraints(
                                maxWidth: constraints.maxWidth,
                              ),
                              child: InputChip(
                                label: Text(
                                  'Search: ${_search.text.trim()}',
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                ),
                                deleteButtonTooltipMessage: 'Clear search',
                                onPressed: _toggleSearch,
                                onDeleted: () => setState(_search.clear),
                              ),
                            ),
                        ],
                      ),
                    ],
                    const SizedBox(height: 16),
                    Semantics(
                      liveRegion: true,
                      label: _loading
                          ? 'Loading guides'
                          : '${documents.length} guides',
                      child: _loading
                          ? const Padding(
                              padding: EdgeInsets.all(32),
                              child: Center(child: CircularProgressIndicator()),
                            )
                          : _error != null
                          ? _empty(_error!, retry: true)
                          : documents.isEmpty
                          ? _empty(
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
                          : Wrap(
                              spacing: 20,
                              runSpacing: 20,
                              children: [
                                for (final document in documents)
                                  SizedBox(
                                    width: cardWidth,
                                    child: LibraryDocumentCard(
                                      key: ValueKey(
                                        'library-card-${document.id}',
                                      ),
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
  );

  Widget _addButton(bool compact) => compact
      ? IconButton(
          key: const ValueKey('library-add-guide'),
          tooltip: 'Add a guide',
          onPressed: _adding ? null : _add,
          icon: const Icon(Icons.add, size: 22),
          color: AppTheme.rust,
          constraints: const BoxConstraints(minWidth: 44, minHeight: 44),
        )
      : TextButton.icon(
          key: const ValueKey('library-add-guide'),
          onPressed: _adding ? null : _add,
          icon: const Icon(Icons.add, size: 18),
          label: const Text('Add guide'),
          style: TextButton.styleFrom(
            foregroundColor: AppTheme.rust,
            padding: const EdgeInsets.symmetric(horizontal: 10),
            minimumSize: const Size(44, 44),
          ),
        );

  Widget _scopeTabs() => SingleChildScrollView(
    scrollDirection: Axis.horizontal,
    child: Row(
      children: [
        for (final tab in [
          ('network', 'Network'),
          ('saved', 'Saved'),
          ('yours', 'Yours'),
        ])
          Semantics(
            selected: _scope == tab.$1,
            button: true,
            label: tab.$1 == 'network' ? 'From your network' : null,
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
                style: TextButton.styleFrom(
                  minimumSize: const Size(44, 44),
                  padding: const EdgeInsets.symmetric(horizontal: 10),
                  foregroundColor: _scope == tab.$1
                      ? AppTheme.rust
                      : AppTheme.ink,
                ),
                child: Text(tab.$2),
              ),
            ),
          ),
      ],
    ),
  );

  Widget _searchField() => TextField(
    key: const ValueKey('library-search'),
    controller: _search,
    focusNode: _searchFocus,
    textInputAction: TextInputAction.search,
    style: Theme.of(context).textTheme.bodyMedium,
    decoration: InputDecoration(
      isDense: true,
      hintText: 'Search title, destination, or person',
      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      prefixIcon: const Icon(Icons.search, size: 18),
      prefixIconConstraints: const BoxConstraints(minWidth: 40, minHeight: 44),
      suffixIcon: _search.text.isEmpty
          ? null
          : IconButton(
              tooltip: 'Clear search',
              icon: const Icon(Icons.clear, size: 18),
              constraints: const BoxConstraints(minWidth: 44, minHeight: 44),
              onPressed: () => setState(_search.clear),
            ),
    ),
    onSubmitted: (_) => _searchFocus.unfocus(),
    onChanged: (_) => setState(() {}),
  );

  Widget _formatMenu() => PopupMenuButton<String>(
    key: const ValueKey('library-format-filter'),
    tooltip: _format == 'all'
        ? 'Filter by format'
        : 'Format: ${_formats[_format]}',
    initialValue: _format,
    iconSize: 20,
    icon: Badge(
      isLabelVisible: _format != 'all',
      smallSize: 6,
      backgroundColor: AppTheme.rust,
      child: Icon(
        Icons.tune,
        color: _format == 'all' ? AppTheme.ink : AppTheme.rust,
      ),
    ),
    onSelected: (value) {
      _searchFocus.unfocus();
      setState(() => _format = value);
    },
    itemBuilder: (_) => [
      for (final format in _formats.entries)
        CheckedPopupMenuItem(
          value: format.key,
          checked: _format == format.key,
          child: Text(format.value),
        ),
    ],
  );

  Widget _libraryOptions() => PopupMenuButton<String>(
    tooltip: 'Library options',
    iconSize: 20,
    onSelected: (_) async {
      await Navigator.push<void>(
        context,
        MaterialPageRoute(
          builder: (_) => Scaffold(
            appBar: AppBar(title: const Text('Manage files')),
            body: const DocumentsScreen(),
          ),
        ),
      );
      if (mounted) _load();
    },
    itemBuilder: (_) => [
      const PopupMenuItem(value: 'manage', child: Text('Manage files')),
    ],
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
