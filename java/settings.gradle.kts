pluginManagement {
    plugins {
        id("com.diffplug.spotless") version "8.4.0"
        id("com.github.ben-manes.versions") version "0.53.0"
        id("com.gradleup.shadow") version "9.4.1"
    }
}

dependencyResolutionManagement {
    repositories {
        mavenCentral()
        mavenLocal()
    }
}

include(":awssdk-autoconfigure")
include(":dependencyManagement")
include(":layer-javaagent")
include(":layer-wrapper")
include(":sample-apps:aws-sdk")
include(":sample-apps:okhttp")
include(":sample-apps:sqs")

rootProject.name = "opentelemetry-lambda-java"
