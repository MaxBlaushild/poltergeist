import 'package:flutter/material.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';

/// A scoped preference link is sufficient to unsubscribe; no account is needed.
class TravelPreferencesScreen extends StatefulWidget {
  const TravelPreferencesScreen({super.key, required this.token, this.service});
  final String token;
  final TravelCalendarService? service;

  @override
  State<TravelPreferencesScreen> createState() =>
      _TravelPreferencesScreenState();
}

class _TravelPreferencesScreenState extends State<TravelPreferencesScreen> {
  late final _service = widget.service ?? TravelCalendarService();
  bool _loading = true;
  bool _saving = false;
  String? _error;
  String _email = '';
  List<Map<String, dynamic>> _subscriptions = [];
  Set<String> _selected = {};
  bool _participationNotices = true;

  String get _path =>
      '/calendar-guests/preferences/${Uri.encodeComponent(widget.token)}';

  @override
  void initState() {
    super.initState();
    _load();
  }

  void _apply(Map<String, dynamic> data) {
    _email = data['email']?.toString() ?? '';
    _participationNotices = data['participationNotices'] != false;
    _subscriptions = (data['subscriptions'] as List? ?? [])
        .whereType<Map>()
        .map((item) => Map<String, dynamic>.from(item))
        .toList();
    _selected = _subscriptions
        .where((item) => item['status'] == 'active')
        .map((item) => item['id'].toString())
        .toSet();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final data = await _service.get(_path);
      if (mounted) setState(() => _apply(data));
    } catch (error) {
      if (mounted) setState(() => _error = error.toString());
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _save({bool unsubscribeAll = false}) async {
    setState(() {
      _saving = true;
      _error = null;
    });
    try {
      final data = await _service.put(
        _path,
        data: {
          'subscriptionIds': _selected.toList(),
          'unsubscribeAll': unsubscribeAll,
          'participationNotices': _participationNotices,
        },
      );
      if (!mounted) return;
      setState(() => _apply(data));
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            unsubscribeAll
                ? 'Unsubscribed from travel updates.'
                : 'Preferences saved.',
          ),
        ),
      );
    } catch (error) {
      if (mounted) setState(() => _error = error.toString());
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Travel update preferences')),
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 640),
          child: _loading
              ? const CircularProgressIndicator()
              : ListView(
                  padding: const EdgeInsets.all(24),
                  children: [
                    const Text('Manage your updates without signing in.'),
                    if (_email.isNotEmpty)
                      Padding(
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        child: Text(
                          _email,
                          style: Theme.of(context).textTheme.titleMedium,
                        ),
                      ),
                    if (_error != null) ...[
                      Text(
                        _error!,
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.error,
                        ),
                      ),
                      TextButton(
                        onPressed: _load,
                        child: const Text('Try again'),
                      ),
                    ],
                    if (_email.isNotEmpty) ...[
                      if (_subscriptions.isEmpty)
                        const Padding(
                          padding: EdgeInsets.symmetric(vertical: 20),
                          child: Text(
                            'You have no travel update subscriptions.',
                          ),
                        ),
                      for (final item in _subscriptions)
                        CheckboxListTile(
                          contentPadding: EdgeInsets.zero,
                          title: Text(
                            item['title']?.toString() ??
                                (item['stopId'] != null
                                    ? 'Selected stop updates'
                                    : 'Calendar updates'),
                          ),
                          subtitle: Text(
                            item['status'] == 'paused'
                                ? 'Paused after access changed. Reopen an authorized stop to follow again.'
                                : item['status'] == 'unsubscribed'
                                ? 'Unsubscribed. Follow again from the calendar or stop.'
                                : 'Only includes plans you have permission to view.',
                          ),
                          value: _selected.contains(item['id'].toString()),
                          onChanged: _saving || item['status'] != 'active'
                              ? null
                              : (value) => setState(() {
                                  if (value == true) {
                                    _selected.add(item['id'].toString());
                                  } else {
                                    _selected.remove(item['id'].toString());
                                  }
                                }),
                        ),
                      const Divider(),
                      SwitchListTile(
                        contentPadding: EdgeInsets.zero,
                        title: const Text('Participation notices'),
                        subtitle: const Text(
                          'Approvals and changes to stops you requested or joined. Withdrawing from a stop ends its participation notices.',
                        ),
                        value: _participationNotices,
                        onChanged: _saving
                            ? null
                            : (value) =>
                                  setState(() => _participationNotices = value),
                      ),
                      const SizedBox(height: 24),
                      FilledButton(
                        onPressed: _saving ? null : _save,
                        child: Text(_saving ? 'Saving…' : 'Save preferences'),
                      ),
                      const SizedBox(height: 12),
                      OutlinedButton(
                        onPressed: _saving
                            ? null
                            : () => _save(unsubscribeAll: true),
                        child: const Text(
                          'Unsubscribe from all travel updates',
                        ),
                      ),
                    ],
                  ],
                ),
        ),
      ),
    );
  }
}
