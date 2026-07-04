import SwiftUI

struct MainTabsView: View {
    var body: some View {
        TabView {
            HomeView()
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
            ContentUnavailableView(title, systemImage: systemImage, description: Text("Coming soon."))
                .navigationTitle(title)
        }
    }
}
