import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../services/poi_service.dart';

class CreatePoiScreen extends StatefulWidget {
  const CreatePoiScreen({super.key});

  @override
  State<CreatePoiScreen> createState() => _CreatePoiScreenState();
}

class _CreatePoiScreenState extends State<CreatePoiScreen> {
  final _formKey = GlobalKey<FormState>();
  final _groupIdController = TextEditingController();
  final _nameController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _imageUrlController = TextEditingController();
  final _latController = TextEditingController();
  final _lonController = TextEditingController();
  final _clueController = TextEditingController();
  final _unlockTierController = TextEditingController();
  bool _submitting = false;
  String? _error;
  String? _success;

  @override
  void dispose() {
    _groupIdController.dispose();
    _nameController.dispose();
    _descriptionController.dispose();
    _imageUrlController.dispose();
    _latController.dispose();
    _lonController.dispose();
    _clueController.dispose();
    _unlockTierController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() {
      _error = null;
      _success = null;
      _submitting = true;
    });
    if (!_formKey.currentState!.validate()) {
      setState(() => _submitting = false);
      return;
    }
    final groupId = _groupIdController.text.trim();
    final unlockTierStr = _unlockTierController.text.trim();
    int? unlockTier;
    if (unlockTierStr.isNotEmpty) {
      unlockTier = int.tryParse(unlockTierStr);
    }

    try {
      await context.read<PoiService>().createPointOfInterestForGroup(
            groupId,
            name: _nameController.text.trim(),
            description: _descriptionController.text.trim(),
            imageUrl: _imageUrlController.text.trim(),
            latitude: _latController.text.trim(),
            longitude: _lonController.text.trim(),
            clue: _clueController.text.trim(),
            unlockTier: unlockTier,
          );
      if (mounted) {
        setState(() {
          _submitting = false;
          _success = 'Point of interest created successfully.';
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _submitting = false;
          _error = e.toString();
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Create Point of Interest')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Form(
          key: _formKey,
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
              TextFormField(
                controller: _groupIdController,
                decoration: const InputDecoration(
                  labelText: 'Point of Interest Group ID (UUID)',
                  border: OutlineInputBorder(),
                ),
                validator: (v) {
                  if (v == null || v.trim().isEmpty) {
                    return 'Group ID is required';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _nameController,
                decoration: const InputDecoration(
                  labelText: 'Name',
                  border: OutlineInputBorder(),
                ),
                validator: (v) =>
                    (v == null || v.trim().isEmpty) ? 'Required' : null,
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _descriptionController,
                decoration: const InputDecoration(
                  labelText: 'Description',
                  border: OutlineInputBorder(),
                ),
                maxLines: 3,
                validator: (v) =>
                    (v == null || v.trim().isEmpty) ? 'Required' : null,
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _imageUrlController,
                decoration: const InputDecoration(
                  labelText: 'Image URL',
                  border: OutlineInputBorder(),
                ),
                validator: (v) =>
                    (v == null || v.trim().isEmpty) ? 'Required' : null,
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  Expanded(
                    child: TextFormField(
                      controller: _latController,
                      decoration: const InputDecoration(
                        labelText: 'Latitude',
                        border: OutlineInputBorder(),
                      ),
                      validator: (v) =>
                          (v == null || v.trim().isEmpty) ? 'Required' : null,
                    ),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: TextFormField(
                      controller: _lonController,
                      decoration: const InputDecoration(
                        labelText: 'Longitude',
                        border: OutlineInputBorder(),
                      ),
                      validator: (v) =>
                          (v == null || v.trim().isEmpty) ? 'Required' : null,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _clueController,
                decoration: const InputDecoration(
                  labelText: 'Undiscovered',
                  border: OutlineInputBorder(),
                ),
                validator: (v) =>
                    (v == null || v.trim().isEmpty) ? 'Required' : null,
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _unlockTierController,
                decoration: const InputDecoration(
                  labelText: 'Unlock tier (optional)',
                  border: OutlineInputBorder(),
                ),
                keyboardType: TextInputType.number,
              ),
              const SizedBox(height: 24),
              FilledButton(
                onPressed: _submitting ? null : _submit,
                child: Text(_submitting ? 'Creating…' : 'Create Point of Interest'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
