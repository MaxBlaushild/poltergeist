import 'package:flutter/material.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/community_poll_service.dart';

/// Bottom sheet for creating a community poll
class CommunityPollBottomSheet extends StatefulWidget {
  const CommunityPollBottomSheet({super.key});

  @override
  State<CommunityPollBottomSheet> createState() => _CommunityPollBottomSheetState();
}

class _CommunityPollBottomSheetState extends State<CommunityPollBottomSheet> {
  final _formKey = GlobalKey<FormState>();
  final _questionController = TextEditingController();
  final List<TextEditingController> _optionControllers = [];
  final _communityPollService = CommunityPollService(APIClient(ApiConstants.baseUrl));

  bool _isSubmitting = false;

  @override
  void initState() {
    super.initState();
    // Initialize with 3 option controllers
    for (int i = 0; i < 3; i++) {
      _optionControllers.add(TextEditingController());
    }
  }

  @override
  void dispose() {
    _questionController.dispose();
    for (var controller in _optionControllers) {
      controller.dispose();
    }
    super.dispose();
  }

  void _addOption() {
    if (_optionControllers.length < 10) {
      setState(() {
        _optionControllers.add(TextEditingController());
      });
    }
  }

  void _removeOption(int index) {
    if (_optionControllers.length > 3) {
      setState(() {
        _optionControllers[index].dispose();
        _optionControllers.removeAt(index);
      });
    }
  }

  Future<void> _handleSubmit() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    // Collect non-empty options
    final options = _optionControllers
        .map((controller) => controller.text.trim())
        .where((text) => text.isNotEmpty)
        .toList();

    // Validate we have 3-10 options
    if (options.length < 3 || options.length > 10) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Please provide between 3 and 10 options'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    setState(() {
      _isSubmitting = true;
    });

    try {
      await _communityPollService.createCommunityPoll(
        question: _questionController.text.trim(),
        options: options,
      );

      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Community poll created successfully!'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _isSubmitting = false;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to create poll: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Container(
      padding: const EdgeInsets.all(16.0),
      decoration: BoxDecoration(
        color: theme.scaffoldBackgroundColor,
        borderRadius: const BorderRadius.vertical(
          top: Radius.circular(16.0),
        ),
      ),
      child: SafeArea(
        child: Form(
          key: _formKey,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Handle bar
              Center(
                child: Container(
                  width: 40,
                  height: 4,
                  margin: const EdgeInsets.only(bottom: 16.0),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.onSurface.withOpacity(0.3),
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
              ),
              // Title
              Text(
                'Community Poll',
                style: theme.textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 24),
              // Scrollable content
              Flexible(
                child: SingleChildScrollView(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      // Question field
                      TextFormField(
                        controller: _questionController,
                        decoration: const InputDecoration(
                          labelText: 'Your Question',
                          hintText: 'e.g., Rank these Hawaiian islands: Maui, Big Island, Kauai, Oahu',
                          border: OutlineInputBorder(),
                        ),
                        maxLines: 3,
                        minLines: 2,
                        validator: (value) {
                          if (value == null || value.trim().isEmpty) {
                            return 'Please enter your question';
                          }
                          return null;
                        },
                      ),
                      const SizedBox(height: 24),
                      // Options section
                      Text(
                        'Options (3-10 required)',
                        style: theme.textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 12),
                      // Options list
                      ...List.generate(
                        _optionControllers.length,
                        (index) => Padding(
                          padding: const EdgeInsets.only(bottom: 12.0),
                          child: Row(
                            children: [
                              Expanded(
                                child: TextFormField(
                                  controller: _optionControllers[index],
                                  decoration: InputDecoration(
                                    labelText: 'Option ${index + 1}',
                                    hintText: 'Enter option text',
                                    border: const OutlineInputBorder(),
                                  ),
                                  validator: (value) {
                                    if (value == null || value.trim().isEmpty) {
                                      return 'Required';
                                    }
                                    return null;
                                  },
                                ),
                              ),
                              if (_optionControllers.length > 3)
                                IconButton(
                                  icon: const Icon(Icons.remove_circle_outline),
                                  color: theme.colorScheme.error,
                                  onPressed: () => _removeOption(index),
                                  tooltip: 'Remove option',
                                ),
                            ],
                          ),
                        ),
                      ),
                      const SizedBox(height: 12),
                      // Add option button
                      if (_optionControllers.length < 10)
                        OutlinedButton.icon(
                          onPressed: _addOption,
                          icon: const Icon(Icons.add),
                          label: const Text('Add Option'),
                        ),
                      const SizedBox(height: 24),
                    ],
                  ),
                ),
              ),
              // Submit button
              ElevatedButton(
                onPressed: _isSubmitting ? null : _handleSubmit,
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16.0),
                ),
                child: _isSubmitting
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                        ),
                      )
                    : const Text('Create Poll'),
              ),
              const SizedBox(height: 16),
            ],
          ),
        ),
      ),
    );
  }
}
