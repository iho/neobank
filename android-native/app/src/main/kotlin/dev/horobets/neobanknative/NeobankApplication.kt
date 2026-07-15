package dev.horobets.neobanknative

import android.app.Application
import dev.horobets.neobanknative.core.AppEnvironment

class NeobankApplication : Application() {
    val environment: AppEnvironment by lazy { AppEnvironment(applicationContext) }
}
