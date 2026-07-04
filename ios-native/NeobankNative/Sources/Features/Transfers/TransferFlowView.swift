import SwiftUI

struct TransferFlowView: View {
    private enum Step {
        case recipient
        case amount
        case confirmAndResult
    }

    private enum RecipientType: String, CaseIterable {
        case phone = "Phone"
        case email = "Email"
    }

    @Environment(\.dismiss) private var dismiss
    @Environment(WalletHomeController.self) private var walletController
    @Environment(TransferSubmitController.self) private var controller

    @State private var step: Step = .recipient
    @State private var recipientType: RecipientType = .phone
    @State private var recipient = ""
    @State private var amount = ""
    @State private var memo = ""

    private var recipientValid: Bool {
        !recipient.trimmingCharacters(in: .whitespaces).isEmpty
    }

    private var amountValid: Bool {
        guard let value = Double(amount.trimmingCharacters(in: .whitespaces)) else { return false }
        return value > 0 && amount.range(of: #"^\d+(\.\d{1,2})?$"#, options: .regularExpression) != nil
    }

    var body: some View {
        NavigationStack {
            ZStack {
                BrandBackground()

                Group {
                    switch step {
                    case .recipient:
                        recipientStep
                    case .amount:
                        amountStep
                    case .confirmAndResult:
                        confirmAndResultStep
                    }
                }
                .padding(24)
            }
            .navigationTitle("Send money")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }
        }
        .onAppear { controller.reset() }
    }

    private var recipientStep: some View {
        VStack(alignment: .leading, spacing: 20) {
            Text("Who are you sending to?")
                .font(.title3.bold())

            Picker("Recipient type", selection: $recipientType) {
                ForEach(RecipientType.allCases, id: \.self) { Text($0.rawValue).tag($0) }
            }
            .pickerStyle(.segmented)

            TextField(recipientType == .phone ? "Recipient phone" : "Recipient email", text: $recipient)
                .brandField()
                .keyboardType(recipientType == .phone ? .phonePad : .emailAddress)
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()

            Spacer()

            Button("Next") { step = .amount }
                .buttonStyle(.brandPrimary)
                .disabled(!recipientValid)
        }
    }

    private var amountStep: some View {
        VStack(alignment: .leading, spacing: 20) {
            Text("How much?")
                .font(.title3.bold())

            HStack {
                Text("$").foregroundStyle(.secondary)
                TextField("Amount (USD)", text: $amount)
                    .keyboardType(.decimalPad)
            }
            .brandField()

            TextField("Memo (optional)", text: $memo)
                .brandField()

            Spacer()

            Button("Next") { step = .confirmAndResult }
                .buttonStyle(.brandPrimary)
                .disabled(!amountValid)
        }
    }

    @ViewBuilder
    private var confirmAndResultStep: some View {
        switch controller.state {
        case .idle:
            confirmationView
        case .submitting:
            ProgressView()
        case .result(let transfer):
            resultView(for: transfer)
        case .failed(let error):
            ResultView(
                systemImage: "exclamationmark.circle.fill",
                tint: .red,
                title: "Something went wrong",
                message: error.message,
                primaryLabel: "Retry",
                onPrimary: submit,
                secondaryLabel: "Cancel",
                onSecondary: { dismiss() }
            )
        }
    }

    private var confirmationView: some View {
        VStack(alignment: .leading, spacing: 20) {
            Text("Confirm transfer")
                .font(.title3.bold())

            VStack(spacing: 12) {
                SummaryRow(label: "To", value: recipient)
                SummaryRow(label: "Amount", value: "$\(amount)")
                if !memo.isEmpty {
                    SummaryRow(label: "Memo", value: memo)
                }
            }
            .padding(16)
            .surfaceCard(cornerRadius: 14)

            Spacer()

            VStack(spacing: 12) {
                Button("Confirm & Send", action: submit)
                    .buttonStyle(.brandPrimary)
                Button("Edit") { step = .amount }
                    .font(.footnote)
            }
        }
    }

    private func resultView(for transfer: Transfer) -> some View {
        Group {
            if transfer.isCompleted {
                ResultView(
                    systemImage: "checkmark.circle.fill",
                    tint: .green,
                    title: "Sent!",
                    message: "$\(Self.formattedAmount(transfer.amount ?? amount)) to \(recipient)",
                    primaryLabel: "Done",
                    onPrimary: {
                        Task { await walletController.load() }
                        dismiss()
                    }
                )
            } else if transfer.isFailed {
                ResultView(
                    systemImage: "xmark.circle.fill",
                    tint: .red,
                    title: "Transfer declined",
                    message: transfer.failureReason ?? "The transfer could not be completed.",
                    primaryLabel: "Retry",
                    onPrimary: submit,
                    secondaryLabel: "Cancel",
                    onSecondary: { dismiss() }
                )
            } else {
                ResultView(
                    systemImage: "hourglass",
                    tint: .blue,
                    title: "Processing",
                    message: "Status: \(transfer.status ?? "unknown")",
                    primaryLabel: "Check again",
                    onPrimary: submit
                )
            }
        }
    }

    private func submit() {
        Task {
            await controller.submit(
                amount: amount.trimmingCharacters(in: .whitespaces),
                recipientPhone: recipientType == .phone ? recipient.trimmingCharacters(in: .whitespaces) : nil,
                recipientEmail: recipientType == .email ? recipient.trimmingCharacters(in: .whitespaces) : nil,
                memo: memo.trimmingCharacters(in: .whitespaces)
            )
        }
    }

    /// The ledger returns amounts at full stored precision (e.g.
    /// "2.50000000"); render at 2 decimal places for display. Goes through
    /// `Decimal`, not `Double`, so this can't introduce the binary-float
    /// drift `WalletBalance`/`WalletTransaction` avoid by keeping amounts as
    /// strings in the first place. Locale is pinned to `en_US_POSIX` — the
    /// device's locale would otherwise substitute "," for the decimal point
    /// (same class of bug as the Card expiry formatting fix: a "convenience"
    /// API silently applying locale-aware formatting to something that's a
    /// plain ASCII ledger value, not user-facing prose).
    private static func formattedAmount(_ raw: String) -> String {
        guard let decimal = Decimal(string: raw) else { return raw }
        let formatter = NumberFormatter()
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.numberStyle = .decimal
        formatter.minimumFractionDigits = 2
        formatter.maximumFractionDigits = 2
        return formatter.string(from: decimal as NSDecimalNumber) ?? raw
    }
}

private struct SummaryRow: View {
    let label: String
    let value: String

    var body: some View {
        HStack {
            Text(label).foregroundStyle(.secondary)
            Spacer()
            Text(value).fontWeight(.semibold)
        }
    }
}

private struct ResultView: View {
    let systemImage: String
    let tint: Color
    let title: String
    let message: String
    let primaryLabel: String
    let onPrimary: () -> Void
    var secondaryLabel: String?
    var onSecondary: (() -> Void)?

    var body: some View {
        VStack(spacing: 16) {
            Spacer()
            Image(systemName: systemImage)
                .font(.system(size: 56))
                .foregroundStyle(tint)
            Text(title).font(.title3.bold())
            Text(message)
                .font(.subheadline)
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)
            Spacer()
            Button(primaryLabel, action: onPrimary)
                .buttonStyle(.brandPrimary)
            if let secondaryLabel, let onSecondary {
                Button(secondaryLabel, action: onSecondary)
                    .font(.footnote)
            }
        }
        .frame(maxWidth: .infinity)
    }
}

#Preview {
    TransferFlowView()
        .environment(WalletHomeController.preview)
        .environment(TransferSubmitController.preview)
}
