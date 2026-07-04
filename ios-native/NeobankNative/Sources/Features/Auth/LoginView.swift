import SwiftUI

struct LoginView: View {
    @Environment(AuthController.self) private var authController

    @State private var email = ""
    @State private var password = ""
    @State private var isSubmitting = false
    @State private var errorMessage: String?
    @FocusState private var focusedField: Field?

    private enum Field {
        case email, password
    }

    private var canSubmit: Bool {
        email.contains("@") && !password.isEmpty
    }

    var body: some View {
        ZStack {
            BrandBackground()

            ScrollView {
                VStack(spacing: 32) {
                    VStack(spacing: 16) {
                        GlowIcon(systemName: "banknote.fill")
                        Text("Neobank")
                            .font(.system(size: 32, weight: .bold, design: .rounded))
                    }
                    .padding(.top, 56)

                    VStack(spacing: 12) {
                        TextField("Email", text: $email)
                            .brandField()
                            .keyboardType(.emailAddress)
                            .textInputAutocapitalization(.never)
                            .autocorrectionDisabled()
                            .textContentType(.username)
                            .focused($focusedField, equals: .email)
                            .submitLabel(.next)
                            .onSubmit { focusedField = .password }

                        SecureField("Password", text: $password)
                            .brandField()
                            .textContentType(.password)
                            .focused($focusedField, equals: .password)
                            .submitLabel(.go)
                            .onSubmit { submit() }
                    }
                    .padding(.horizontal, 24)

                    if let errorMessage {
                        Text(errorMessage)
                            .font(.footnote)
                            .foregroundStyle(.red)
                            .padding(.horizontal, 24)
                    }

                    VStack(spacing: 16) {
                        Button {
                            submit()
                        } label: {
                            if isSubmitting {
                                ProgressView().tint(.white)
                            } else {
                                Text("Log in")
                            }
                        }
                        .buttonStyle(.brandPrimary)
                        .disabled(isSubmitting || !canSubmit)

                        NavigationLink("Don't have an account? Register") {
                            RegisterView()
                        }
                        .font(.footnote.weight(.medium))
                    }
                    .padding(.horizontal, 24)
                }
                .padding(.bottom, 24)
            }
            .scrollDismissesKeyboard(.interactively)
        }
    }

    private func submit() {
        guard canSubmit, !isSubmitting else { return }
        errorMessage = nil
        isSubmitting = true
        Task {
            defer { isSubmitting = false }
            do {
                try await authController.login(email: email.trimmingCharacters(in: .whitespaces), password: password)
            } catch let error as APIError {
                errorMessage = error.message
            } catch {
                errorMessage = "Something went wrong. Please try again."
            }
        }
    }
}

#Preview {
    NavigationStack { LoginView() }
        .environment(AuthController.preview)
}
