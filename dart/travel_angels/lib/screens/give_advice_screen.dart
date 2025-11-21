import 'package:flutter/material.dart';

/// Give Advice screen for providing travel advice to others
class GiveAdviceScreen extends StatelessWidget {
  const GiveAdviceScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Text(
          'Give Advice',
          style: TextStyle(
            fontSize: 32,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
    );
  }
}
