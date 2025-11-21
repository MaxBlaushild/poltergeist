import 'package:flutter/material.dart';

/// Get Advice screen for requesting travel advice
class GetAdviceScreen extends StatelessWidget {
  const GetAdviceScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Text(
          'Get Advice',
          style: TextStyle(
            fontSize: 32,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
    );
  }
}
