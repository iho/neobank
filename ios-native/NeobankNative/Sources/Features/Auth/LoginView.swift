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
        ScrollView {
            VStack(spacing: 24) {
                VStack(spacing: 8) {
                    Image(systemName: "banknote.fill")
                        .font(.system(size: 40))
                        .foregroundStyle(.tint)
                    Text("Neobank")
                        .font(.largeTitle.bold())
                }
                .padding(.top, 48)

                VStack(spacing: 16) {
                    TextField("Email", text: $email)
                        .textFieldStyle(.roundedBorder)
                        .keyboardType(.emailAddress)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()
                        .textContentType(.username)
                        .focused($focusedField, equals: .email)
                        .submitLabel(.next)
                        .onSubmit { focusedField = .password }

                    SecureField("Password", text: $password)
                        .textFieldStyle(.roundedBorder)
                        .textContentType(.password)
                        .focused($focusedField, equals: .password)
                        .submitLabel(.go)
                        .onSubmit { submit() }
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
                        ProgressView().frame(maxWidth: .infinity)
                    } else {
                        Text("Log in").frame(maxWidth: .infinity)
                    }
                }
                .buttonStyle(.borderedProminent)
                .controlSize(.large)
                .disabled(isSubmitting || !canSubmit)

                NavigationLink("Don't have an account? Register") {
                    RegisterView()
                }
                .font(.footnote)
            }
            .padding(24)
        }
        .scrollDismissesKeyboard(.interactively)
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
