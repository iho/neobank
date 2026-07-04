import Foundation

struct APIError: Error, LocalizedError, Sendable {
    let message: String
    let statusCode: Int?
    let correlationId: String?

    var errorDescription: String? { message }

    static func network() -> APIError {
        APIError(message: "No connection. Check your network and try again.", statusCode: nil, correlationId: nil)
    }

    static func timeout() -> APIError {
        APIError(message: "The request timed out. Please try again.", statusCode: nil, correlationId: nil)
    }

    static func unauthenticated() -> APIError {
        APIError(message: "Your session has expired. Please log in again.", statusCode: 401, correlationId: nil)
    }

    static func decoding() -> APIError {
        APIError(message: "Something went wrong. Please try again.", statusCode: nil, correlationId: nil)
    }
}
