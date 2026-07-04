import SwiftUI

struct CardDetailView: View {
    let cardId: String

    /// Rendered until `cardsController`'s live list resolves (or if this
    /// card somehow drops out of it) — same fallback the Flutter screen
    /// uses when navigated to with an in-memory copy from the list.
    let initialCard: BankCard

    @Environment(CardsController.self) private var cardsController
    @State private var isSubmitting = false
    @State private var errorMessage: String?

    private var card: BankCard {
        if case .loaded(let cards) = cardsController.state, let match = cards.first(where: { $0.id == cardId }) {
            return match
        }
        return initialCard
    }

    var body: some View {
        ZStack {
            BrandBackground()

            VStack(spacing: 20) {
                VStack(alignment: .leading, spacing: 10) {
                    Text("•••• •••• •••• \(card.lastFour)")
                        .font(.title3.bold())

                    Text("Expires \(card.expiry)")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)

                    StatusPill(
                        text: card.status.capitalized,
                        systemImage: card.isFrozen ? "snowflake" : "checkmark.circle.fill",
                        tint: card.isFrozen ? .blue : .green
                    )

                    if let dailyLimit = card.dailyLimit {
                        Text("Daily limit: $\(LedgerAmount.formatted(dailyLimit))")
                            .font(.footnote)
                            .foregroundStyle(.secondary)
                    }

                    Text(card.onlineOnly ? "Online purchases only" : "Online + in-person")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(20)
                .surfaceCard()

                if let errorMessage {
                    Text(errorMessage)
                        .font(.footnote)
                        .foregroundStyle(.red)
                }

                Button {
                    toggleFreeze()
                } label: {
                    if isSubmitting {
                        ProgressView().tint(.white)
                    } else {
                        Label(
                            card.isFrozen ? "Unfreeze card" : "Freeze card",
                            systemImage: card.isFrozen ? "snowflake.circle.fill" : "snowflake"
                        )
                    }
                }
                .buttonStyle(.brandPrimary)
                .disabled(isSubmitting)

                Spacer()
            }
            .padding(20)
        }
        .navigationTitle("•••• \(card.lastFour)")
        .navigationBarTitleDisplayMode(.inline)
    }

    private func toggleFreeze() {
        guard !isSubmitting else { return }
        errorMessage = nil
        isSubmitting = true
        Task {
            defer { isSubmitting = false }
            do {
                try await cardsController.toggleFreeze(card)
            } catch let error as APIError {
                errorMessage = error.message
            } catch {
                errorMessage = "Something went wrong. Please try again."
            }
        }
    }
}

#Preview {
    NavigationStack {
        CardDetailView(
            cardId: "1",
            initialCard: BankCard(
                id: "1", userId: "u1", walletId: "w1", lastFour: "4242",
                status: "active", expiryMonth: 12, expiryYear: 2027,
                onlineOnly: false, dailyLimit: "500.00"
            )
        )
    }
    .environment(CardsController.preview)
}
