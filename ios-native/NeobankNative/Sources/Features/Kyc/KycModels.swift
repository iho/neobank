import Foundation

enum KycStatus: Sendable {
    case pending
    case approved
    case rejected
    case manualReview

    init(raw: String) {
        switch raw {
        case "approved": self = .approved
        case "rejected": self = .rejected
        case "manual_review": self = .manualReview
        default: self = .pending
        }
    }
}

struct KycStatusInfo: Decodable, Sendable {
    let status: KycStatus
    let rejectionReason: String?

    private enum CodingKeys: String, CodingKey {
        case status, rejectionReason
    }

    init(status: KycStatus, rejectionReason: String?) {
        self.status = status
        self.rejectionReason = rejectionReason
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        status = KycStatus(raw: try container.decode(String.self, forKey: .status))
        rejectionReason = try container.decodeIfPresent(String.self, forKey: .rejectionReason)
    }
}

struct KycSubmitResult: Decodable, Sendable {
    let kycCaseId: String
    let status: KycStatus
    let walletId: String?
    let rejectionReason: String?

    private enum CodingKeys: String, CodingKey {
        case kycCaseId, status, walletId, rejectionReason
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        kycCaseId = try container.decode(String.self, forKey: .kycCaseId)
        status = KycStatus(raw: try container.decode(String.self, forKey: .status))
        walletId = try container.decodeIfPresent(String.self, forKey: .walletId)
        rejectionReason = try container.decodeIfPresent(String.self, forKey: .rejectionReason)
    }
}
