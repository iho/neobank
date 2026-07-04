import SwiftUI

struct WalletView: View {
    @Environment(SessionStore.self) private var sessionStore
    @Environment(WalletHomeController.self) private var walletController
    @Environment(AuthController.self) private var authController
    @AppStorage("appAppearance") private var appearance: AppAppearance = .system
    @State private var showSendFlow = false

    var body: some View {
        NavigationStack {
            ZStack {
                BrandBackground()

                switch walletController.state {
                case .loading:
                    ProgressView()
                case .failed(let error):
                    WalletErrorView(error: error)
                case .loaded(let snapshot):
                    walletContent(snapshot)
                }
            }
            .navigationTitle("Wallet")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Menu {
                        Picker("Appearance", selection: $appearance) {
                            ForEach(AppAppearance.allCases) { option in
                                Label(option.label, systemImage: option.systemImage).tag(option)
                            }
                        }
                        Button(role: .destructive) {
                            authController.logout()
                        } label: {
                            Label("Log out", systemImage: "rectangle.portrait.and.arrow.right")
                        }
                    } label: {
                        Image(systemName: appearance.systemImage)
                    }
                    .accessibilityLabel("Settings")
                }
            }
        }
        .task(id: sessionStore.generation) { await walletController.load() }
    }

    @ViewBuilder
    private func walletContent(_ snapshot: WalletHomeController.Snapshot) -> some View {
        ScrollView {
            VStack(spacing: 24) {
                BalanceCard(balance: snapshot.balance)

                VStack(alignment: .leading, spacing: 12) {
                    Text("Transactions")
                        .font(.headline)
                        .frame(maxWidth: .infinity, alignment: .leading)

                    if snapshot.transactions.isEmpty {
                        Text("No transactions yet")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 32)
                    } else {
                        VStack(spacing: 8) {
                            ForEach(snapshot.transactions) { transaction in
                                TransactionRow(transaction: transaction)
                                    .onAppear {
                                        if transaction.id == snapshot.transactions.last?.id {
                                            Task { await walletController.loadMore() }
                                        }
                                    }
                            }
                        }
                    }

                    if snapshot.isLoadingMore {
                        ProgressView().frame(maxWidth: .infinity).padding(.vertical, 12)
                    }
                }
            }
            .padding(20)
        }
        .refreshable { await walletController.load() }
        .safeAreaInset(edge: .bottom) {
            Button {
                showSendFlow = true
            } label: {
                Label("Send", systemImage: "paperplane.fill")
            }
            .buttonStyle(.brandPrimary)
            .padding(.horizontal, 20)
            .padding(.bottom, 8)
        }
        .sheet(isPresented: $showSendFlow) {
            TransferFlowView()
        }
    }
}

private struct BalanceCard: View {
    let balance: WalletBalance

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Available balance")
                .font(.subheadline)
                .foregroundStyle(.secondary)
            Text("\(balance.availableBalance) \(balance.currency)")
                .font(.system(size: 34, weight: .bold, design: .rounded))
            if let encumbered = balance.encumberedBalance {
                Text("Held: \(encumbered) \(balance.currency)")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(20)
        .surfaceCard()
    }
}

private struct TransactionRow: View {
    let transaction: WalletTransaction

    var body: some View {
        HStack(spacing: 14) {
            ZStack {
                Circle()
                    .fill((transaction.isCredit ? Color.green : Color.primary).opacity(0.15))
                    .frame(width: 40, height: 40)
                Image(systemName: transaction.isCredit ? "arrow.down" : "arrow.up")
                    .font(.subheadline.weight(.semibold))
                    .foregroundStyle(transaction.isCredit ? .green : .primary)
            }

            VStack(alignment: .leading, spacing: 2) {
                Text(transaction.counterparty ?? transaction.type)
                    .font(.subheadline.weight(.medium))
                Text("\(transaction.status) · \(Self.dateFormatter.string(from: transaction.createdAt))")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Spacer()

            Text("\(transaction.isCredit ? "+" : "-")\(transaction.amount) \(transaction.currency)")
                .font(.subheadline.weight(.semibold))
                .foregroundStyle(transaction.isCredit ? .green : .primary)
        }
        .padding(12)
        .surfaceCard(cornerRadius: 14)
    }

    private static let dateFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateStyle = .medium
        formatter.timeStyle = .short
        return formatter
    }()
}

private struct WalletErrorView: View {
    let error: APIError

    @Environment(WalletHomeController.self) private var walletController

    var body: some View {
        VStack(spacing: 12) {
            Text(error.message)
                .multilineTextAlignment(.center)
            Button("Retry") {
                Task { await walletController.load() }
            }
            .buttonStyle(.bordered)
        }
        .padding()
    }
}

#Preview {
    WalletView()
        .environment(SessionStore())
        .environment(AuthController.preview)
        .environment(WalletHomeController.preview)
}
