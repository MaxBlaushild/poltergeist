import 'package:flutter/material.dart';
import 'package:travel_angels/models/document.dart';

/// Utility functions for document-related operations
class DocumentUtils {
  /// Gets a human-readable label for a document provider
  static String getProviderLabel(CloudDocumentProvider provider) {
    switch (provider) {
      case CloudDocumentProvider.googleDocs:
        return 'Google Docs';
      case CloudDocumentProvider.googleSheets:
        return 'Google Sheets';
      case CloudDocumentProvider.internal:
        return 'Internal';
      case CloudDocumentProvider.unknown:
      default:
        return 'Unknown';
    }
  }

  /// Gets an icon for a document provider
  static IconData getProviderIcon(CloudDocumentProvider provider) {
    switch (provider) {
      case CloudDocumentProvider.googleDocs:
        return Icons.description;
      case CloudDocumentProvider.googleSheets:
        return Icons.table_chart;
      case CloudDocumentProvider.internal:
        return Icons.insert_drive_file;
      case CloudDocumentProvider.unknown:
      default:
        return Icons.help_outline;
    }
  }
}

