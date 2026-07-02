/// Build-time configuration, one value per flavor (dev/staging/prod).
///
/// Selected via `--dart-define`, e.g.:
///   flutter run --dart-define=API_BASE_URL=http://localhost:8080 --dart-define=FLAVOR=dev
class AppConfig {
  const AppConfig({required this.flavor, required this.apiBaseUrl});

  final String flavor;
  final String apiBaseUrl;

  static const _flavor = String.fromEnvironment('FLAVOR', defaultValue: 'dev');
  static const _apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080',
  );

  static const current = AppConfig(flavor: _flavor, apiBaseUrl: _apiBaseUrl);

  bool get isProd => flavor == 'prod';
}
