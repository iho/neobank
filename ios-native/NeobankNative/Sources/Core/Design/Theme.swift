import SwiftUI

/// Shared visual language: dark navy-to-black gradient / soft gray-to-white
/// in light mode, a blue-to-purple brand gradient for accents, and glassy
/// (low-opacity fill) cards. Colors are computed from `colorScheme` rather
/// than stored, so the app follows system appearance with no manual toggle.
enum Theme {
    static let brandGradient = LinearGradient(
        colors: [.blue, .purple],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )

    static func backgroundGradient(for colorScheme: ColorScheme) -> LinearGradient {
        colorScheme == .dark
            ? LinearGradient(
                colors: [Color(red: 0.1, green: 0.1, blue: 0.2), Color(red: 0.05, green: 0.05, blue: 0.15)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
            : LinearGradient(
                colors: [Color(white: 0.95), .white],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
    }

    static func surfaceFill(for colorScheme: ColorScheme) -> Color {
        colorScheme == .dark ? .white.opacity(0.05) : .black.opacity(0.04)
    }

    static func surfaceStroke(for colorScheme: ColorScheme) -> Color {
        colorScheme == .dark ? .white.opacity(0.1) : .black.opacity(0.06)
    }

    static func fieldFill(for colorScheme: ColorScheme) -> Color {
        colorScheme == .dark ? .white.opacity(0.08) : .white
    }
}

/// Full-bleed background — drop behind any screen's content.
struct BrandBackground: View {
    @Environment(\.colorScheme) private var colorScheme

    var body: some View {
        Theme.backgroundGradient(for: colorScheme).ignoresSafeArea()
    }
}

/// A glassy rounded-rect surface, used for cards and grouped content.
struct SurfaceCardStyle: ViewModifier {
    @Environment(\.colorScheme) private var colorScheme
    var cornerRadius: CGFloat = 20

    func body(content: Content) -> some View {
        content
            .background(
                RoundedRectangle(cornerRadius: cornerRadius)
                    .fill(Theme.surfaceFill(for: colorScheme))
                    .overlay(
                        RoundedRectangle(cornerRadius: cornerRadius)
                            .stroke(Theme.surfaceStroke(for: colorScheme), lineWidth: 1)
                    )
            )
    }
}

extension View {
    func surfaceCard(cornerRadius: CGFloat = 20) -> some View {
        modifier(SurfaceCardStyle(cornerRadius: cornerRadius))
    }
}

/// Text field chrome matching the brand: filled rounded rect, no border.
struct BrandFieldStyle: ViewModifier {
    @Environment(\.colorScheme) private var colorScheme

    func body(content: Content) -> some View {
        content
            .padding(14)
            .background(Theme.fieldFill(for: colorScheme))
            .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

extension View {
    func brandField() -> some View {
        modifier(BrandFieldStyle())
    }
}

/// Full-width primary call-to-action: brand gradient in dark mode, solid
/// blue in light (a plain gradient reads muddier against a white background).
struct PrimaryButtonStyle: ButtonStyle {
    @Environment(\.colorScheme) private var colorScheme
    @Environment(\.isEnabled) private var isEnabled

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.headline)
            .foregroundStyle(.white)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 16)
            .background(
                Group {
                    if colorScheme == .dark {
                        Theme.brandGradient
                    } else {
                        Color.blue
                    }
                }
            )
            .clipShape(RoundedRectangle(cornerRadius: 16))
            .opacity(isEnabled ? (configuration.isPressed ? 0.85 : 1) : 0.4)
    }
}

extension ButtonStyle where Self == PrimaryButtonStyle {
    static var brandPrimary: PrimaryButtonStyle { PrimaryButtonStyle() }
}

/// A circular icon badge with a soft color glow behind it — used for hero
/// icons on auth/status screens.
struct GlowIcon: View {
    let systemName: String
    var diameter: CGFloat = 80
    var iconSize: CGFloat = 36

    var body: some View {
        ZStack {
            Circle()
                .fill(Theme.brandGradient.opacity(0.35))
                .frame(width: diameter, height: diameter)
                .blur(radius: diameter * 0.25)

            Image(systemName: systemName)
                .font(.system(size: iconSize, weight: .medium))
                .foregroundStyle(Theme.brandGradient)
        }
    }
}

/// A small colored capsule with an icon — status/state indicators (KYC
/// status, verification badges, etc).
struct StatusPill: View {
    let text: String
    let systemImage: String
    let tint: Color

    var body: some View {
        Label(text, systemImage: systemImage)
            .font(.caption.weight(.medium))
            .foregroundStyle(tint)
            .padding(.horizontal, 10)
            .padding(.vertical, 5)
            .background(tint.opacity(0.15))
            .clipShape(Capsule())
    }
}
