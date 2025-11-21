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
  bool _hasInitialized = false;

  @override
  void initState() {
    super.initState();
    _initializeUserLevel();
  }

  void _initializeUserLevel() {
    if (_hasInitialized) return;
    
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      
      final authProvider = context.read<AuthProvider>();
      final userLevelProvider = context.read<UserLevelProvider>();
      
      if (!authProvider.isAuthenticated) {
        userLevelProvider.clear();
      } else if (!userLevelProvider.loading && userLevelProvider.userLevel == null) {
        userLevelProvider.fetchUserLevel();
      }
      
      _hasInitialized = true;
    });
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    
    final authProvider = context.watch<AuthProvider>();
    final userLevelProvider = context.read<UserLevelProvider>();
    
    // Reset initialization flag if auth state changes
    if (!authProvider.isAuthenticated && _hasInitialized) {
      _hasInitialized = false;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) {
          userLevelProvider.clear();
        }
      });
    } else if (authProvider.isAuthenticated && !userLevelProvider.loading && userLevelProvider.userLevel == null && _hasInitialized) {
      // Only fetch if we've already initialized and auth is now valid
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && !userLevelProvider.loading && userLevelProvider.userLevel == null) {
          userLevelProvider.fetchUserLevel();
        }
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final authProvider = context.watch<AuthProvider>();

    // Show loading indicator while checking auth
    if (authProvider.loading) {
      return const Scaffold(
        body: Center(
          child: CircularProgressIndicator(),
        ),
      );
    }

    // Show login screen if not authenticated
    if (!authProvider.isAuthenticated) {
      return const LoginScreen();
    }

    // Show main app if authenticated
    return const MainScaffold();
  }
}

