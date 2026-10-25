import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/document_service.dart';
import 'package:travel_angels/services/friend_service.dart';
import 'package:travel_angels/theme/app_theme.dart';

String libraryError(
  Object error, [
  String fallback = 'Something went wrong. Please try again.',
]) {
  if (error is DioException) {
    final status = error.response?.statusCode;
    if (status == 403 || status == 404) {
      return 'This guide is no longer available to you.';
    }
    final data = error.response?.data;
    if (data is Map && data['error'] is String) return data['error'] as String;
  }
  return fallback;
}

bool libraryAccessDenied(Object error) =>
    error is DioException &&
    [401, 403, 404].contains(error.response?.statusCode);

class LibraryShareDialog extends StatefulWidget {
  const LibraryShareDialog({
    super.key,
    required this.document,
    required this.service,
    this.friendService,
  });
  final Document document;
  final DocumentService service;
  final FriendService? friendService;
  @override
  State<LibraryShareDialog> createState() => _LibraryShareDialogState();
}

class _LibraryShareDialogState extends State<LibraryShareDialog> {
  late String _visibility = widget.document.visibility;
  late final Set<String> _selected = widget.document.sharedWithUserIds.toSet();
  List<User> _friends = [];
  bool _loading = true;
  bool _saving = false;
  String? _loadError;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _loadError = null;
    });
    try {
      final friends =
          await (widget.friendService ??
                  FriendService(APIClient(ApiConstants.baseUrl)))
              .getFriends();
      if (!mounted) return;
      setState(
        () => _friends = friends
            .where((friend) => friend.id?.isNotEmpty == true)
            .toList(),
      );
    } catch (error) {
      if (mounted) {
        setState(() => _loadError = 'Your friends could not be loaded.');
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _save() async {
    if (_visibility == 'selected' && _selected.isEmpty) {
      setState(() => _error = 'Choose at least one friend, or select Only me.');
      return;
    }
    setState(() {
      _saving = true;
      _error = null;
    });
    try {
      final updated = await widget.service.updateDocumentSharing(
        widget.document.id,
        visibility: _visibility,
        userIds: _visibility == 'selected' ? _selected.toList() : [],
      );
      if (mounted) Navigator.pop(context, Document.fromJson(updated));
    } catch (error) {
      if (mounted) {
        setState(
          () => _error = libraryError(
            error,
            'Sharing could not be saved. Please try again.',
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final missing = _selected.difference(
      _friends.map((user) => user.id!).toSet(),
    );
    return PopScope(
      canPop: !_saving,
      child: AlertDialog(
        insetPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 24),
        title: const Text('Who can see this?'),
        scrollable: true,
        content: SizedBox(
          width: 460,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              DropdownButtonFormField<String>(
                key: const ValueKey('library-sharing-visibility'),
                initialValue: _visibility,
                isExpanded: true,
                decoration: const InputDecoration(labelText: 'Audience'),
                items: const [
                  DropdownMenuItem(value: 'private', child: Text('Only me')),
                  DropdownMenuItem(value: 'network', child: Text('My network')),
                  DropdownMenuItem(
                    value: 'selected',
                    child: Text('Selected friends'),
                  ),
                ],
                onChanged: _saving
                    ? null
                    : (value) => setState(() {
                        _visibility = value!;
                        _error = null;
                      }),
              ),
              const SizedBox(height: 16),
              const Text(
                'Shares the guide saved in Travel Angels. Access to the original file stays separate.',
              ),
              if (_visibility == 'network')
                const Padding(
                  padding: EdgeInsets.only(top: 12),
                  child: Text(
                    'Your current friends can read it. New friends will be able to read it too.',
                  ),
                ),
              if (_visibility == 'selected') ...[
                const SizedBox(height: 16),
                if (_loading)
                  const LinearProgressIndicator()
                else if (_loadError != null)
                  Column(
                    children: [
                      Text(_loadError!),
                      TextButton(onPressed: _load, child: const Text('Retry')),
                    ],
                  )
                else ...[
                  if (_friends.isEmpty)
                    const Text(
                      'Your network is empty. Add friends in My Circle to share a guide with them.',
                    ),
                  for (final friend in _friends)
                    CheckboxListTile(
                      key: ValueKey('library-sharing-friend-${friend.id}'),
                      contentPadding: EdgeInsets.zero,
                      value: _selected.contains(friend.id),
                      title: Text(
                        friend.name?.isNotEmpty == true
                            ? friend.name!
                            : friend.username ?? 'Friend',
                      ),
                      onChanged: _saving
                          ? null
                          : (value) => setState(() {
                              if (value == true) {
                                _selected.add(friend.id!);
                              } else {
                                _selected.remove(friend.id!);
                              }
                            }),
                    ),
                  for (final id in missing)
                    CheckboxListTile(
                      value: true,
                      contentPadding: EdgeInsets.zero,
                      title: const Text('Previously selected person'),
                      subtitle: const Text(
                        'No longer in your network. Uncheck to remove this selection.',
                      ),
                      onChanged: _saving
                          ? null
                          : (_) => setState(() => _selected.remove(id)),
                    ),
                ],
              ],
              if (_error != null)
                Padding(
                  padding: const EdgeInsets.only(top: 16),
                  child: Text(
                    _error!,
                    style: const TextStyle(color: AppTheme.error),
                  ),
                ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: _saving ? null : () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          FilledButton(
            key: const ValueKey('library-sharing-save'),
            onPressed:
                _saving ||
                    (_visibility == 'selected' &&
                        (_loading || _loadError != null))
                ? null
                : _save,
            child: Text(_saving ? 'Saving…' : 'Save sharing'),
          ),
        ],
      ),
    );
  }
}

class LibraryTripPickerDialog extends StatefulWidget {
  const LibraryTripPickerDialog({
    super.key,
    required this.document,
    required this.service,
  });
  final Document document;
  final DocumentService service;
  @override
  State<LibraryTripPickerDialog> createState() =>
      _LibraryTripPickerDialogState();
}

class _LibraryTripPickerDialogState extends State<LibraryTripPickerDialog> {
  List<Map<String, dynamic>> _trips = [];
  final _selected = <String>{};
  final _saved = <String>{};
  bool _loading = true;
  bool _saving = false;
  String? _error;
  Document? _updated;
  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final trips = await widget.service.getDocumentTrips(widget.document.id);
      if (!mounted) return;
      setState(() {
        _trips = trips;
        _saved.clear();
        _saved.addAll(
          trips
              .where((trip) => trip['isSaved'] == true)
              .map((trip) => trip['id'].toString()),
        );
        _selected.clear();
        _selected.addAll(_saved);
      });
    } catch (error) {
      if (mounted) {
        setState(
          () => _error = libraryError(error, 'Your trips could not be loaded.'),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _save() async {
    setState(() {
      _saving = true;
      _error = null;
    });
    try {
      for (final trip in _trips) {
        final id = trip['id'].toString();
        if (_selected.contains(id) == _saved.contains(id)) continue;
        final updated = await widget.service.setDocumentTrip(
          widget.document.id,
          id,
          _selected.contains(id),
        );
        _updated = Document.fromJson(updated);
        if (_selected.contains(id)) {
          _saved.add(id);
        } else {
          _saved.remove(id);
        }
      }
      if (mounted) Navigator.pop(context, _updated ?? widget.document);
    } catch (error) {
      if (mounted) {
        setState(
          () => _error = libraryError(
            error,
            'Some trip references could not be saved. Your completed changes are kept; retry to finish.',
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) => PopScope(
    canPop: !_saving,
    child: AlertDialog(
      insetPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 24),
      title: const Text('Save to a trip'),
      scrollable: true,
      content: SizedBox(
        width: 480,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Adds a personal reference. It does not change who can access the guide or share it with the people on the trip.',
            ),
            const SizedBox(height: 16),
            if (_loading)
              const LinearProgressIndicator()
            else if (_trips.isEmpty && _error == null)
              const Text(
                'No trips available yet. Create or join a trip from your calendar first.',
              ),
            for (final trip in _trips)
              CheckboxListTile(
                key: ValueKey('library-trip-${trip['id']}'),
                contentPadding: EdgeInsets.zero,
                title: Text(
                  (trip['title'] as String?)?.isNotEmpty == true
                      ? trip['title'] as String
                      : 'Trip',
                ),
                subtitle: Text(
                  [
                        trip['destination'],
                        if ((trip['startDate'] ?? '').toString().isNotEmpty)
                          '${trip['startDate']} – ${trip['endDate']}',
                      ]
                      .where(
                        (value) => value != null && value.toString().isNotEmpty,
                      )
                      .join('\n'),
                ),
                value: _selected.contains(trip['id'].toString()),
                onChanged: _saving
                    ? null
                    : (value) => setState(() {
                        final id = trip['id'].toString();
                        if (value == true) {
                          _selected.add(id);
                        } else {
                          _selected.remove(id);
                        }
                      }),
              ),
            if (_error != null) ...[
              const SizedBox(height: 12),
              Text(_error!, style: const TextStyle(color: AppTheme.error)),
              if (_trips.isEmpty)
                TextButton(onPressed: _load, child: const Text('Retry')),
            ],
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: _saving ? null : () => Navigator.pop(context),
          child: const Text('Cancel'),
        ),
        FilledButton(
          key: const ValueKey('library-trips-save'),
          onPressed: _saving || _loading || _trips.isEmpty ? null : _save,
          child: Text(_saving ? 'Saving…' : 'Save to trips'),
        ),
      ],
    ),
  );
}
