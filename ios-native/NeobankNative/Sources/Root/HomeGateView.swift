import SwiftUI

/// Sits between the session gate and the main tabs: nothing past this point
/// should be reachable with KYC outstanding, so this is where that check
/// lives — not sprinkled across each destination.
struct HomeGateView: View {
    @Environment(SessionStore.self) private var sessionStore
    @Environment(KycController.self) private var kycController

    var body: some View {
        Group {
            switch kycController.state {
            case .loading:
                ProgressView()
            case .failed(let error):
                KycLoadErrorView(error: error)
            case .loaded(let info):
                switch info.status {
                case .approved:
                    MainTabsView()
                case .rejected:
                    KycFormView(rejectionReason: info.rejectionReason)
                case .pending, .manualReview:
                    if kycController.hasSubmittedThisSession {
                        KycPendingView()
                    } else {
                        KycFormView(rejectionReason: nil)
                    }
                }
            }
        }
        .task(id: sessionStore.generation) { await kycController.load() }
    }
}

private struct KycLoadErrorView: View {
    let error: APIError

    @Environment(KycController.self) private var kycController

    var body: some View {
        VStack(spacing: 12) {
            Text(error.message)
                .multilineTextAlignment(.center)
            Button("Retry") {
                Task { await kycController.load() }
            }
            .buttonStyle(.bordered)
        }
        .padding()
    }
}
