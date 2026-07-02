import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:neobank_mobile/app.dart';

void main() {
  // flutter_secure_storage talks to a real platform channel; stub it so
  // session restore (read -> null) resolves without a device/emulator.
  const channel = MethodChannel('plugins.it_nomads.com/flutter_secure_storage');

  setUp(() {
    TestWidgetsFlutterBinding.ensureInitialized();
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(channel, (call) async {
      switch (call.method) {
        case 'read':
          return null;
        case 'readAll':
          return <String, String>{};
        default:
          return null;
      }
    });
  });

  tearDown(() {
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(channel, null);
  });

  testWidgets('app boots to the login screen when there is no stored session', (tester) async {
    await tester.pumpWidget(const ProviderScope(child: NeobankApp()));
    await tester.pumpAndSettle();

    expect(find.text('Neobank'), findsOneWidget);
    expect(find.text('Log in'), findsOneWidget);
  });
}
