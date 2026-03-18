import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../providers/auth_provider.dart';
import '../screens/create_poi_screen.dart';
import '../screens/base_management_screen.dart';
import '../screens/home_screen.dart';
import '../screens/layout_shell.dart';
import '../screens/logout_screen.dart';
import '../screens/single_player_screen.dart';
import '../screens/user_character_screen.dart';

const _tokenKey = 'token';

final GlobalKey<NavigatorState> rootNavigatorKey = GlobalKey<NavigatorState>();
final GlobalKey<NavigatorState> shellNavigatorKey = GlobalKey<NavigatorState>();

Future<bool> _hasToken() async {
  final prefs = await SharedPreferences.getInstance();
  final t = prefs.getString(_tokenKey);
  return t != null && t.isNotEmpty;
}

bool _isProtected(String path) {
  if (path == '/' || path.startsWith('/logout')) return false;
  return true;
}

GoRouter createRouter({
  required AuthProvider authProvider,
  Listenable? refreshListenable,
}) {
  return GoRouter(
    navigatorKey: rootNavigatorKey,
    initialLocation: '/',
    refreshListenable: refreshListenable,
    redirect: (context, state) async {
      final path = state.uri.path;
      if (path == '/logout') return null;

      final hasToken = await _hasToken();
      final canAccessProtectedRoutes = authProvider.isAuthenticated;
      final isRegistrationFlowActive = authProvider.isRegistrationFlowActive;

      if (path == '/' &&
          hasToken &&
          canAccessProtectedRoutes &&
          !isRegistrationFlowActive) {
        return '/single-player';
      }
      if (_isProtected(path) &&
          (!hasToken ||
              !canAccessProtectedRoutes ||
              isRegistrationFlowActive)) {
        final from = state.uri.path;
        return from == '/' ? '/' : '/?from=${Uri.encodeComponent(from)}';
      }
      return null;
    },
    routes: [
      GoRoute(
        path: '/logout',
        builder: (context, state) => const LogoutScreen(),
      ),
      ShellRoute(
        navigatorKey: shellNavigatorKey,
        builder: (context, state, child) =>
            LayoutShell(routeUri: state.uri, child: child),
        routes: [
          GoRoute(
            path: '/',
            builder: (context, state) => HomeScreen(
              key: ValueKey(state.uri.toString()),
              from: state.uri.queryParameters['from'],
            ),
          ),
          GoRoute(
            path: '/single-player',
            pageBuilder: (_, state) =>
                const NoTransitionPage(child: SinglePlayerScreen()),
          ),
          GoRoute(
            path: '/create-point-of-interest',
            pageBuilder: (_, state) =>
                const NoTransitionPage(child: CreatePoiScreen()),
          ),
          GoRoute(
            path: '/base-management/:id',
            pageBuilder: (_, state) => NoTransitionPage(
              child: BaseManagementScreen(
                baseId: state.pathParameters['id'] ?? '',
              ),
            ),
          ),
          GoRoute(
            path: '/character/:id',
            pageBuilder: (_, state) => NoTransitionPage(
              child: UserCharacterScreen(
                userId: state.pathParameters['id'] ?? '',
              ),
            ),
          ),
        ],
      ),
    ],
  );
}
