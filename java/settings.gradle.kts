pluginManagement {
    plugins {
        id("com.diffplug.spotless") version "6.15.0"
        id("com.github.ben-manes.versions") version "0.45.0"
        id("com.github.johnrengelman.shadow") version "7.1.2"
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
