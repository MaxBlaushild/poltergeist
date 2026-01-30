import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/audit_item.dart';
import '../providers/log_provider.dart';

class LogPanel extends StatelessWidget {
  const LogPanel({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<LogProvider>(
      builder: (context, log, _) {
        if (log.loading && log.items.isEmpty) {
          return const Center(child: CircularProgressIndicator());
        }
        if (log.items.isEmpty) {
          return const Padding(
            padding: EdgeInsets.all(24),
            child: Center(child: Text('No log entries yet')),
          );
        }
        return ListView.builder(
          shrinkWrap: true,
          itemCount: log.items.length,
          itemBuilder: (_, i) {
            final a = log.items[i];
            return ListTile(
              title: Text(a.message),
              subtitle: Text(a.createdAt),
            );
          },
        );
      },
    );
  }
}
