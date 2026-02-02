import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:skunkworks/constants/app_colors.dart';

/// Widget that composites the post image with "Verified by Vera" badge and QR code.
/// Used for capture via RepaintBoundary.
class ExportCompositeWidget extends StatelessWidget {
  final Uint8List imageBytes;
  final String qrData;

  const ExportCompositeWidget({
    super.key,
    required this.imageBytes,
    required this.qrData,
  });

  @override
  Widget build(BuildContext context) {
    return Stack(
      fit: StackFit.expand,
      children: [
        // Background image
        Image.memory(
          imageBytes,
          fit: BoxFit.cover,
        ),
        // "Verified by Vera" badge - top right
        Positioned(
          top: 12,
          right: 12,
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: Colors.black.withValues(alpha: 0.6),
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: Colors.white.withValues(alpha: 0.3), width: 1),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(
                  Icons.verified,
                  color: AppColors.softRealBlue,
                  size: 20,
                ),
                const SizedBox(width: 6),
                Text(
                  'Verified by Vera',
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
          ),
        ),
        // QR code - bottom right
        Positioned(
          bottom: 12,
          right: 12,
          child: Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(8),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.3),
                  blurRadius: 8,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: QrImageView(
              data: qrData,
              version: QrVersions.auto,
              size: 100,
              backgroundColor: Colors.white,
              gapless: true,
            ),
          ),
        ),
      ],
    );
  }
}
