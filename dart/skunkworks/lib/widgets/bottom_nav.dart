import 'package:flutter/material.dart';

enum NavTab { home, search, upload, profile }

class BottomNav extends StatelessWidget {
  final NavTab currentTab;
  final Function(NavTab) onTabChanged;

  const BottomNav({
    super.key,
    required this.currentTab,
    required this.onTabChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border(
          top: BorderSide(color: Colors.grey.shade300, width: 0.5),
        ),
      ),
      child: SafeArea(
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceAround,
          children: [
            _buildNavItem(Icons.home, NavTab.home, context),
            _buildNavItem(Icons.search, NavTab.search, context),
            _buildNavItem(Icons.add_box_outlined, NavTab.upload, context),
            _buildNavItem(Icons.person_outline, NavTab.profile, context),
          ],
        ),
      ),
    );
  }

  Widget _buildNavItem(IconData icon, NavTab tab, BuildContext context) {
    final isSelected = currentTab == tab;
    return InkWell(
      onTap: () => onTabChanged(tab),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 16),
        child: Icon(
          icon,
          color: isSelected ? Colors.black : Colors.grey.shade600,
          size: 28,
        ),
      ),
    );
  }
}

