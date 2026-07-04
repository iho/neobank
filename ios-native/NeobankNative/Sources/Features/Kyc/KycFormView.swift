import SwiftUI

/// Shown by the home gate whenever KYC status isn't `approved`. Handles both
/// the never-submitted case (plain form) and the rejected case (banner +
/// resubmit) — the backend reports both as ordinary KYC statuses, there's no
/// separate "not started" signal.
struct KycFormView: View {
    let rejectionReason: String?

    @Environment(AuthController.self) private var authController
    @Environment(KycController.self) private var kycController

    @State private var fullName = ""
    @State private var dateOfBirth = Self.defaultDateOfBirth
    @State private var countryCode = ""
    @State private var documentType = ""
    @State private var documentNumber = ""
    @State private var isSubmitting = false
    @State private var errorMessage: String?
    @State private var showLogoutConfirmation = false

    private static let defaultDateOfBirth = Calendar.current.date(
        byAdding: .year, value: -18, to: .now
    ) ?? .now

    private static let minimumAge: DateComponents = {
        var components = DateComponents()
        components.year = -13
        return components
    }()

    private static let maximumAgeYears = 120

    private var dateOfBirthRange: ClosedRange<Date> {
        let now = Date.now
        let earliest = Calendar.current.date(byAdding: .year, value: -Self.maximumAgeYears, to: now) ?? .distantPast
        let latest = Calendar.current.date(byAdding: Self.minimumAge, to: now) ?? now
        return earliest...latest
    }

    private var canSubmit: Bool {
        !fullName.trimmingCharacters(in: .whitespaces).isEmpty && countryCode.count == 2
    }

    var body: some View {
        NavigationStack {
            ZStack {
                BrandBackground()

                ScrollView {
                    VStack(spacing: 24) {
                        VStack(spacing: 12) {
                            GlowIcon(systemName: "checkmark.shield.fill", diameter: 72, iconSize: 32)
                            Text("A few quick details")
                                .font(.title2.bold())
                            Text("We're required to verify your identity before your wallet can go live.")
                                .font(.subheadline)
                                .foregroundStyle(.secondary)
                                .multilineTextAlignment(.center)
                        }
                        .padding(.top, 8)

                        if let rejectionReason {
                            Label("Previous submission was rejected: \(rejectionReason)", systemImage: "exclamationmark.triangle.fill")
                                .font(.subheadline)
                                .foregroundStyle(.red)
                                .padding(12)
                                .frame(maxWidth: .infinity, alignment: .leading)
                                .surfaceCard(cornerRadius: 14)
                        }

                        VStack(spacing: 12) {
                            TextField("Full legal name", text: $fullName)
                                .brandField()
                                .textContentType(.name)

                            HStack {
                                Text("Date of birth")
                                    .foregroundStyle(.secondary)
                                Spacer()
                                DatePicker("", selection: $dateOfBirth, in: dateOfBirthRange, displayedComponents: .date)
                                    .labelsHidden()
                            }
                            .brandField()

                            TextField("Country code (ISO-2, e.g. US)", text: $countryCode)
                                .brandField()
                                .textInputAutocapitalization(.characters)
                                .autocorrectionDisabled()
                                .onChange(of: countryCode) { _, newValue in
                                    countryCode = String(newValue.uppercased().prefix(2))
                                }

                            TextField("Document type (optional)", text: $documentType)
                                .brandField()

                            TextField("Document number (optional)", text: $documentNumber)
                                .brandField()
                        }

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
                                Text("Submit")
                            }
                        }
                        .buttonStyle(.brandPrimary)
                        .disabled(isSubmitting || !canSubmit)
                    }
                    .padding(.horizontal, 24)
                    .padding(.bottom, 24)
                }
            }
            .navigationTitle("Verify your identity")
            .navigationBarTitleDisplayMode(.inline)
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
            } message: {
                Text("You'll need to sign back in to finish verifying your identity.")
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
                try await kycController.submit(
                    fullName: fullName.trimmingCharacters(in: .whitespaces),
                    dateOfBirth: Self.dobFormatter.string(from: dateOfBirth),
                    countryCode: countryCode,
                    documentType: documentType.trimmingCharacters(in: .whitespaces),
                    documentNumber: documentNumber.trimmingCharacters(in: .whitespaces)
                )
            } catch let error as APIError {
                errorMessage = error.message
            } catch {
                errorMessage = "Something went wrong. Please try again."
            }
        }
    }

    private static let dobFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.calendar = Calendar(identifier: .gregorian)
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.timeZone = TimeZone(secondsFromGMT: 0)
        formatter.dateFormat = "yyyy-MM-dd"
        return formatter
    }()
}

#Preview {
    KycFormView(rejectionReason: nil)
        .environment(AuthController.preview)
        .environment(KycController.preview)
}

#Preview("Rejected") {
    KycFormView(rejectionReason: "Document photo was blurry")
        .environment(AuthController.preview)
        .environment(KycController.preview)
}
