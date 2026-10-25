import 'package:flutter/material.dart';

/// Give Advice screen for providing travel advice to others
class GiveAdviceScreen extends StatelessWidget {
  const GiveAdviceScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Text(
          'Give Advice',
          style: Theme.of(context).textTheme.headlineLarge,
        ),
      ),
    );
  }
}
