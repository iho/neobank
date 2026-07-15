package dev.horobets.neobanknative.core.design

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.LocalContentColor
import androidx.compose.material3.LocalTextStyle
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.compositionLocalOf
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.blur
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

/**
 * Shared visual language: dark navy-to-black gradient / soft gray-to-white
 * in light mode, a blue-to-purple brand gradient for accents, and glassy
 * (low-opacity fill) cards. Colors are computed from [LocalIsDarkTheme]
 * rather than stored, so the app follows system appearance with no manual
 * per-screen toggle.
 */
object Theme {
    val brandBlue = Color(0xFF3B82F6)
    val brandPurple = Color(0xFFA855F7)

    val brandGradientBrush = Brush.linearGradient(listOf(brandBlue, brandPurple))

    fun backgroundGradientBrush(isDark: Boolean): Brush = if (isDark) {
        Brush.linearGradient(listOf(Color(0xFF1A1A33), Color(0xFF0D0D26)))
    } else {
        Brush.linearGradient(listOf(Color(0xFFF2F2F2), Color.White))
    }

    fun surfaceFill(isDark: Boolean): Color = if (isDark) Color.White.copy(alpha = 0.05f) else Color.Black.copy(alpha = 0.04f)

    fun surfaceStroke(isDark: Boolean): Color = if (isDark) Color.White.copy(alpha = 0.1f) else Color.Black.copy(alpha = 0.06f)

    fun fieldFill(isDark: Boolean): Color = if (isDark) Color.White.copy(alpha = 0.08f) else Color.White
}

/** Whether the *effective* theme (system default, or pinned via [AppAppearance]) is dark. */
val LocalIsDarkTheme = compositionLocalOf { false }

@Composable
fun NeobankTheme(isDark: Boolean, content: @Composable () -> Unit) {
    val colorScheme = if (isDark) {
        darkColorScheme(primary = Theme.brandBlue, secondary = Theme.brandPurple)
    } else {
        lightColorScheme(primary = Theme.brandBlue, secondary = Theme.brandPurple)
    }
    CompositionLocalProvider(LocalIsDarkTheme provides isDark) {
        MaterialTheme(colorScheme = colorScheme, content = content)
    }
}

/** Full-bleed background — drop behind any screen's content. */
@Composable
fun BrandBackground(modifier: Modifier = Modifier) {
    val isDark = LocalIsDarkTheme.current
    Box(modifier = modifier.fillMaxSize().background(Theme.backgroundGradientBrush(isDark)))
}

/** A glassy rounded-rect surface, used for cards and grouped content. */
@Composable
fun Modifier.surfaceCard(cornerRadius: Dp = 20.dp): Modifier {
    val isDark = LocalIsDarkTheme.current
    return this
        .clip(RoundedCornerShape(cornerRadius))
        .background(Theme.surfaceFill(isDark))
}

/** Text field chrome matching the brand: filled rounded rect, no border. */
@Composable
fun BrandTextField(
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    modifier: Modifier = Modifier,
    singleLine: Boolean = true,
    isPassword: Boolean = false,
    keyboardOptions: KeyboardOptions = KeyboardOptions.Default,
    keyboardActions: KeyboardActions = KeyboardActions.Default,
) {
    val isDark = LocalIsDarkTheme.current
    val visualTransformation = if (isPassword) PasswordVisualTransformation() else VisualTransformation.None

    Box(
        modifier = modifier
            .clip(RoundedCornerShape(12.dp))
            .background(Theme.fieldFill(isDark))
            .padding(14.dp),
    ) {
        if (value.isEmpty()) {
            Text(placeholder, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
        }
        BasicTextField(
            value = value,
            onValueChange = onValueChange,
            modifier = Modifier.fillMaxWidth(),
            singleLine = singleLine,
            visualTransformation = visualTransformation,
            keyboardOptions = keyboardOptions,
            keyboardActions = keyboardActions,
            textStyle = LocalTextStyle.current.copy(color = MaterialTheme.colorScheme.onSurface, fontSize = 16.sp),
            cursorBrush = Brush.linearGradient(listOf(Theme.brandBlue, Theme.brandBlue)),
        )
    }
}

/**
 * Full-width primary call-to-action: brand gradient in dark mode, solid
 * blue in light (a plain gradient reads muddier against a white background).
 */
@Composable
fun PrimaryButton(
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    isLoading: Boolean = false,
    content: @Composable () -> Unit,
) {
    val isDark = LocalIsDarkTheme.current
    val background = if (isDark) Theme.brandGradientBrush else Brush.linearGradient(listOf(Theme.brandBlue, Theme.brandBlue))
    val isEnabled = enabled && !isLoading

    Box(
        modifier = modifier
            .clip(RoundedCornerShape(16.dp))
            .background(background)
            .clickable(enabled = isEnabled, onClick = onClick)
            .padding(vertical = 16.dp),
        contentAlignment = Alignment.Center,
    ) {
        CompositionLocalProvider(LocalContentColor provides Color.White) {
            if (isLoading) {
                CircularProgressIndicator(color = Color.White, modifier = Modifier.size(20.dp), strokeWidth = 2.dp)
            } else {
                content()
            }
        }
    }
}

/** A circular icon badge with a soft color glow behind it — used for hero icons on auth/status screens. */
@Composable
fun GlowIcon(icon: ImageVector, modifier: Modifier = Modifier, diameter: Dp = 80.dp, iconSize: Dp = 36.dp) {
    Box(modifier = modifier.size(diameter), contentAlignment = Alignment.Center) {
        Box(
            modifier = Modifier
                .size(diameter)
                .blur(diameter * 0.25f)
                .clip(CircleShape)
                .background(Brush.linearGradient(listOf(Theme.brandBlue.copy(alpha = 0.35f), Theme.brandPurple.copy(alpha = 0.35f)))),
        )
        Icon(imageVector = icon, contentDescription = null, modifier = Modifier.size(iconSize), tint = Theme.brandBlue)
    }
}

/** A small colored capsule with an icon — status/state indicators (KYC status, verification badges, etc). */
@Composable
fun StatusPill(text: String, icon: ImageVector, tint: Color, modifier: Modifier = Modifier) {
    Box(
        modifier = modifier
            .clip(RoundedCornerShape(50))
            .background(tint.copy(alpha = 0.15f))
            .padding(horizontal = 10.dp, vertical = 5.dp),
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Icon(icon, contentDescription = null, tint = tint, modifier = Modifier.size(14.dp))
            Spacer(modifier = Modifier.size(4.dp))
            Text(text, color = tint, style = TextStyle(fontSize = 12.sp))
        }
    }
}
