import Foundation

struct AuthTokens: Decodable, Sendable {
    let userId: String
    let accessToken: String
    let refreshToken: String
}

struct Profile: Decodable, Sendable {
    let userId: String
    let email: String
    let phone: String
    let status: String
    let kycStatus: String
    let createdAt: Date
    let fullName: String?
    let dateOfBirth: String?
    let countryCode: String?
}
