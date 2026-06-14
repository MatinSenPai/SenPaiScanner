# Add project specific ProGuard rules here.
# By default, the flags in this file are appended to flags specified
# in /usr/local/Cellar/android-sdk/24.3.3/tools/proguard/proguard-android.txt

# Keep Go mobile bindings
-keep class com.ehsatn.magik.Mobile { *; }
-keep class com.ehsatn.magik.Callback { *; }
