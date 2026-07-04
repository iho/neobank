import Foundation

struct APIError: Error, LocalizedError, Sendable {
    let message: String
    let statusCode: Int?
    let correlationId: String?

    /// Raw response body for non-2xx responses. Most call sites ignore this
    /// — it exists for the rare case (e.g. `POST /v1/transfers` returning a
    /// 422 with a full declined-transfer payload, not an error envelope)
    /// where the caller needs to decode the body itself instead of treating
    /// the status code as a bare failure.
    let responseData: Data?

    var errorDescription: String? { message }

    init(message: String, statusCode: Int?, correlationId: String?, responseData: Data? = nil) {
        self.message = message
        self.statusCode = statusCode
        self.correlationId = correlationId
        self.responseData = responseData
    }

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
