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
        Form {
            Section {
                TextField("Email", text: $email)
                    .keyboardType(.emailAddress)
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled()
                TextField("Phone (optional)", text: $phone)
                    .keyboardType(.phonePad)
                SecureField("Password", text: $password)
                TextField("Invite code (optional)", text: $inviteCode)
                    .textInputAutocapitalization(.characters)
            } footer: {
                if password.isEmpty == false && password.count < 8 {
                    Text("At least 8 characters")
                        .foregroundStyle(.red)
                }
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
                        Text("Create account").frame(maxWidth: .infinity)
                    }
                }
                .disabled(isSubmitting || !canSubmit)
            }
        }
        .navigationTitle("Create account")
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
