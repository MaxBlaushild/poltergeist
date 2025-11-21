import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';

/// Utility class for platform and viewport detection
class PlatformUtils {
  /// Determines if the app is running on web platform
  static bool get isWeb => kIsWeb;

  /// Determines if the current viewport is mobile-sized (< 600px width)
  /// Returns true for mobile viewports (native mobile or mobile web)
  static bool isMobileViewport(BuildContext context) {
    final width = MediaQuery.of(context).size.width;
    return width < 600;
  }

  /// Determines if the current viewport is desktop-sized (>= 600px width)
  /// Returns true for desktop web viewports
  static bool isDesktopViewport(BuildContext context) {
    return !isMobileViewport(context);
  }

  /// Determines if navigation should be displayed as bottom bar
  /// True for native mobile and mobile web viewports
  static bool shouldShowBottomNav(BuildContext context) {
    // On web, show bottom nav only for mobile viewports
    // On native, always show bottom nav
    return !isWeb || isMobileViewport(context);
  }

  /// Determines if navigation should be displayed as top header
  /// True for desktop web viewports only
  static bool shouldShowTopNav(BuildContext context) {
    return isWeb && isDesktopViewport(context);
  }
}
