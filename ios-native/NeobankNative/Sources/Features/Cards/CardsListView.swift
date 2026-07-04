import SwiftUI

struct CardsListView: View {
    @Environment(SessionStore.self) private var sessionStore
    @Environment(CardsController.self) private var cardsController
    @State private var showIssueSheet = false

    var body: some View {
        NavigationStack {
            ZStack {
                BrandBackground()

                switch cardsController.state {
                case .loading:
                    ProgressView()
                case .failed(let error):
                    CardsErrorView(error: error)
                case .loaded(let cards):
                    if cards.isEmpty {
                        emptyState
                    } else {
                        list(cards)
                    }
                }
            }
            .navigationTitle("Cards")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button {
                        showIssueSheet = true
                    } label: {
                        Image(systemName: "plus")
                    }
                }
            }
            .sheet(isPresented: $showIssueSheet) {
                IssueCardView()
            }
        }
        .task(id: sessionStore.generation) { await cardsController.load() }
    }

    private var emptyState: some View {
        VStack(spacing: 16) {
            GlowIcon(systemName: "creditcard", diameter: 72, iconSize: 32)
            Text("No cards yet")
                .font(.title3.bold())
            Button("Issue a card") {
                showIssueSheet = true
            }
            .buttonStyle(.brandPrimary)
            .padding(.horizontal, 40)
        }
    }

    private func list(_ cards: [BankCard]) -> some View {
        ScrollView {
            VStack(spacing: 12) {
                ForEach(cards) { card in
                    NavigationLink {
                        CardDetailView(cardId: card.id, initialCard: card)
                    } label: {
                        CardRow(card: card)
                    }
                    .buttonStyle(.plain)
                }
            }
            .padding(20)
        }
        .refreshable { await cardsController.load() }
    }
}

private struct CardRow: View {
    let card: BankCard

    var body: some View {
        HStack(spacing: 14) {
            ZStack {
                Circle()
                    .fill((card.isFrozen ? Color.blue : Color.green).opacity(0.15))
                    .frame(width: 40, height: 40)
                Image(systemName: card.isFrozen ? "snowflake" : "creditcard.fill")
                    .font(.subheadline.weight(.semibold))
                    .foregroundStyle(card.isFrozen ? .blue : .green)
            }

            VStack(alignment: .leading, spacing: 2) {
                Text("•••• \(card.lastFour)")
                    .font(.subheadline.weight(.medium))
                Text("\(card.status.capitalized) · exp \(card.expiry)")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .lineLimit(1)
            }

            Spacer()

            Image(systemName: "chevron.right")
                .font(.caption.weight(.semibold))
                .foregroundStyle(.secondary)
        }
        .padding(12)
        .surfaceCard(cornerRadius: 14)
    }
}

private struct CardsErrorView: View {
    let error: APIError

    @Environment(CardsController.self) private var cardsController

    var body: some View {
        VStack(spacing: 12) {
            Text(error.message).multilineTextAlignment(.center)
            Button("Retry") {
                Task { await cardsController.load() }
            }
            .buttonStyle(.bordered)
        }
        .padding()
    }
}

#Preview {
    CardsListView()
        .environment(SessionStore())
        .environment(CardsController.preview)
}
