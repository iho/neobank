import SwiftUI

struct MainTabsView: View {
    var body: some View {
        TabView {
            WalletView()
                .tabItem { Label("Wallet", systemImage: "wallet.pass") }
            ComingSoonView(title: "Cards", systemImage: "creditcard")
                .tabItem { Label("Cards", systemImage: "creditcard") }
            ComingSoonView(title: "Alerts", systemImage: "bell")
                .tabItem { Label("Alerts", systemImage: "bell") }
        }
    }
}

struct ComingSoonView: View {
    let title: String
    let systemImage: String

    var body: some View {
        NavigationStack {
            ZStack {
                BrandBackground()
                VStack(spacing: 16) {
                    GlowIcon(systemName: systemImage, diameter: 72, iconSize: 32)
                    Text(title).font(.title2.bold())
                    Text("Coming soon.")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
            }
            .navigationTitle(title)
        }
    }
}
