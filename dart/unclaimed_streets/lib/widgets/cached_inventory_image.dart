import 'dart:async';
import 'dart:io';

import 'package:flutter/material.dart';

import '../services/inventory_image_cache.dart';

class CachedInventoryImage extends StatefulWidget {
  const CachedInventoryImage({
    super.key,
    required this.imageUrl,
    required this.fit,
    this.cacheWidth,
    this.errorBuilder,
  });

  final String imageUrl;
  final BoxFit fit;
  final int? cacheWidth;
  final Widget Function(BuildContext context)? errorBuilder;

  @override
  State<CachedInventoryImage> createState() => _CachedInventoryImageState();
}

class _CachedInventoryImageState extends State<CachedInventoryImage> {
  File? _cachedFile;
  String _activeUrl = '';

  @override
  void initState() {
    super.initState();
    _resolveImage();
  }

  @override
  void didUpdateWidget(covariant CachedInventoryImage oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.imageUrl != widget.imageUrl) {
      _resolveImage();
    }
  }

  Future<void> _resolveImage() async {
    final imageUrl = widget.imageUrl.trim();
    _activeUrl = imageUrl;
    if (imageUrl.isEmpty) {
      if (mounted) {
        setState(() {
          _cachedFile = null;
        });
      }
      return;
    }

    final cachedFile = await InventoryImageCache.instance.getCachedFile(
      imageUrl,
    );
    if (!mounted || _activeUrl != imageUrl) {
      return;
    }
    if (cachedFile != null) {
      setState(() {
        _cachedFile = cachedFile;
      });
      return;
    }

    setState(() {
      _cachedFile = null;
    });
    unawaited(
      InventoryImageCache.instance.warmImage(imageUrl).then((storedFile) {
        if (!mounted || _activeUrl != imageUrl || storedFile == null) {
          return;
        }
        setState(() {
          _cachedFile = storedFile;
        });
      }),
    );
  }

  @override
  Widget build(BuildContext context) {
    final cachedFile = _cachedFile;
    if (cachedFile != null) {
      return Image.file(
        cachedFile,
        fit: widget.fit,
        cacheWidth: widget.cacheWidth,
        errorBuilder: (context, error, stackTrace) => _buildError(context),
      );
    }
    final imageUrl = widget.imageUrl.trim();
    if (imageUrl.isEmpty) {
      return _buildError(context);
    }
    return Image.network(
      imageUrl,
      fit: widget.fit,
      cacheWidth: widget.cacheWidth,
      errorBuilder: (context, error, stackTrace) => _buildError(context),
    );
  }

  Widget _buildError(BuildContext context) {
    final builder = widget.errorBuilder;
    if (builder != null) {
      return builder(context);
    }
    return const SizedBox.shrink();
  }
}
