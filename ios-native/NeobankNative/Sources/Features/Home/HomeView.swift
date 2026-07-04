import SwiftUI

struct HomeView: View {
    @Environment(AuthController.self) private var authController

    var body: some View {
        NavigationStack {
            VStack(spacing: 16) {
                Image(systemName: "checkmark.seal.fill")
                    .font(.system(size: 48))
                    .foregroundStyle(.green)
                Text("You're logged in")
                    .font(.title2.bold())
                Button("Log out", role: .destructive) {
                    authController.logout()
                }
                .buttonStyle(.bordered)
            }
            .navigationTitle("Neobank")
        }
    }
}

#Preview {
    HomeView().environment(AuthController.preview)
}
