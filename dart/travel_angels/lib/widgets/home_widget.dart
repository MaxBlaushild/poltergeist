import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/providers/user_level_provider.dart';
import 'package:travel_angels/screens/login_screen.dart';
import 'package:travel_angels/widgets/main_scaffold.dart';

/// Home widget that manages authentication and user level fetching
class HomeWidget extends StatefulWidget {
  const HomeWidget({super.key});

  @override
  State<HomeWidget> createState() => _HomeWidgetState();
}

class _HomeWidgetState extends State<HomeWidget> {
  String? _levelUserId;
  bool _observedIdentity = false;
  int _levelRevision = 0;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final auth = context.watch<AuthProvider>();
    final userId = auth.isAuthenticated ? auth.user?.id : null;
    if (_observedIdentity && userId == _levelUserId) return;
    _observedIdentity = true;
    _levelUserId = userId;
    final revision = ++_levelRevision;
    final level = context.read<UserLevelProvider>();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted || revision != _levelRevision) return;
      level.clear();
      if (userId != null) level.fetchUserLevel();
    });
  }

  @override
  Widget build(BuildContext context) {
    final authProvider = context.watch<AuthProvider>();

    // Show loading indicator while checking auth
    if (authProvider.loading && !authProvider.isAuthenticated) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }

    // Show login screen if not authenticated
    if (!authProvider.isAuthenticated) {
      return const LoginScreen();
    }

    // Show main app if authenticated
    return MainScaffold(key: ValueKey(authProvider.user?.id));
  }
}
