import 'package:flutter/material.dart';
import 'package:travel_angels/screens/discover_screen.dart';
import 'package:travel_angels/screens/advice_screen.dart';
import 'package:travel_angels/screens/documents_screen.dart';
import 'package:travel_angels/screens/profile_screen.dart';
import 'package:travel_angels/screens/my_network_screen.dart';
import 'package:travel_angels/utils/platform_utils.dart';
import 'package:travel_angels/widgets/main_navbar.dart';

/// Main scaffold wrapper that provides responsive navigation for logged-in users
class MainScaffold extends StatefulWidget {
  const MainScaffold({super.key});

  @override
  State<MainScaffold> createState() => _MainScaffoldState();
}

class _MainScaffoldState extends State<MainScaffold> {
  int _currentIndex = 0;

  final List<Widget> _screens = const [
    DiscoverScreen(),
    MyNetworkScreen(),
    ProfileScreen(),
    AdviceScreen(),
    DocumentsScreen(),
  ];

  void _onDestinationChanged(int index) {
    setState(() {
      _currentIndex = index;
    });
  }

  @override
  Widget build(BuildContext context) {
    final isDesktop = PlatformUtils.shouldShowTopNav(context);

    if (isDesktop) {
      // Desktop: Top navigation bar in AppBar
      return Scaffold(
        appBar: MainNavbarAppBar(
          currentIndex: _currentIndex,
          onDestinationChanged: _onDestinationChanged,
        ),
        body: _screens[_currentIndex],
      );
    } else {
      // Mobile/Native: Bottom navigation bar
      // Note: Screens with AppBar handle their own safe areas
      // Screens without AppBar (like DiscoverScreen) need SafeArea, which they handle internally
      return Scaffold(
        body: _screens[_currentIndex],
        bottomNavigationBar: MainNavbar(
          currentIndex: _currentIndex,
          onDestinationChanged: _onDestinationChanged,
        ),
      );
    }
  }
}
