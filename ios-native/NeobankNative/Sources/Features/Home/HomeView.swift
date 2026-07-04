import SwiftUI

struct HomeView: View {
    @Environment(AuthController.self) private var authController
    @AppStorage("appAppearance") private var appearance: AppAppearance = .system

    var body: some View {
        NavigationStack {
            ZStack {
                BrandBackground()

                VStack(spacing: 20) {
                    StatusPill(text: "Verified", systemImage: "checkmark.seal.fill", tint: .green)
                    Text("You're logged in")
                        .font(.title2.bold())
                    Button("Log out", role: .destructive) {
                        authController.logout()
                    }
                    .buttonStyle(.bordered)
                }
                .padding()
            }
            .navigationTitle("Neobank")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Menu {
                        Picker("Appearance", selection: $appearance) {
                            ForEach(AppAppearance.allCases) { option in
                                Label(option.label, systemImage: option.systemImage).tag(option)
                            }
                        }
                    } label: {
                        Image(systemName: appearance.systemImage)
                    }
                    .accessibilityLabel("Appearance")
                }
            }
        }
    }
}

#Preview {
    HomeView().environment(AuthController.preview)
}
