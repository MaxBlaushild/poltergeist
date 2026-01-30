import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../providers/auth_provider.dart';
import '../services/media_service.dart';
import 'logister_modal.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Consumer<AuthProvider>(
          builder: (context, auth, _) {
            if (auth.loading) {
              return const CircularProgressIndicator();
            }
            if (auth.isAuthenticated) {
              return const SizedBox.shrink(); // redirect handles navigation
            }
            return _HomeContent();
          },
        ),
      ),
    );
  }
}

class _HomeContent extends StatefulWidget {
  @override
  State<_HomeContent> createState() => _HomeContentState();
}

class _HomeContentState extends State<_HomeContent> {
  bool _showLogister = false;

  @override
  Widget build(BuildContext context) {
    final from = Uri.base.queryParameters['from'];
    final shouldShowLogister = _showLogister || (from != null && from.isNotEmpty);

    if (shouldShowLogister) {
      final mediaService = context.read<MediaService>();
      return LogisterModal(
        mediaService: mediaService,
        onSuccess: () {
          final dest = from != null && from.isNotEmpty
              ? Uri.decodeComponent(from)
              : '/single-player';
          context.go(dest);
        },
        onSkip: () {
          if (from != null && from.isNotEmpty) {
            context.go('/');
          } else {
            setState(() => _showLogister = false);
          }
        },
      );
    }

    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Text(
            'Find your crew and set your sights on adventure',
            style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 24),
          FilledButton(
            onPressed: () => setState(() => _showLogister = true),
            child: const Text('Get started'),
          ),
        ],
      ),
    );
  }
}
