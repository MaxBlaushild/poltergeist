import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:travel_angels/services/credits_service.dart';
import 'package:url_launcher/url_launcher.dart';

class CreditsPurchaseDialog extends StatefulWidget {
  final CreditsService creditsService;
  final VoidCallback? onPurchaseComplete;

  const CreditsPurchaseDialog({
    super.key,
    required this.creditsService,
    this.onPurchaseComplete,
  });

  @override
  State<CreditsPurchaseDialog> createState() => _CreditsPurchaseDialogState();
}

class _CreditsPurchaseDialogState extends State<CreditsPurchaseDialog> {
  final _customAmountController = TextEditingController();
  bool _isCustomAmount = false;
  bool _isProcessing = false;
  String? _error;

  // Predefined credit packages
  final List<CreditPackage> _packages = [
    CreditPackage(credits: 10, price: 10),
    CreditPackage(credits: 25, price: 25),
    CreditPackage(credits: 50, price: 50),
    CreditPackage(credits: 100, price: 100),
  ];

  @override
  void dispose() {
    _customAmountController.dispose();
    super.dispose();
  }

  /// Extracts error message from DioException response body
  String _extractErrorMessage(dynamic error) {
    if (error is DioException) {
      // Try to get error message from response body
      if (error.response?.data != null) {
        final data = error.response!.data;
        if (data is Map<String, dynamic> && data.containsKey('error')) {
          return data['error'] as String;
        }
        if (data is String) {
          return data;
        }
      }
      // Fall back to DioException message or status code
      if (error.response != null) {
        return 'Server error (${error.response!.statusCode}): ${error.message ?? 'Unknown error'}';
      }
      return error.message ?? 'Network error occurred';
    }
    // Fall back for non-Dio errors
    return error.toString();
  }

  Future<void> _handlePurchase(int amountInDollars) async {
    if (amountInDollars < 1 || amountInDollars > 1000) {
      setState(() {
        _error = 'Amount must be between \$1 and \$1000';
      });
      return;
    }

    setState(() {
      _isProcessing = true;
      _error = null;
    });

    try {
      final checkoutUrl = await widget.creditsService.purchaseCredits(amountInDollars);
      
      // Launch Stripe checkout URL
      final uri = Uri.parse(checkoutUrl);
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
        
        // Close dialog
        if (mounted) {
          Navigator.of(context).pop();
        }
        
        // Show success message
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Redirecting to payment...'),
              duration: Duration(seconds: 2),
            ),
          );
        }
      } else {
        throw Exception('Could not launch checkout URL');
      }
    } catch (e) {
      setState(() {
        _error = 'Failed to initiate purchase: ${_extractErrorMessage(e)}';
        _isProcessing = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return AlertDialog(
      title: const Text('Buy Credits'),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              '1 dollar = 1 credit',
              style: TextStyle(fontSize: 14, color: Colors.grey),
            ),
            const SizedBox(height: 16),
            // Predefined packages
            if (!_isCustomAmount) ...[
              Text(
                'Select a package:',
                style: theme.textTheme.titleSmall,
              ),
              const SizedBox(height: 12),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: _packages.map((package) {
                  return InkWell(
                    onTap: _isProcessing
                        ? null
                        : () => _handlePurchase(package.price),
                    child: Container(
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        border: Border.all(color: theme.colorScheme.outline),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Column(
                        children: [
                          Text(
                            '${package.credits}',
                            style: theme.textTheme.titleLarge?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          Text(
                            'credits',
                            style: theme.textTheme.bodySmall,
                          ),
                          const SizedBox(height: 4),
                          Text(
                            '\$${package.price}',
                            style: theme.textTheme.titleMedium?.copyWith(
                              color: theme.colorScheme.primary,
                            ),
                          ),
                        ],
                      ),
                    ),
                  );
                }).toList(),
              ),
              const SizedBox(height: 16),
              TextButton(
                onPressed: _isProcessing
                    ? null
                    : () {
                        setState(() {
                          _isCustomAmount = true;
                        });
                      },
                child: const Text('Enter custom amount'),
              ),
            ] else ...[
              // Custom amount input
              TextField(
                controller: _customAmountController,
                keyboardType: TextInputType.number,
                decoration: const InputDecoration(
                  labelText: 'Amount in dollars',
                  hintText: 'Enter amount (1-1000)',
                  prefixText: '\$',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  TextButton(
                    onPressed: _isProcessing
                        ? null
                        : () {
                            setState(() {
                              _isCustomAmount = false;
                              _customAmountController.clear();
                            });
                          },
                    child: const Text('Back to packages'),
                  ),
                ],
              ),
            ],
            if (_error != null) ...[
              const SizedBox(height: 12),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.red.shade50,
                  borderRadius: BorderRadius.circular(4),
                  border: Border.all(color: Colors.red.shade200),
                ),
                child: Text(
                  _error!,
                  style: TextStyle(
                    color: Colors.red.shade700,
                    fontSize: 14,
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
      actions: [
        if (_isCustomAmount) ...[
          TextButton(
            onPressed: _isProcessing
                ? null
                : () {
                    Navigator.of(context).pop();
                  },
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: _isProcessing
                ? null
                : () {
                    final amountText = _customAmountController.text.trim();
                    if (amountText.isEmpty) {
                      setState(() {
                        _error = 'Please enter an amount';
                      });
                      return;
                    }
                    final amount = int.tryParse(amountText);
                    if (amount == null) {
                      setState(() {
                        _error = 'Please enter a valid number';
                      });
                      return;
                    }
                    _handlePurchase(amount);
                  },
            child: _isProcessing
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Purchase'),
          ),
        ] else ...[
          TextButton(
            onPressed: _isProcessing
                ? null
                : () {
                    Navigator.of(context).pop();
                  },
            child: const Text('Cancel'),
          ),
        ],
      ],
    );
  }
}

class CreditPackage {
  final int credits;
  final int price;

  CreditPackage({required this.credits, required this.price});
}

