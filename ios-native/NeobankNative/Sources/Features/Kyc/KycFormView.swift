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
            Form {
                Section {
                    VStack(alignment: .leading, spacing: 8) {
                        Text("A few quick details")
                            .font(.title2.bold())
                        Text("We're required to verify your identity before your wallet can go live.")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                    .listRowInsets(EdgeInsets())
                    .padding(.vertical, 8)
                }

                if let rejectionReason {
                    Section {
                        Label("Previous submission was rejected: \(rejectionReason)", systemImage: "exclamationmark.triangle.fill")
                            .foregroundStyle(.red)
                    }
                }

                Section {
                    TextField("Full legal name", text: $fullName)
                        .textContentType(.name)
                    DatePicker("Date of birth", selection: $dateOfBirth, in: dateOfBirthRange, displayedComponents: .date)
                    TextField("Country code (ISO-2, e.g. US)", text: $countryCode)
                        .textInputAutocapitalization(.characters)
                        .autocorrectionDisabled()
                        .onChange(of: countryCode) { _, newValue in
                            countryCode = String(newValue.uppercased().prefix(2))
                        }
                    TextField("Document type (optional)", text: $documentType)
                    TextField("Document number (optional)", text: $documentNumber)
                }

                if let errorMessage {
                    Section {
                        Text(errorMessage).foregroundStyle(.red)
                    }
                }

                Section {
                    Button {
                        submit()
                    } label: {
                        if isSubmitting {
                            ProgressView().frame(maxWidth: .infinity)
                        } else {
                            Text("Submit").frame(maxWidth: .infinity)
                        }
                    }
                    .disabled(isSubmitting || !canSubmit)
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
