package dev.horobets.neobanknative.core.design

import android.content.Context
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Brightness4
import androidx.compose.material.icons.filled.BrightnessAuto
import androidx.compose.material.icons.filled.LightMode
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.graphics.vector.ImageVector
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

/**
 * User-selected override for the app's appearance. [SYSTEM] (the default)
 * tracks the device setting; [LIGHT]/[DARK] pin it regardless of device
 * setting, same as ios-native's `AppAppearance`.
 */
enum class AppAppearance(val label: String) {
    SYSTEM("Automatic"),
    LIGHT("Light"),
    DARK("Dark"),
    ;

    val icon: ImageVector
        get() = when (this) {
            SYSTEM -> Icons.Filled.BrightnessAuto
            LIGHT -> Icons.Filled.LightMode
            DARK -> Icons.Filled.Brightness4
        }
}

/** `@AppStorage`-equivalent: persists the appearance choice and exposes it as a hot [StateFlow]. */
class AppAppearanceStore(context: Context) {
    private val prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    private val _appearance = MutableStateFlow(
        AppAppearance.entries.find { it.name == prefs.getString(KEY, null) } ?: AppAppearance.SYSTEM,
    )
    val appearance: StateFlow<AppAppearance> = _appearance.asStateFlow()

    fun setAppearance(value: AppAppearance) {
        prefs.edit().putString(KEY, value.name).apply()
        _appearance.value = value
    }

    private companion object {
        const val PREFS_NAME = "dev.horobets.neobanknative.prefs"
        const val KEY = "app_appearance"
    }
}

val LocalAppAppearanceStore = staticCompositionLocalOf<AppAppearanceStore> {
    error("No AppAppearanceStore provided — wrap the composition in AppEnvironment's providers")
}
