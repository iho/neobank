import SwiftUI

struct MainTabsView: View {
    var body: some View {
        TabView {
            WalletView()
                .tabItem { Label("Wallet", systemImage: "wallet.pass") }
            CardsListView()
                .tabItem { Label("Cards", systemImage: "creditcard") }
            NotificationsListView()
                .tabItem { Label("Alerts", systemImage: "bell") }
        }
    }
}
