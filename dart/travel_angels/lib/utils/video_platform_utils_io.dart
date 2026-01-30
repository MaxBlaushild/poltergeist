import 'dart:io';

/// IO implementation: full video editing supported on Android, iOS, macOS only.

bool get supportsFullVideoEditing =>
    Platform.isAndroid || Platform.isIOS || Platform.isMacOS;
