import 'package:flutter/material.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/quick_decision_service.dart';

/// Bottom sheet for creating a quick decision request
class QuickDecisionBottomSheet extends StatefulWidget {
  const QuickDecisionBottomSheet({super.key});

  @override
  State<QuickDecisionBottomSheet> createState() => _QuickDecisionBottomSheetState();
}

class _QuickDecisionBottomSheetState extends State<QuickDecisionBottomSheet> {
  final _formKey = GlobalKey<FormState>();
  final _questionController = TextEditingController();
  final _option1Controller = TextEditingController();
  final _option2Controller = TextEditingController();
  final _option3Controller = TextEditingController();
  final _quickDecisionService = QuickDecisionService(APIClient(ApiConstants.baseUrl));

  bool _isSubmitting = false;

  @override
  void dispose() {
    _questionController.dispose();
    _option1Controller.dispose();
    _option2Controller.dispose();
    _option3Controller.dispose();
    super.dispose();
  }

  Future<void> _handleSubmit() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    setState(() {
      _isSubmitting = true;
    });

    try {
      await _quickDecisionService.createQuickDecisionRequest(
        question: _questionController.text.trim(),
        option1: _option1Controller.text.trim(),
        option2: _option2Controller.text.trim(),
        option3: _option3Controller.text.trim().isEmpty
            ? null
            : _option3Controller.text.trim(),
      );

      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Quick decision request submitted successfully!'),
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
            content: Text('Failed to submit request: ${e.toString()}'),
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
                'Quick Decision',
                style: theme.textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 24),
              // Question field
              TextFormField(
                controller: _questionController,
                decoration: const InputDecoration(
                  labelText: 'Your Question',
                  hintText: 'e.g., Which restaurant should I try in Tokyo: Sukiyabashi Jiro or Sushi Saito?',
                  border: OutlineInputBorder(),
                ),
                maxLines: 4,
                minLines: 3,
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'Please enter your question';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),
              // Option 1 field
              TextFormField(
                controller: _option1Controller,
                decoration: const InputDecoration(
                  labelText: 'Option 1',
                  hintText: 'e.g., Sukiyabashi Jiro',
                  border: OutlineInputBorder(),
                ),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'Please enter option 1';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),
              // Option 2 field
              TextFormField(
                controller: _option2Controller,
                decoration: const InputDecoration(
                  labelText: 'Option 2',
                  hintText: 'e.g., Sushi Saito',
                  border: OutlineInputBorder(),
                ),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'Please enter option 2';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),
              // Option 3 field (optional)
              TextFormField(
                controller: _option3Controller,
                decoration: const InputDecoration(
                  labelText: 'Option 3 (optional)',
                  hintText: 'Enter a third option if needed',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 24),
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
                    : const Text('Submit'),
              ),
              const SizedBox(height: 16),
            ],
          ),
        ),
      ),
    );
  }
}
