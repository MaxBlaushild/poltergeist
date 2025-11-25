import 'dart:async';

import 'package:app_links/app_links.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/providers/user_level_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/user_level_service.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/home_widget.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatefulWidget {
  const MyApp({super.key});

  @override
  State<MyApp> createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> {
  final AppLinks _appLinks = AppLinks();
  StreamSubscription<Uri>? _linkSubscription;
  AuthProvider? _authProvider;

  @override
  void initState() {
    super.initState();
    _initDeepLinkListener();
  }

  void _initDeepLinkListener() {
    // Listen for deep links globally
    _linkSubscription = _appLinks.uriLinkStream.listen(
      (Uri uri) {
        _handleDeepLink(uri);
      },
      onError: (err) {
        debugPrint('Deep link error: $err');
      },
    );

    // Check for initial link (if app was opened via deep link)
    _appLinks.getInitialLink().then((Uri? uri) {
      if (uri != null) {
        _handleDeepLink(uri);
      }
    });
  }

  void _handleDeepLink(Uri uri) {
    // Handle deep link routing
    // OAuth callbacks are handled by PermissionsPanel, but we can add
    // global routing logic here for other deep links in the future
    debugPrint('Deep link received: $uri');
    
    if (uri.scheme == 'travelangels') {
      // Handle credit purchase success
      if (uri.host == 'credits' && uri.pathSegments.length >= 3 && uri.pathSegments[1] == 'purchase' && uri.pathSegments[2] == 'success') {
        debugPrint('Credit purchase success detected, refreshing user data');
        // Refresh user data to get updated credits
        _authProvider?.verifyToken();
      }
      // Route to appropriate screen based on deep link
      // For now, OAuth callbacks are handled by PermissionsPanel widget
      // This can be extended for other deep link routes
    }
  }

  @override
  void dispose() {
    _linkSubscription?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    // Initialize services
    final apiClient = APIClient(ApiConstants.baseUrl);
    final authService = AuthService(apiClient);
    final authProvider = AuthProvider(authService);
    final userLevelService = UserLevelService(apiClient);
    final userLevelProvider = UserLevelProvider(userLevelService);

    // Store authProvider reference for deep link handling
    _authProvider = authProvider;

    // Set up auth error callback to log out user when 401/403 occurs
    apiClient.setOnAuthError(() {
      authProvider.logout();
    });

    return MultiProvider(
      providers: [
        ChangeNotifierProvider.value(value: authProvider),
        ChangeNotifierProvider.value(value: userLevelProvider),
      ],
      child: MaterialApp(
        title: 'Travel Angels',
        theme: AppTheme.lightTheme,
        home: const HomeWidget(),
      ),
    );
  }
}
