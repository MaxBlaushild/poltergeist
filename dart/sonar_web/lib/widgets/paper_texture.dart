import 'dart:math' as math;

import 'package:flutter/material.dart';

class PaperTexture extends StatelessWidget {
  const PaperTexture({
    super.key,
    required this.child,
    this.opacity = 0.08,
    this.borderRadius,
    this.seed = 90210,
  });

  final Widget child;
  final double opacity;
  final BorderRadius? borderRadius;
  final int seed;

  @override
  Widget build(BuildContext context) {
    final content = Stack(
      children: [
        child,
        Positioned.fill(
          child: IgnorePointer(
            child: CustomPaint(
              painter: _PaperTexturePainter(
                opacity: opacity,
                seed: seed,
              ),
            ),
          ),
        ),
      ],
    );

    if (borderRadius != null) {
      return ClipRRect(
        borderRadius: borderRadius!,
        child: content,
      );
    }

    return content;
  }
}

class PaperSheet extends StatelessWidget {
  const PaperSheet({
    super.key,
    required this.child,
    this.borderRadius = const BorderRadius.vertical(top: Radius.circular(16)),
    this.opacity = 0.08,
    this.color,
  });

  final Widget child;
  final BorderRadius borderRadius;
  final double opacity;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    return PaperTexture(
      opacity: opacity,
      borderRadius: borderRadius,
      child: Container(
        decoration: BoxDecoration(
          color: color ?? Theme.of(context).colorScheme.surface,
          borderRadius: borderRadius,
        ),
        child: child,
      ),
    );
  }
}

class _PaperTexturePainter extends CustomPainter {
  _PaperTexturePainter({
    required this.opacity,
    required this.seed,
  });

  final double opacity;
  final int seed;

  @override
  void paint(Canvas canvas, Size size) {
    if (size.width <= 0 || size.height <= 0) return;
    final rng = math.Random(seed);
    final count = (size.width * size.height / 900).clamp(220, 1400).round();
    final dark = Paint()
      ..color = const Color(0xFF3B2F1C).withOpacity(opacity * 0.8);
    final light = Paint()
      ..color = const Color(0xFFFFF7E8).withOpacity(opacity * 0.6);

    for (var i = 0; i < count; i++) {
      final dx = rng.nextDouble() * size.width;
      final dy = rng.nextDouble() * size.height;
      final radius = rng.nextDouble() * 0.6 + 0.2;
      canvas.drawCircle(
        Offset(dx, dy),
        radius,
        rng.nextBool() ? dark : light,
      );
    }
  }

  @override
  bool shouldRepaint(covariant _PaperTexturePainter oldDelegate) {
    return oldDelegate.opacity != opacity || oldDelegate.seed != seed;
  }
}
