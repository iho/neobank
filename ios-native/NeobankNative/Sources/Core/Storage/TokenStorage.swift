import Foundation
import Security

final class TokenStorage: Sendable {
    private let service = "dev.horobets.neobankNative.tokens"

    private static let accessTokenKey = "access_token"
    private static let refreshTokenKey = "refresh_token"

    func readAccessToken() -> String? { read(key: Self.accessTokenKey) }

    func readRefreshToken() -> String? { read(key: Self.refreshTokenKey) }

    func saveTokens(accessToken: String, refreshToken: String) {
        write(key: Self.accessTokenKey, value: accessToken)
        write(key: Self.refreshTokenKey, value: refreshToken)
    }

    func clear() {
        delete(key: Self.accessTokenKey)
        delete(key: Self.refreshTokenKey)
    }

    private func read(key: String) -> String? {
        var query = baseQuery(key: key)
        query[kSecReturnData as String] = true
        query[kSecMatchLimit as String] = kSecMatchLimitOne

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        guard status == errSecSuccess, let data = result as? Data else { return nil }
        return String(data: data, encoding: .utf8)
    }

    private func write(key: String, value: String) {
        let data = Data(value.utf8)
        let query = baseQuery(key: key)
        let attributes: [String: Any] = [kSecValueData as String: data]

        let status = SecItemUpdate(query as CFDictionary, attributes as CFDictionary)
        if status == errSecItemNotFound {
            var newItem = query
            newItem[kSecValueData as String] = data
            SecItemAdd(newItem as CFDictionary, nil)
        }
    }

    private func delete(key: String) {
        let query = baseQuery(key: key)
        SecItemDelete(query as CFDictionary)
    }

    private func baseQuery(key: String) -> [String: Any] {
        [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
        ]
    }
}
