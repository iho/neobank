import SwiftUI

struct IssueCardView: View {
    @Environment(\.dismiss) private var dismiss
    @Environment(CardsController.self) private var cardsController

    @State private var cardholderName = ""
    @State private var dailyLimit = ""
    @State private var onlineOnly = false
    @State private var isSubmitting = false
    @State private var errorMessage: String?

    private var canSubmit: Bool {
        !cardholderName.trimmingCharacters(in: .whitespaces).isEmpty
    }

    var body: some View {
        NavigationStack {
            ZStack {
                BrandBackground()

                ScrollView {
                    VStack(spacing: 16) {
                        TextField("Cardholder name", text: $cardholderName)
                            .brandField()
                            .textContentType(.name)

                        TextField("Daily limit (optional)", text: $dailyLimit)
                            .brandField()
                            .keyboardType(.decimalPad)

                        Toggle("Online purchases only", isOn: $onlineOnly)
                            .brandField()

                        if let errorMessage {
                            Text(errorMessage)
                                .font(.footnote)
                                .foregroundStyle(.red)
                        }

                        Button {
                            submit()
                        } label: {
                            if isSubmitting {
                                ProgressView().tint(.white)
                            } else {
                                Text("Issue card")
                            }
                        }
                        .buttonStyle(.brandPrimary)
                        .disabled(isSubmitting || !canSubmit)
                    }
                    .padding(20)
                }
            }
            .navigationTitle("Issue a virtual card")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }
        }
    }

    private func submit() {
        guard canSubmit, !isSubmitting else { return }
        errorMessage = nil
        isSubmitting = true
        Task {
            defer { isSubmitting = false }
            do {
                try await cardsController.issueCard(
                    cardholderName: cardholderName.trimmingCharacters(in: .whitespaces),
                    dailyLimit: dailyLimit.trimmingCharacters(in: .whitespaces),
                    onlineOnly: onlineOnly
                )
                dismiss()
            } catch let error as APIError {
                errorMessage = error.message
            } catch {
                errorMessage = "Something went wrong. Please try again."
            }
        }
    }
}

#Preview {
    IssueCardView().environment(CardsController.preview)
}
