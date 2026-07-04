import SwiftUI

struct RootView: View {
    @Environment(SessionStore.self) private var sessionStore
    @Environment(AuthController.self) private var authController

    var body: some View {
        Group {
            switch sessionStore.status {
            case .unknown:
                ProgressView()
            case .unauthenticated:
                NavigationStack { LoginView() }
            case .authenticated:
                HomeView()
            }
        }
        .task { authController.bootstrap() }
    }
}
