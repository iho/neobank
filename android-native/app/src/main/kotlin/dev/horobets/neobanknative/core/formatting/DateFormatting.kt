package dev.horobets.neobanknative.core.formatting

import java.time.OffsetDateTime
import java.time.format.DateTimeFormatter
import java.time.format.FormatStyle
import java.util.Locale

object DateFormatting {
    private val displayFormatter = DateTimeFormatter.ofLocalizedDateTime(FormatStyle.MEDIUM, FormatStyle.SHORT)
        .withLocale(Locale.getDefault())

    /** Renders a gateway ISO-8601 timestamp as a locale-formatted medium date / short time string. */
    fun formatted(iso: String): String = try {
        OffsetDateTime.parse(iso).format(displayFormatter)
    } catch (e: Exception) {
        iso
    }
}
