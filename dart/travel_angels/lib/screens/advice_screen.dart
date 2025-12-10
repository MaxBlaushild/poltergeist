import 'package:flutter/material.dart';
import 'package:travel_angels/widgets/quick_decision_bottom_sheet.dart';

/// Consolidated Advice screen that combines Get Advice and Give Advice functionality
class AdviceScreen extends StatefulWidget {
  const AdviceScreen({super.key});

  @override
  State<AdviceScreen> createState() => _AdviceScreenState();
}

class _AdviceScreenState extends State<AdviceScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    
    return Scaffold(
      body: SafeArea(
        child: Column(
          children: [
            // Advice header
            Padding(
              padding: const EdgeInsets.fromLTRB(16.0, 16.0, 16.0, 8.0),
              child: Align(
                alignment: Alignment.centerLeft,
                child: Text(
                  'Advice',
                  style: theme.textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ),
            // Enhanced TabBar with more prominent styling
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8.0),
              child: TabBar(
                controller: _tabController,
                labelColor: theme.colorScheme.primary,
                unselectedLabelColor: theme.colorScheme.onSurface.withOpacity(0.6),
                indicatorColor: theme.colorScheme.primary,
                indicatorWeight: 3.0,
                indicatorSize: TabBarIndicatorSize.tab,
                labelStyle: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                  fontSize: 16,
                ),
                unselectedLabelStyle: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.normal,
                  fontSize: 16,
                ),
                labelPadding: const EdgeInsets.symmetric(vertical: 12.0, horizontal: 8.0),
                tabs: const [
                  Tab(
                    icon: Icon(Icons.help_outline, size: 28),
                    text: 'Get Advice',
                    iconMargin: EdgeInsets.only(bottom: 4.0),
                  ),
                  Tab(
                    icon: Icon(Icons.lightbulb_outline, size: 28),
                    text: 'Give Advice',
                    iconMargin: EdgeInsets.only(bottom: 4.0),
                  ),
                ],
              ),
            ),
            Expanded(
              child: TabBarView(
                controller: _tabController,
                children: const [
                  _GetAdviceTab(),
                  _GiveAdviceTab(),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Get Advice tab content
class _GetAdviceTab extends StatefulWidget {
  const _GetAdviceTab();

  @override
  State<_GetAdviceTab> createState() => _GetAdviceTabState();
}

class _GetAdviceTabState extends State<_GetAdviceTab> {
  void _showQuickDecisionBottomSheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) => const QuickDecisionBottomSheet(),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Title
          Text(
            'What do you need help with',
            style: theme.textTheme.headlineSmall?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 24),
          // Quick decision option
          InkWell(
            onTap: _showQuickDecisionBottomSheet,
            borderRadius: BorderRadius.circular(12),
            child: Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Row(
                  children: [
                    Icon(
                      Icons.speed,
                      size: 32,
                      color: theme.colorScheme.primary,
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Quick decision',
                            style: theme.textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            'choose between 2-3 options',
                            style: theme.textTheme.bodyMedium?.copyWith(
                              color: theme.colorScheme.onSurface.withOpacity(0.7),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
          const SizedBox(height: 12),
          // Community Poll option
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Row(
                children: [
                  Icon(
                    Icons.poll,
                    size: 32,
                    color: theme.colorScheme.primary,
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Community Poll',
                          style: theme.textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          'Get ranked input from 10+ travelers like you',
                          style: theme.textTheme.bodyMedium?.copyWith(
                            color: theme.colorScheme.onSurface.withOpacity(0.7),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 12),
          // Chat Consultation option
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Row(
                children: [
                  Icon(
                    Icons.chat_bubble_outline,
                    size: 32,
                    color: theme.colorScheme.primary,
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Chat Consultation',
                          style: theme.textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          '5-10 min 1-on-1 advice',
                          style: theme.textTheme.bodyMedium?.copyWith(
                            color: theme.colorScheme.onSurface.withOpacity(0.7),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// Give Advice tab content
class _GiveAdviceTab extends StatelessWidget {
  const _GiveAdviceTab();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Text(
        'Give Advice',
        style: TextStyle(
          fontSize: 32,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }
}

