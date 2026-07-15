package dev.horobets.neobanknative.core.formatting

import java.math.BigDecimal
import java.text.DecimalFormat
import java.text.DecimalFormatSymbols
import java.util.Locale

object LedgerAmount {
    private val formatter = DecimalFormat("#,##0.00", DecimalFormatSymbols(Locale.US))

    /**
     * The ledger returns amounts at full stored precision (e.g.
     * "2.50000000" or "500.00000000"); render at 2 decimal places for
     * display. Goes through [BigDecimal], not [Double], to avoid binary-float
     * drift — and pins formatting to `Locale.US`, so grouping/decimal
     * separators don't drift with device locale.
     */
    fun formatted(raw: String): String {
        val decimal = raw.toBigDecimalOrNull() ?: return raw
        return formatter.format(decimal)
    }
}
