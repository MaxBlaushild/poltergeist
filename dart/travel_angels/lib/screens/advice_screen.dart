import 'package:flutter/material.dart';

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
    return Scaffold(
      body: Column(
        children: [
          TabBar(
            controller: _tabController,
            tabs: const [
              Tab(
                icon: Icon(Icons.help_outline),
                text: 'Get Advice',
              ),
              Tab(
                icon: Icon(Icons.lightbulb_outline),
                text: 'Give Advice',
              ),
            ],
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
    );
  }
}

/// Get Advice tab content
class _GetAdviceTab extends StatelessWidget {
  const _GetAdviceTab();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Text(
        'Get Advice',
        style: TextStyle(
          fontSize: 32,
          fontWeight: FontWeight.bold,
        ),
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

