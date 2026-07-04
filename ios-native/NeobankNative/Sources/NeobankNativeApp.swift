import SwiftUI

@main
struct NeobankNativeApp: App {
    @State private var environment = AppEnvironment()
    @AppStorage("appAppearance") private var appearance: AppAppearance = .system

    var body: some Scene {
        WindowGroup {
            RootView()
                .environment(environment.sessionStore)
                .environment(environment.authController)
                .environment(environment.kycController)
                .preferredColorScheme(appearance.colorScheme)
        }
    }
}
