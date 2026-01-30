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

  /// Checks if a document is a video based on its link extension or path
  static bool isVideo(Document document) {
    final link = document.link;
    print('[DocumentUtils.isVideo] Checking document: id=${document.id}, title="${document.title}", link=$link');
    if (link == null) {
      print('[DocumentUtils.isVideo] Link is null, not a video');
      return false;
    }
    final lowerLink = link.toLowerCase();
    // Check for video extensions in the URL path
    final uri = Uri.tryParse(lowerLink);
    if (uri != null) {
      final path = uri.path.toLowerCase();
      print('[DocumentUtils.isVideo] Parsed URI path: $path');
      if (path.endsWith('.mp4') ||
          path.endsWith('.mov') ||
          path.endsWith('.m4v') ||
          path.contains('/videos/') ||
          path.contains('.mp4') ||
          path.contains('.mov') ||
          path.contains('.m4v')) {
        print('[DocumentUtils.isVideo] ✅ Detected as VIDEO (path match)');
        return true;
      }
    }
    // Fallback: check the raw link string
    final isVideoResult = lowerLink.endsWith('.mp4') ||
        lowerLink.endsWith('.mov') ||
        lowerLink.endsWith('.m4v') ||
        lowerLink.contains('.mp4') ||
        lowerLink.contains('.mov') ||
        lowerLink.contains('.m4v') ||
        lowerLink.contains('/videos/');
    if (isVideoResult) {
      print('[DocumentUtils.isVideo] ✅ Detected as VIDEO (string match)');
    } else {
      print('[DocumentUtils.isVideo] ❌ Not a video');
    }
    return isVideoResult;
  }
}

