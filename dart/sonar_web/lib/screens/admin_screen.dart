import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../services/admin_service.dart';

class AdminScreen extends StatefulWidget {
  const AdminScreen({super.key});

  @override
  State<AdminScreen> createState() => _AdminScreenState();
}

class _AdminScreenState extends State<AdminScreen> {
  final _teamIdController = TextEditingController();
  final _pointOfInterestIdController = TextEditingController();
  final _quantityController = TextEditingController();
  bool _unlockLoading = false;
  bool _captureLoading = false;
  String? _error;
  String? _success;

  @override
  void dispose() {
    _teamIdController.dispose();
    _pointOfInterestIdController.dispose();
    _quantityController.dispose();
    super.dispose();
  }

  Future<void> _unlock() async {
    final teamId = _teamIdController.text.trim();
    final poiId = _pointOfInterestIdController.text.trim();
    if (teamId.isEmpty || poiId.isEmpty) return;
    setState(() {
      _error = null;
      _success = null;
      _unlockLoading = true;
    });
    try {
      await context.read<AdminService>().unlockPointOfInterestForTeam(
            teamId: teamId,
            pointOfInterestId: poiId,
          );
      if (mounted) {
        setState(() {
          _unlockLoading = false;
          _success = 'Unlocked successfully.';
        });
        _teamIdController.clear();
        _pointOfInterestIdController.clear();
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _unlockLoading = false;
          _error = e.toString();
        });
      }
    }
  }

  Future<void> _capture() async {
    final teamId = _teamIdController.text.trim();
    final poiId = _pointOfInterestIdController.text.trim();
    final q = _quantityController.text.trim();
    if (teamId.isEmpty || poiId.isEmpty || q.isEmpty) return;
    final tier = int.tryParse(q);
    if (tier == null) {
      setState(() => _error = 'Quantity must be an integer (tier).');
      return;
    }
    setState(() {
      _error = null;
      _success = null;
      _captureLoading = true;
    });
    try {
      await context.read<AdminService>().capturePointOfInterestForTeam(
            teamId: teamId,
            pointOfInterestId: poiId,
            tier: tier,
          );
      if (mounted) {
        setState(() {
          _captureLoading = false;
          _success = 'Capture successful.';
        });
        _teamIdController.clear();
        _pointOfInterestIdController.clear();
        _quantityController.clear();
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _captureLoading = false;
          _error = e.toString();
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Admin')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (_error != null)
              Padding(
                padding: const EdgeInsets.only(bottom: 16),
                child: Text(
                  _error!,
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              ),
            if (_success != null)
              Padding(
                padding: const EdgeInsets.only(bottom: 16),
                child: Text(
                  _success!,
                  style: TextStyle(color: Colors.green.shade700),
                ),
              ),
            Text(
              'Team & POI',
              style: Theme.of(context).textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const Divider(),
            const SizedBox(height: 8),
            TextField(
              controller: _teamIdController,
              decoration: const InputDecoration(
                labelText: 'Team ID',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _pointOfInterestIdController,
              decoration: const InputDecoration(
                labelText: 'Point of Interest ID',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 24),
            Text(
              'Unlock point for team',
              style: Theme.of(context).textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const Divider(),
            const SizedBox(height: 8),
            FilledButton(
              onPressed: _unlockLoading ||
                      _teamIdController.text.trim().isEmpty ||
                      _pointOfInterestIdController.text.trim().isEmpty
                  ? null
                  : _unlock,
              child: Text(_unlockLoading ? 'Unlocking…' : 'Unlock'),
            ),
            const SizedBox(height: 24),
            Text(
              'Capture for team',
              style: Theme.of(context).textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const Divider(),
            const SizedBox(height: 8),
            TextField(
              controller: _quantityController,
              decoration: const InputDecoration(
                labelText: 'Quantity (tier)',
                border: OutlineInputBorder(),
              ),
              keyboardType: TextInputType.number,
            ),
            const SizedBox(height: 12),
            FilledButton(
              onPressed: _captureLoading ||
                      _teamIdController.text.trim().isEmpty ||
                      _pointOfInterestIdController.text.trim().isEmpty ||
                      _quantityController.text.trim().isEmpty
                  ? null
                  : _capture,
              child: Text(_captureLoading ? 'Capturing…' : 'Capture'),
            ),
          ],
        ),
      ),
    );
  }
}
