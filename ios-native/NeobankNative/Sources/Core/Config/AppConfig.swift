import Foundation

enum AppConfig {
    static let apiBaseURL: URL = {
        let raw = ProcessInfo.processInfo.environment["API_BASE_URL"] ?? "http://localhost:8080"
        guard let url = URL(string: raw) else {
            fatalError("Invalid API_BASE_URL: \(raw)")
        }
        return url
    }()
}
