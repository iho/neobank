import Foundation

enum HTTPMethod: String, Sendable {
    case get = "GET"
    case post = "POST"
    case put = "PUT"
    case patch = "PATCH"
    case delete = "DELETE"
}

struct EmptyResponse: Decodable, Sendable {}

/// Talks to the gateway BFF. Two instances are wired up in practice: one with
/// `auth == nil` for `/v1/auth/login` and `/v1/auth/register` (which must not
/// try to attach or refresh a bearer token), and one with `auth` set for
/// every other call.
final class APIClient: Sendable {
    struct AuthContext: Sendable {
        let tokenProvider: @Sendable () -> String?
        let refresh: @Sendable () async throws -> String?
        let onSessionExpired: @Sendable () async -> Void
    }

    private let baseURL: URL
    private let session: URLSession
    private let auth: AuthContext?
    private let refreshCoordinator = RefreshCoordinator()

    private static let mutatingMethods: Set<HTTPMethod> = [.post, .put, .patch, .delete]

    private static let decoder: JSONDecoder = {
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        decoder.dateDecodingStrategy = .iso8601
        return decoder
    }()

    init(baseURL: URL, session: URLSession = .shared, auth: AuthContext? = nil) {
        self.baseURL = baseURL
        self.session = session
        self.auth = auth
    }

    @discardableResult
    func send<Response: Decodable>(
        _ path: String,
        method: HTTPMethod = .get,
        query: [String: String?]? = nil,
        body: [String: Any?]? = nil
    ) async throws -> Response {
        try await send(path, method: method, query: query, body: body, isRetryAfterRefresh: false)
    }

    private func send<Response: Decodable>(
        _ path: String,
        method: HTTPMethod,
        query: [String: String?]?,
        body: [String: Any?]?,
        isRetryAfterRefresh: Bool
    ) async throws -> Response {
        let request = try makeRequest(path: path, method: method, query: query, body: body)

        let data: Data
        let httpResponse: HTTPURLResponse
        do {
            let (responseData, response) = try await session.data(for: request)
            guard let http = response as? HTTPURLResponse else { throw APIError.network() }
            data = responseData
            httpResponse = http
        } catch let error as APIError {
            throw error
        } catch let error as URLError where error.code == .timedOut {
            throw APIError.timeout()
        } catch {
            throw APIError.network()
        }

        if httpResponse.statusCode == 401,
           let auth,
           !path.contains("/v1/auth/"),
           !isRetryAfterRefresh {
            guard (try? await refreshCoordinator.refresh(auth.refresh)) != nil else {
                await auth.onSessionExpired()
                throw APIError.unauthenticated()
            }
            return try await send(path, method: method, query: query, body: body, isRetryAfterRefresh: true)
        }

        guard (200...299).contains(httpResponse.statusCode) else {
            throw mapFailure(statusCode: httpResponse.statusCode, data: data, response: httpResponse)
        }

        if Response.self == EmptyResponse.self {
            return EmptyResponse() as! Response // swiftlint:disable:this force_cast
        }

        do {
            return try Self.decoder.decode(Response.self, from: data)
        } catch {
            throw APIError.decoding()
        }
    }

    private func makeRequest(
        path: String,
        method: HTTPMethod,
        query: [String: String?]?,
        body: [String: Any?]?
    ) throws -> URLRequest {
        var components = URLComponents(url: baseURL.appendingPathComponent(path), resolvingAgainstBaseURL: false)!
        if let query {
            let items = query.compactMapValues { $0 }.map { URLQueryItem(name: $0.key, value: $0.value) }
            if !items.isEmpty { components.queryItems = items }
        }
        var request = URLRequest(url: components.url!)
        request.httpMethod = method.rawValue
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue(UUID().uuidString, forHTTPHeaderField: "X-Correlation-Id")

        if Self.mutatingMethods.contains(method) {
            request.setValue(UUID().uuidString, forHTTPHeaderField: "Idempotency-Key")
        }
        if let body {
            let compacted = body.compactMapValues { $0 }
            request.httpBody = try? JSONSerialization.data(withJSONObject: compacted)
        }
        if let auth, let token = auth.tokenProvider() {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        return request
    }

    private func mapFailure(statusCode: Int, data: Data, response: HTTPURLResponse) -> APIError {
        if statusCode == 401 { return .unauthenticated() }

        var message = "Something went wrong. Please try again."
        if let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
           let apiMessage = json["error"] as? String {
            message = apiMessage
        }
        let correlationId = response.value(forHTTPHeaderField: "x-correlation-id")
        return APIError(message: message, statusCode: statusCode, correlationId: correlationId)
    }
}
