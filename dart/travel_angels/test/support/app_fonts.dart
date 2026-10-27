import 'package:flutter/services.dart';

/// Exercise real font metrics in responsive widget checks, including CI.
Future<void> loadAppFonts() async {
  for (final entry in {
    'DMSerifDisplay': [
      'assets/fonts/DMSerifDisplay-Regular.ttf',
      'assets/fonts/DMSerifDisplay-Italic.ttf',
    ],
    'Marcellus': ['assets/fonts/Marcellus-Regular.ttf'],
    'Roboto': ['assets/fonts/Roboto-Regular.ttf'],
    'MaterialIcons': ['fonts/MaterialIcons-Regular.otf'],
  }.entries) {
    final loader = FontLoader(entry.key);
    for (final asset in entry.value) {
      loader.addFont(rootBundle.load(asset));
    }
    await loader.load();
  }
}
