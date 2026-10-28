import 'package:flutter/material.dart';
import 'package:travel_angels/theme/app_theme.dart';
import 'package:travel_angels/utils/platform_utils.dart';

/// Enum for navigation destinations
enum NavDestination {
  calendar(Icons.luggage_outlined, 'Trips'),
  library(Icons.auto_stories_outlined, 'Library'),
  myCircle(Icons.people, 'My Circle'),
  profile(Icons.person, 'Profile'),
  advice(Icons.chat_bubble_outline, 'Advice');

  const NavDestination(this.icon, this.label);
  final IconData icon;
  final String label;
}

class MainNavbar extends StatelessWidget {
  final int currentIndex;
  final ValueChanged<int> onDestinationChanged;

  const MainNavbar({
    super.key,
    required this.currentIndex,
    required this.onDestinationChanged,
  });

  @override
  Widget build(BuildContext context) {
    // Show bottom navigation bar for mobile/native
    if (PlatformUtils.shouldShowBottomNav(context)) {
      return NavigationBar(
        selectedIndex: currentIndex,
        onDestinationSelected: onDestinationChanged,
        destinations: NavDestination.values
            .map(
              (dest) => NavigationDestination(
                icon: Icon(dest.icon),
                label: dest.label,
              ),
            )
            .toList(),
      );
    }

    // Show top navigation bar for desktop web
    return NavigationBar(
      selectedIndex: currentIndex,
      onDestinationSelected: onDestinationChanged,
      labelBehavior: NavigationDestinationLabelBehavior.alwaysShow,
      destinations: NavDestination.values
          .map(
            (dest) =>
                NavigationDestination(icon: Icon(dest.icon), label: dest.label),
          )
          .toList(),
    );
  }
}

/// Custom AppBar widget for desktop header navigation
class MainNavbarAppBar extends StatelessWidget implements PreferredSizeWidget {
  final int currentIndex;
  final ValueChanged<int> onDestinationChanged;

  const MainNavbarAppBar({
    super.key,
    required this.currentIndex,
    required this.onDestinationChanged,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return AppBar(
      title: const TravelAngelsWordmark(),
      titleSpacing: 24,
      bottom: PreferredSize(
        preferredSize: const Size.fromHeight(48.0),
        child: Container(
          height: 48.0,
          decoration: BoxDecoration(
            border: Border(
              bottom: BorderSide(color: AppTheme.outline, width: 1),
            ),
          ),
          child: Row(
            children: NavDestination.values.map((dest) {
              final isSelected = currentIndex == dest.index;
              return Expanded(
                child: Semantics(
                  selected: isSelected,
                  button: true,
                  child: InkWell(
                    onTap: () => onDestinationChanged(dest.index),
                    child: Container(
                      alignment: Alignment.center,
                      decoration: BoxDecoration(
                        border: Border(
                          bottom: BorderSide(
                            color: isSelected
                                ? AppTheme.rust
                                : Colors.transparent,
                            width: 2,
                          ),
                        ),
                      ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          if (MediaQuery.sizeOf(context).width >= 900) ...[
                            Icon(
                              dest.icon,
                              size: 18,
                              color: isSelected ? AppTheme.rust : AppTheme.ink,
                            ),
                            const SizedBox(width: 7),
                          ],
                          Flexible(
                            child: Text(
                              dest.label,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: theme.textTheme.labelLarge?.copyWith(
                                color: isSelected
                                    ? AppTheme.rust
                                    : AppTheme.ink,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              );
            }).toList(),
          ),
        ),
      ),
    );
  }

  @override
  Size get preferredSize => const Size.fromHeight(104.0); // AppBar height + navigation bar height
}
