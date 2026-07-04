import SwiftUI

/// Rare with MVP auto-approve (submission usually resolves synchronously),
/// but the contract allows an async `pending` outcome — surfaced here with a
/// manual refresh in case a future backend makes KYC actually asynchronous.
struct KycPendingView: View {
    @Environment(AuthController.self) private var authController
    @Environment(KycController.self) private var kycController
    @State private var showLogoutConfirmation = false

    var body: some View {
        NavigationStack {
            VStack(spacing: 16) {
                Image(systemName: "hourglass")
                    .font(.system(size: 40))
                    .foregroundStyle(.tint)
                Text("We're reviewing your details")
                    .font(.title3.bold())
                Text("This usually only takes a moment.")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                Button("Check status") {
                    Task { await kycController.load() }
                }
                .buttonStyle(.bordered)
            }
            .padding()
            .toolbar {
                ToolbarItem(placement: .topBarLeading) {
                    Button {
                        showLogoutConfirmation = true
                    } label: {
                        Image(systemName: "arrow.backward")
                    }
                    .accessibilityLabel("Log out")
                }
            }
            .confirmationDialog(
                "Log out?",
                isPresented: $showLogoutConfirmation,
                titleVisibility: .visible
            ) {
                Button("Log out", role: .destructive) { authController.logout() }
                Button("Cancel", role: .cancel) {}
            }
        }
    }
}

#Preview {
    KycPendingView()
        .environment(AuthController.preview)
        .environment(KycController.preview)
}
