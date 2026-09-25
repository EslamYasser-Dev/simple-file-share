# FileShare — Mobile (Flutter)

Flutter app for the Simple File Share server: browse and upload files, manage
share links, watch usage, and react to live server events over SSE.

## Stack

- **Flutter** (stable, Dart 3) — Android and iOS
- **Riverpod** for state management
- **dio** for HTTP (REST, resumable uploads, authenticated downloads)
- **flutter_secure_storage** for the JWT
- **file_picker** / **share_plus** / **path_provider** for native integrations

## Development

```bash
flutter pub get
flutter run
```

Point the app at your API with a compile-time define (defaults: Android
emulator `http://10.0.2.2:3000`, otherwise `http://localhost:3000`):

```bash
flutter run --dart-define=API_BASE_URL=http://10.0.2.2:3000
```

Sign in with a username and password (JWT). OAuth is only available in the
web app.

## Scripts

| Command | Description |
| --- | --- |
| `flutter pub get` | Install dependencies |
| `flutter run` | Build and run on a connected device/emulator |
| `flutter analyze` | Static analysis (must be clean) |
| `flutter test` | Unit tests (formatting, SSE parsing, models) |
| `dart run flutter_launcher_icons` | Regenerate launcher icons from `assets/` |

## Configuration

`lib/src/config.dart` reads `API_BASE_URL` via `String.fromEnvironment`:

```bash
flutter run --dart-define=API_BASE_URL=https://files.example.com
```

## Project layout

```
lib/
├── main.dart                 # ProviderScope + MaterialApp
└── src/
    ├── config.dart           # API_BASE_URL resolution
    ├── theme.dart            # SfsColors + dark theme
    ├── format.dart           # bytes/date/path helpers (tested)
    ├── models.dart           # API DTOs (tested)
    ├── state/                # Riverpod auth controller + providers
    ├── services/             # API client, SSE events, token store
    └── screens/              # login, files, shares, account
test/                         # unit tests
assets/                       # launcher icon sources
```
