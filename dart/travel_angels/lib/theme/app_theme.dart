import 'package:flutter/material.dart';

/// Shared palette and bundled typography for every Travel Angels surface.
class AppTheme {
  static const ink = Color(0xFF1C4E69);
  static const peach = Color(0xFFFFDBB3);
  static const cream = Color(0xFFFCDAB6);
  static const orange = Color(0xFFD34A17);
  static const paper = Color(0xFFFFF7ED);
  static const rust = Color(0xFFA53B18);
  static const muted = Color(0xFF4A6070);
  static const outline = Color(0xFFB9B0A1);
  static const success = Color(0xFF2F6655);
  static const error = Color(0xFF94382B);
  static const displayFont = 'DMSerifDisplay';
  static const bodyFont = 'Marcellus';
  static const _shape = RoundedRectangleBorder(
    borderRadius: BorderRadius.all(Radius.circular(4)),
  );

  static TextStyle _display(double size, {double height = 1.15}) => TextStyle(
    fontFamily: displayFont,
    fontSize: size,
    fontWeight: FontWeight.w400,
    height: height,
    color: ink,
  );

  static ThemeData get lightTheme {
    final scheme =
        ColorScheme.fromSeed(
          seedColor: ink,
          brightness: Brightness.light,
        ).copyWith(
          primary: ink,
          onPrimary: cream,
          primaryContainer: peach,
          onPrimaryContainer: ink,
          secondary: rust,
          onSecondary: paper,
          secondaryContainer: cream,
          onSecondaryContainer: ink,
          tertiary: rust,
          onTertiary: paper,
          tertiaryContainer: peach,
          onTertiaryContainer: ink,
          surface: paper,
          onSurface: ink,
          onSurfaceVariant: muted,
          surfaceContainerLowest: paper,
          surfaceContainerLow: const Color(0xFFFFEDDA),
          surfaceContainer: peach,
          surfaceContainerHigh: cream,
          surfaceContainerHighest: const Color(0xFFEBD0AF),
          surfaceTint: Colors.transparent,
          outline: const Color(0xFF6A7C85),
          outlineVariant: outline,
          error: error,
          onError: paper,
          errorContainer: const Color(0xFFF8DDD6),
          onErrorContainer: const Color(0xFF67261D),
          inverseSurface: ink,
          onInverseSurface: cream,
          inversePrimary: peach,
        );
    final text = ThemeData.light().textTheme
        .apply(fontFamily: bodyFont, bodyColor: ink, displayColor: ink)
        .copyWith(
          displayLarge: _display(56, height: 1.05),
          displayMedium: _display(46, height: 1.08),
          displaySmall: _display(38),
          headlineLarge: _display(34),
          headlineMedium: _display(30),
          headlineSmall: _display(26),
          titleLarge: _display(23, height: 1.25),
          titleMedium: const TextStyle(
            fontFamily: bodyFont,
            fontSize: 17,
            height: 1.4,
            color: ink,
          ),
          titleSmall: const TextStyle(
            fontFamily: bodyFont,
            fontSize: 15,
            height: 1.4,
            color: ink,
          ),
          bodyLarge: const TextStyle(
            fontFamily: bodyFont,
            fontSize: 16,
            height: 1.5,
            color: ink,
          ),
          bodyMedium: const TextStyle(
            fontFamily: bodyFont,
            fontSize: 14,
            height: 1.5,
            color: ink,
          ),
          bodySmall: const TextStyle(
            fontFamily: bodyFont,
            fontSize: 12,
            height: 1.5,
            color: muted,
          ),
          labelLarge: const TextStyle(
            fontFamily: bodyFont,
            fontSize: 14,
            fontWeight: FontWeight.w400,
            height: 1.3,
            color: ink,
          ),
          labelMedium: const TextStyle(
            fontFamily: bodyFont,
            fontSize: 12,
            fontWeight: FontWeight.w400,
            height: 1.3,
            color: ink,
          ),
          labelSmall: const TextStyle(
            fontFamily: bodyFont,
            fontSize: 11,
            fontWeight: FontWeight.w400,
            height: 1.3,
            color: muted,
          ),
        );
    final primaryButton = ButtonStyle(
      backgroundColor: WidgetStateProperty.resolveWith(
        (states) => states.contains(WidgetState.disabled)
            ? ink.withValues(alpha: .12)
            : ink,
      ),
      foregroundColor: WidgetStateProperty.resolveWith(
        (states) => states.contains(WidgetState.disabled)
            ? ink.withValues(alpha: .45)
            : cream,
      ),
      elevation: const WidgetStatePropertyAll(0),
      shape: const WidgetStatePropertyAll(_shape),
      minimumSize: const WidgetStatePropertyAll(Size(44, 48)),
      padding: const WidgetStatePropertyAll(
        EdgeInsets.symmetric(horizontal: 20, vertical: 13),
      ),
      textStyle: WidgetStatePropertyAll(text.labelLarge),
    );
    return ThemeData(
      useMaterial3: true,
      colorScheme: scheme,
      fontFamily: bodyFont,
      textTheme: text,
      scaffoldBackgroundColor: paper,
      canvasColor: paper,
      dividerColor: outline,
      disabledColor: muted.withValues(alpha: .45),
      appBarTheme: AppBarTheme(
        backgroundColor: peach,
        foregroundColor: ink,
        surfaceTintColor: Colors.transparent,
        elevation: 0,
        scrolledUnderElevation: 0,
        centerTitle: false,
        titleTextStyle: _display(26),
        iconTheme: const IconThemeData(color: ink),
        shape: const Border(bottom: BorderSide(color: outline)),
      ),
      cardTheme: CardThemeData(
        color: paper,
        surfaceTintColor: Colors.transparent,
        elevation: 0,
        margin: const EdgeInsets.symmetric(vertical: 8),
        shape: _shape.copyWith(side: const BorderSide(color: outline)),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(style: primaryButton),
      filledButtonTheme: FilledButtonThemeData(style: primaryButton),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: ink,
          side: const BorderSide(color: ink),
          shape: _shape,
          minimumSize: const Size(44, 48),
          padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 12),
          textStyle: text.labelLarge,
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: ink,
          shape: _shape,
          minimumSize: const Size(44, 44),
          textStyle: text.labelLarge,
        ),
      ),
      floatingActionButtonTheme: const FloatingActionButtonThemeData(
        backgroundColor: ink,
        foregroundColor: cream,
        elevation: 0,
        shape: _shape,
      ),
      inputDecorationTheme: InputDecorationThemeData(
        filled: true,
        fillColor: paper,
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 14,
          vertical: 16,
        ),
        border: const OutlineInputBorder(
          borderRadius: BorderRadius.all(Radius.circular(4)),
          borderSide: BorderSide(color: ink),
        ),
        enabledBorder: const OutlineInputBorder(
          borderRadius: BorderRadius.all(Radius.circular(4)),
          borderSide: BorderSide(color: Color(0xFF6A7C85)),
        ),
        focusedBorder: const OutlineInputBorder(
          borderRadius: BorderRadius.all(Radius.circular(4)),
          borderSide: BorderSide(color: ink, width: 2),
        ),
        labelStyle: text.bodyMedium,
        hintStyle: text.bodyMedium?.copyWith(color: muted),
        errorMaxLines: 3,
      ),
      dialogTheme: DialogThemeData(
        backgroundColor: paper,
        surfaceTintColor: Colors.transparent,
        shape: _shape,
        titleTextStyle: text.headlineSmall,
        contentTextStyle: text.bodyMedium,
      ),
      bottomSheetTheme: const BottomSheetThemeData(
        backgroundColor: paper,
        modalBackgroundColor: paper,
        surfaceTintColor: Colors.transparent,
        dragHandleColor: muted,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(8)),
        ),
      ),
      navigationBarTheme: NavigationBarThemeData(
        backgroundColor: peach,
        surfaceTintColor: Colors.transparent,
        indicatorColor: ink,
        indicatorShape: _shape,
        elevation: 0,
        height: 76,
        labelTextStyle: WidgetStatePropertyAll(
          text.labelSmall?.copyWith(color: ink, fontSize: 11),
        ),
        iconTheme: WidgetStateProperty.resolveWith(
          (states) => IconThemeData(
            color: states.contains(WidgetState.selected) ? cream : ink,
            size: 22,
          ),
        ),
      ),
      tabBarTheme: TabBarThemeData(
        labelColor: rust,
        unselectedLabelColor: ink,
        labelStyle: text.labelLarge,
        unselectedLabelStyle: text.labelLarge,
        indicatorColor: rust,
        dividerColor: outline,
      ),
      segmentedButtonTheme: SegmentedButtonThemeData(
        style: ButtonStyle(
          shape: const WidgetStatePropertyAll(_shape),
          side: const WidgetStatePropertyAll(BorderSide(color: ink)),
          foregroundColor: const WidgetStatePropertyAll(ink),
          backgroundColor: WidgetStateProperty.resolveWith(
            (states) => states.contains(WidgetState.selected) ? peach : paper,
          ),
          textStyle: WidgetStatePropertyAll(text.labelLarge),
          minimumSize: const WidgetStatePropertyAll(Size(44, 44)),
        ),
      ),
      chipTheme: ChipThemeData(
        backgroundColor: peach,
        selectedColor: peach,
        secondarySelectedColor: peach,
        checkmarkColor: ink,
        labelStyle: text.labelMedium,
        secondaryLabelStyle: text.labelMedium,
        shape: _shape,
        side: const BorderSide(color: outline),
        padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 5),
      ),
      dividerTheme: const DividerThemeData(color: outline, thickness: 1),
      snackBarTheme: SnackBarThemeData(
        backgroundColor: ink,
        contentTextStyle: text.bodyMedium?.copyWith(color: cream),
        actionTextColor: cream,
        shape: _shape,
        behavior: SnackBarBehavior.floating,
      ),
      popupMenuTheme: PopupMenuThemeData(
        color: paper,
        surfaceTintColor: Colors.transparent,
        shape: _shape,
        textStyle: text.bodyMedium,
      ),
      tooltipTheme: TooltipThemeData(
        decoration: const BoxDecoration(
          color: ink,
          borderRadius: BorderRadius.all(Radius.circular(4)),
        ),
        textStyle: text.bodySmall?.copyWith(color: cream),
      ),
      dataTableTheme: DataTableThemeData(
        headingRowColor: const WidgetStatePropertyAll(peach),
        headingTextStyle: text.titleSmall,
        dataTextStyle: text.bodyMedium,
        dividerThickness: 1,
      ),
      progressIndicatorTheme: const ProgressIndicatorThemeData(color: ink),
    );
  }
}

class TravelAngelsWordmark extends StatelessWidget {
  const TravelAngelsWordmark({super.key, this.fontSize = 28});
  final double fontSize;

  @override
  Widget build(BuildContext context) => Text.rich(
    const TextSpan(
      children: [
        TextSpan(text: 'travel '),
        TextSpan(
          text: 'angels',
          style: TextStyle(fontStyle: FontStyle.italic),
        ),
      ],
    ),
    semanticsLabel: 'Travel Angels',
    style: TextStyle(
      fontFamily: AppTheme.displayFont,
      fontSize: fontSize,
      fontWeight: FontWeight.w400,
      color: AppTheme.ink,
    ),
  );
}
