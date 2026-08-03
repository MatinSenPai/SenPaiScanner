plugins {
  alias(libs.plugins.android.application)
  alias(libs.plugins.compose.compiler)
  alias(libs.plugins.kotlin.serialization)
}

android {
    namespace = "com.matinsenpai.senpaiscanner"
    compileSdk = 36
    val releaseStoreFile = System.getenv("ANDROID_KEYSTORE_PATH")
        ?.takeIf { it.isNotBlank() }
        ?.let(::file)
        ?: file("../keystore/release.keystore")
    val releaseStorePassword = System.getenv("ORG_GRADLE_PROJECT_KEYSTORE_PASSWORD").orEmpty()
    val releaseKeyAlias = System.getenv("ORG_GRADLE_PROJECT_KEY_ALIAS").orEmpty()
    val releaseKeyPassword = System.getenv("ORG_GRADLE_PROJECT_KEY_PASSWORD").orEmpty()
    val releaseSigningAvailable = releaseStoreFile.isFile &&
        releaseStorePassword.isNotBlank() &&
        releaseKeyAlias.isNotBlank() &&
        releaseKeyPassword.isNotBlank()
    val appVersionName = System.getenv("ANDROID_VERSION_NAME")
        ?.takeIf { it.isNotBlank() }
        ?: "1.0.0"
    val appVersionCode = System.getenv("ANDROID_VERSION_CODE")
        ?.toIntOrNull()
        ?: 1
    defaultConfig {
        applicationId = "com.matinsenpai.senpaiscanner"
        minSdk = 24
        targetSdk = 36
        versionCode = appVersionCode
        versionName = appVersionName
    }

    signingConfigs {
        if (releaseSigningAvailable) {
            create("release") {
                storeFile = releaseStoreFile
                storePassword = releaseStorePassword
                keyAlias = releaseKeyAlias
                keyPassword = releaseKeyPassword
            }
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            isShrinkResources = true
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro")
            if (releaseSigningAvailable) {
                signingConfig = signingConfigs.getByName("release")
            }
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
    buildFeatures {
      compose = true
      aidl = false
      buildConfig = true
      shaders = false
    }

    splits {
        abi {
            isEnable = providers.gradleProperty("enableAbiSplits")
                .orNull
                ?.toBooleanStrictOrNull()
                ?: true
            reset()
            include("arm64-v8a", "armeabi-v7a")
            isUniversalApk = true
        }
    }

    packaging {
      resources {
        excludes += "/META-INF/{AL2.0,LGPL2.1}"
      }
    }
}

kotlin {
    jvmToolchain(17)
}

dependencies {
  val composeBom = platform(libs.androidx.compose.bom)
  implementation(composeBom)
  androidTestImplementation(composeBom)

  // Core Android dependencies
  implementation(libs.androidx.core.ktx)
  implementation(libs.androidx.lifecycle.runtime.ktx)
  implementation(libs.androidx.activity.compose)

  // Arch Components
  implementation(libs.androidx.lifecycle.runtime.compose)
  implementation(libs.androidx.lifecycle.viewmodel.compose)

  // Compose
  implementation(libs.androidx.compose.ui)
  implementation(libs.androidx.compose.ui.tooling.preview)
  implementation(libs.androidx.compose.material3)
  implementation(libs.androidx.compose.material.icons.core)
  implementation(libs.androidx.compose.material.icons.extended)
  // Tooling
  debugImplementation(libs.androidx.compose.ui.tooling)
  // Instrumented tests
  androidTestImplementation(libs.androidx.compose.ui.test.junit4)
  debugImplementation(libs.androidx.compose.ui.test.manifest)

  // Local tests: jUnit, coroutines, Android runner
  testImplementation(libs.junit)
  testImplementation(libs.kotlinx.coroutines.test)

  // Instrumented tests: jUnit rules and runners
  androidTestImplementation(libs.androidx.test.core)
  androidTestImplementation(libs.androidx.test.ext.junit)
  androidTestImplementation(libs.androidx.test.runner)
  androidTestImplementation(libs.androidx.test.espresso.core)

  implementation(libs.kotlinx.serialization.json)
  implementation(files("libs/senpaiscanner.aar"))
}
