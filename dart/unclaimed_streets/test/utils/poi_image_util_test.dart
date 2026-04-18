import 'package:flutter_test/flutter_test.dart';
import 'package:image/image.dart' as img;
import 'package:unclaimed_streets/utils/poi_image_util.dart';

void main() {
  test('drawMainStoryCrest centers on the opaque content bounds', () {
    final leftAligned = _markerWithContent(x1: 34, y1: 86, x2: 82, y2: 152);
    final leftBefore = leftAligned.clone();
    drawMainStoryCrest(leftAligned);
    final leftDiff = _diffBounds(leftBefore, leftAligned);

    expect(leftDiff, isNotNull);
    expect(leftDiff!.centerX, closeTo(58, 5));
    expect(leftDiff.top, lessThan(86));
    expect(leftDiff.bottom, greaterThan(86));

    final rightAligned = _markerWithContent(x1: 118, y1: 42, x2: 164, y2: 122);
    final rightBefore = rightAligned.clone();
    drawMainStoryCrest(rightAligned);
    final rightDiff = _diffBounds(rightBefore, rightAligned);

    expect(rightDiff, isNotNull);
    expect(rightDiff!.centerX, closeTo(141, 5));
    expect(rightDiff.top, lessThan(42));
    expect(rightDiff.bottom, greaterThan(42));
  });

  test(
    'drawMainStoryCrest still attaches when content starts at the top edge',
    () {
      final topAnchored = _markerWithContent(x1: 68, y1: 0, x2: 122, y2: 92);
      final before = topAnchored.clone();

      drawMainStoryCrest(topAnchored);
      final diff = _diffBounds(before, topAnchored);

      expect(diff, isNotNull);
      expect(diff!.left, greaterThanOrEqualTo(0));
      expect(diff.right, lessThan(topAnchored.width));
      expect(diff.top, greaterThanOrEqualTo(0));
      expect(diff.bottom, greaterThan(0));
    },
  );
}

img.Image _markerWithContent({
  required int x1,
  required int y1,
  required int x2,
  required int y2,
}) {
  final image = img.Image(width: 192, height: 192, numChannels: 4);
  img.fill(image, color: img.ColorRgba8(0, 0, 0, 0));
  img.fillRect(
    image,
    x1: x1,
    y1: y1,
    x2: x2,
    y2: y2,
    color: img.ColorRgba8(96, 104, 118, 255),
  );
  return image;
}

_DiffBounds? _diffBounds(img.Image before, img.Image after) {
  var left = after.width;
  var top = after.height;
  var right = -1;
  var bottom = -1;

  for (var y = 0; y < after.height; y++) {
    for (var x = 0; x < after.width; x++) {
      final beforePixel = before.getPixel(x, y);
      final afterPixel = after.getPixel(x, y);
      final changed =
          beforePixel.r != afterPixel.r ||
          beforePixel.g != afterPixel.g ||
          beforePixel.b != afterPixel.b ||
          beforePixel.a != afterPixel.a;
      if (!changed) continue;
      if (x < left) left = x;
      if (x > right) right = x;
      if (y < top) top = y;
      if (y > bottom) bottom = y;
    }
  }

  if (right < left || bottom < top) {
    return null;
  }
  return _DiffBounds(left: left, top: top, right: right, bottom: bottom);
}

class _DiffBounds {
  const _DiffBounds({
    required this.left,
    required this.top,
    required this.right,
    required this.bottom,
  });

  final int left;
  final int top;
  final int right;
  final int bottom;

  double get centerX => (left + right) / 2;
}
