import 'dart:typed_data';
import 'dart:ui' as ui;
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:image/image.dart' as img;
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/widgets/export_composite_widget.dart';

/// Parameters for exporting a post with authentication stamp and QR code.
class ExportParams {
  final Uint8List imageBytes;
  final String postId;
  final String? manifestHash;
  final String? txHash;

  const ExportParams({
    required this.imageBytes,
    required this.postId,
    this.manifestHash,
    this.txHash,
  });
}

/// Service for exporting posts with "Verified by Vera" stamp and QR code.
class ExportService {
  /// Exports the post image with authentication label and QR code overlay.
  /// Shows a capture page briefly, then returns JPEG bytes.
  /// Requires [context] for navigation and rendering.
  static Future<Uint8List?> exportPostWithStamp(
    BuildContext context,
    ExportParams params,
  ) async {
    final qrData = ApiConstants.exportPostDeepLink(
      params.postId,
      manifestHash: params.manifestHash,
      txHash: params.txHash,
    );

    final result = await Navigator.of(context).push<Uint8List>(
      MaterialPageRoute(
        builder: (context) => _ExportCapturePage(
          imageBytes: params.imageBytes,
          qrData: qrData,
        ),
        fullscreenDialog: true,
      ),
    );

    return result;
  }
}

/// Full-screen page that renders the composite and captures it.
/// Pops with the JPEG bytes on first frame.
class _ExportCapturePage extends StatefulWidget {
  final Uint8List imageBytes;
  final String qrData;

  const _ExportCapturePage({
    required this.imageBytes,
    required this.qrData,
  });

  @override
  State<_ExportCapturePage> createState() => _ExportCapturePageState();
}

class _ExportCapturePageState extends State<_ExportCapturePage> {
  final GlobalKey _repaintKey = GlobalKey();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _captureAndPop());
  }

  Future<void> _captureAndPop() async {
    final boundary = _repaintKey.currentContext?.findRenderObject()
        as RenderRepaintBoundary?;
    if (boundary == null || !mounted) {
      if (mounted) Navigator.of(context).pop();
      return;
    }

    try {
      final pixelRatio = 2.0;
      final image = await boundary.toImage(pixelRatio: pixelRatio);
      final byteData = await image.toByteData(format: ui.ImageByteFormat.png);
      image.dispose();

      if (byteData == null || !mounted) {
        if (mounted) Navigator.of(context).pop();
        return;
      }

      final pngBytes = byteData.buffer.asUint8List();
      final decoded = img.decodeImage(pngBytes);
      if (decoded == null || !mounted) {
        if (mounted) Navigator.of(context).pop();
        return;
      }

      final jpegBytes = img.encodeJpg(decoded, quality: 90);
      if (!mounted) {
        if (mounted) Navigator.of(context).pop();
        return;
      }

      if (mounted) {
        Navigator.of(context).pop(Uint8List.fromList(jpegBytes));
      }
    } catch (e) {
      if (mounted) {
        Navigator.of(context).pop();
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      body: RepaintBoundary(
        key: _repaintKey,
        child: SizedBox.expand(
          child: ExportCompositeWidget(
            imageBytes: widget.imageBytes,
            qrData: widget.qrData,
          ),
        ),
      ),
    );
  }
}
