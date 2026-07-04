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
                .environment(environment.walletController)
                .environment(environment.cardsController)
                .environment(environment.notificationsController)
                .environment(environment.transferSubmitController)
                .preferredColorScheme(appearance.colorScheme)
        }
    }
}
