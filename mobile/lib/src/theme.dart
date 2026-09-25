import 'package:flutter/material.dart';

abstract final class SfsColors {
  static const Color background = Color(0xFF0B1220);
  static const Color card = Color(0xFF121A2B);
  static const Color border = Color(0x14FFFFFF);
  static const Color surfaceOverlay = Color(0x0AFFFFFF);
  static const Color text = Color(0xFFE2E8F0);
  static const Color muted = Color(0xFF94A3B8);
  static const Color accent = Color(0xFF22D3EE);
  static const Color warn = Color(0xFFFBBF24);
  static const Color danger = Color(0xFFF87171);
  static const Color dangerBorder = Color(0x66F87171);
  static const Color onAccent = Color(0xFF0B1220);
}

ThemeData buildAppTheme() {
  return ThemeData(
    useMaterial3: true,
    brightness: Brightness.dark,
    scaffoldBackgroundColor: SfsColors.background,
    colorScheme: const ColorScheme.dark(
      surface: SfsColors.background,
      primary: SfsColors.accent,
      onPrimary: SfsColors.onAccent,
      secondary: SfsColors.accent,
      error: SfsColors.danger,
      onSurface: SfsColors.text,
      outline: SfsColors.border,
    ),
    appBarTheme: const AppBarTheme(
      backgroundColor: SfsColors.background,
      foregroundColor: SfsColors.text,
      elevation: 0,
      titleTextStyle: TextStyle(
        color: SfsColors.text,
        fontSize: 16,
        fontWeight: FontWeight.w600,
      ),
    ),
    bottomNavigationBarTheme: const BottomNavigationBarThemeData(
      backgroundColor: SfsColors.card,
      selectedItemColor: SfsColors.accent,
      unselectedItemColor: SfsColors.muted,
      type: BottomNavigationBarType.fixed,
    ),
    dialogTheme: const DialogThemeData(
      backgroundColor: SfsColors.card,
      surfaceTintColor: Colors.transparent,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.all(Radius.circular(16)),
        side: BorderSide(color: SfsColors.border),
      ),
    ),
    textTheme: const TextTheme(
      bodyMedium: TextStyle(color: SfsColors.text),
      bodySmall: TextStyle(color: SfsColors.muted),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: SfsColors.surfaceOverlay,
      hintStyle: const TextStyle(color: SfsColors.muted),
      contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: const BorderSide(color: SfsColors.border),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: const BorderSide(color: SfsColors.border),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: const BorderSide(color: SfsColors.accent),
      ),
    ),
    snackBarTheme: const SnackBarThemeData(
      backgroundColor: SfsColors.card,
      contentTextStyle: TextStyle(color: SfsColors.text),
    ),
    progressIndicatorTheme: const ProgressIndicatorThemeData(
      color: SfsColors.accent,
    ),
  );
}
