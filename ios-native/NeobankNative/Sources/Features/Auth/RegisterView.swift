import SwiftUI

struct RegisterView: View {
    @Environment(AuthController.self) private var authController

    @State private var email = ""
    @State private var phone = ""
    @State private var password = ""
    @State private var inviteCode = ""
    @State private var isSubmitting = false
    @State private var errorMessage: String?

    private var canSubmit: Bool {
        email.contains("@") && password.count >= 8
    }

    var body: some View {
        ZStack {
            BrandBackground()

            ScrollView {
                VStack(spacing: 24) {
                    VStack(spacing: 8) {
                        Text("Create account")
                            .font(.system(size: 28, weight: .bold, design: .rounded))
                        Text("Set up your Neobank wallet")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                    .padding(.top, 24)

                    VStack(spacing: 12) {
                        TextField("Email", text: $email)
                            .brandField()
                            .keyboardType(.emailAddress)
                            .textInputAutocapitalization(.never)
                            .autocorrectionDisabled()

                        TextField("Phone (optional)", text: $phone)
                            .brandField()
                            .keyboardType(.phonePad)

                        SecureField("Password", text: $password)
                            .brandField()

                        if !password.isEmpty && password.count < 8 {
                            Text("At least 8 characters")
                                .font(.caption)
                                .foregroundStyle(.red)
                                .frame(maxWidth: .infinity, alignment: .leading)
                        }

                        TextField("Invite code (optional)", text: $inviteCode)
                            .brandField()
                            .textInputAutocapitalization(.characters)
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
                            Text("Create account")
                        }
                    }
                    .buttonStyle(.brandPrimary)
                    .disabled(isSubmitting || !canSubmit)
                }
                .padding(.horizontal, 24)
                .padding(.bottom, 24)
            }
            .scrollDismissesKeyboard(.interactively)
        }
        .navigationBarTitleDisplayMode(.inline)
    }

    private func submit() {
        guard canSubmit, !isSubmitting else { return }
        errorMessage = nil
        isSubmitting = true
        Task {
            defer { isSubmitting = false }
            do {
                try await authController.register(
                    email: email.trimmingCharacters(in: .whitespaces),
                    password: password,
                    phone: phone.trimmingCharacters(in: .whitespaces),
                    inviteCode: inviteCode.trimmingCharacters(in: .whitespaces)
                )
            } catch let error as APIError {
                errorMessage = error.message
            } catch {
                errorMessage = "Something went wrong. Please try again."
            }
        }
    }
}

#Preview {
    NavigationStack { RegisterView() }
        .environment(AuthController.preview)
}
