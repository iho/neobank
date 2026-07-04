import SwiftUI

@main
struct NeobankNativeApp: App {
    @State private var environment = AppEnvironment()

    var body: some Scene {
        WindowGroup {
            RootView()
                .environment(environment.sessionStore)
                .environment(environment.authController)
        }
    }
}
