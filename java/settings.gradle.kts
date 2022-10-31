pluginManagement {
    plugins {
        id("com.diffplug.spotless") version "5.14.3"
        id("com.github.ben-manes.versions") version "0.39.0"
        id("com.github.johnrengelman.shadow") version "7.0.0"
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
