import 'dart:async';

import 'package:app_links/app_links.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/travel_calendar_route.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/providers/user_level_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/auth_service.dart';
import 'package:travel_angels/services/user_level_service.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/screens/travel_calendar_screen.dart';
import 'package:travel_angels/screens/travel_preferences_screen.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/widgets/home_widget.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatefulWidget {
  const MyApp({
    super.key,
    this.initialRoute,
    this.calendarService,
    this.listenForDeepLinks = true,
  });

  final String? initialRoute;
  final TravelCalendarService? calendarService;
  final bool listenForDeepLinks;

  @override
  State<MyApp> createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> {
  final AppLinks _appLinks = AppLinks();
  StreamSubscription<Uri>? _linkSubscription;
  late final AuthProvider _authProvider;
  late final UserLevelProvider _userLevelProvider;
  late final TravelCalendarService _calendarService;
  final _navigatorKey = GlobalKey<NavigatorState>();
  String? _pendingRoute;

  @override
  void initState() {
    super.initState();
    final apiClient = APIClient(ApiConstants.baseUrl);
    _authProvider = AuthProvider(AuthService(apiClient));
    _userLevelProvider = UserLevelProvider(UserLevelService(apiClient));
    _calendarService = widget.calendarService ?? TravelCalendarService();
    apiClient.setOnAuthError(_authProvider.logout);
    if (widget.listenForDeepLinks) _initDeepLinkListener();
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
    _appLinks
        .getInitialLink()
        .then((Uri? uri) {
          if (uri != null) {
            _handleDeepLink(uri);
          }
        })
        .catchError((Object _) {});
  }

  void _handleDeepLink(Uri uri) {
    final route = TravelCalendarRoute.parse(uri.toString());
    if (route != null) {
      if (_navigatorKey.currentState == null) {
        _pendingRoute = route.path;
        WidgetsBinding.instance.addPostFrameCallback(
          (_) => _openPendingRoute(),
        );
      } else {
        _navigatorKey.currentState!.pushNamed(route.path);
      }
      return;
    }

    if (uri.scheme == 'travelangels') {
      // Handle credit purchase success
      if (uri.host == 'credits' &&
          uri.pathSegments.length >= 3 &&
          uri.pathSegments[1] == 'purchase' &&
          uri.pathSegments[2] == 'success') {
        debugPrint('Credit purchase success detected, refreshing user data');
        // Refresh user data to get updated credits
        _authProvider.verifyToken();
      }
      // Route to appropriate screen based on deep link
      // For now, OAuth callbacks are handled by PermissionsPanel widget
      // This can be extended for other deep link routes
    }
  }

  void _openPendingRoute() {
    if (!mounted || _pendingRoute == null) return;
    _navigatorKey.currentState?.pushNamed(_pendingRoute!);
    _pendingRoute = null;
  }

  Route<dynamic> _route(RouteSettings settings) {
    final shared = TravelCalendarRoute.parse(settings.name ?? '/');
    final Widget page;
    if (shared?.kind == 'preferences') {
      page = TravelPreferencesScreen(
        token: shared!.token,
        service: _calendarService,
      );
    } else if (shared != null) {
      page = TravelCalendarScreen(
        calendarToken: shared.kind == 'calendar' ? shared.token : null,
        stopToken: shared.kind == 'stops' ? shared.token : null,
        service: _calendarService,
      );
    } else if (settings.name == '/' || settings.name == null) {
      page = const HomeWidget();
    } else {
      page = const Scaffold(
        body: Center(
          child: Text(
            'This link is unavailable. Ask the host for a current sharing link.',
          ),
        ),
      );
    }
    return MaterialPageRoute<void>(settings: settings, builder: (_) => page);
  }

  @override
  void dispose() {
    _linkSubscription?.cancel();
    _authProvider.dispose();
    _userLevelProvider.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider.value(value: _authProvider),
        ChangeNotifierProvider.value(value: _userLevelProvider),
        Provider.value(value: _calendarService),
      ],
      child: MaterialApp(
        title: 'Travel Angels',
        theme: AppTheme.lightTheme,
        navigatorKey: _navigatorKey,
        initialRoute:
            widget.initialRoute ??
            TravelCalendarRoute.parse(Uri.base.toString())?.path ??
            '/',
        onGenerateRoute: _route,
        onGenerateInitialRoutes: (initialRoute) => [
          _route(RouteSettings(name: initialRoute)),
        ],
      ),
    );
  }
}
