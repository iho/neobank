package dev.horobets.neobanknative.core.config

import dev.horobets.neobanknative.BuildConfig

/**
 * `API_BASE_URL` is baked in at build time from the environment (see
 * app/build.gradle.kts), mirroring ios-native's `ProcessInfo`-based
 * override. Defaults to the emulator's host-loopback alias, 10.0.2.2,
 * rather than localhost, since the emulator is its own network namespace.
 */
object AppConfig {
    val apiBaseUrl: String = BuildConfig.API_BASE_URL
}
