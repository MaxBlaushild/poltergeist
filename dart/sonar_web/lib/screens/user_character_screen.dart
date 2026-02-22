import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/user_character_profile.dart';
import '../services/user_character_service.dart';
import '../widgets/character_tab_content.dart';

class UserCharacterScreen extends StatefulWidget {
  const UserCharacterScreen({super.key, required this.userId});

  final String userId;

  @override
  State<UserCharacterScreen> createState() => _UserCharacterScreenState();
}

class _UserCharacterScreenState extends State<UserCharacterScreen> {
  late Future<UserCharacterProfile?> _profileFuture;

  @override
  void initState() {
    super.initState();
    _profileFuture = _loadProfile();
  }

  @override
  void didUpdateWidget(UserCharacterScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.userId != widget.userId) {
      _profileFuture = _loadProfile();
    }
  }

  Future<UserCharacterProfile?> _loadProfile() async {
    final svc = context.read<UserCharacterService>();
    return svc.getProfile(widget.userId);
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return FutureBuilder<UserCharacterProfile?>(
      future: _profileFuture,
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const Center(child: CircularProgressIndicator());
        }
        final profile = snapshot.data;
        if (profile == null) {
          return Padding(
            padding: const EdgeInsets.all(24),
            child: Center(
              child: Text(
                'Unable to load character.',
                style: theme.textTheme.bodyMedium,
              ),
            ),
          );
        }

        return CharacterTabContent(
          userOverride: profile.user,
          statsOverride: profile.stats,
          userLevelOverride: profile.userLevel,
          readOnly: true,
        );
      },
    );
  }
}
