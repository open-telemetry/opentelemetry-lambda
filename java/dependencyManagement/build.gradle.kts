import com.github.benmanes.gradle.versions.updates.DependencyUpdatesTask

plugins {
    `java-platform`

    id("com.github.ben-manes.versions")
}

data class DependencySet(val group: String, val version: String, val modules: List<String>)

val DEPENDENCY_BOMS = listOf(
    "io.opentelemetry.instrumentation:opentelemetry-instrumentation-bom-alpha:2.0.0-alpha",
    "org.apache.logging.log4j:log4j-bom:2.22.1",
    "software.amazon.awssdk:bom:2.24.2"
)

val DEPENDENCIES = listOf(
    "com.amazonaws:aws-lambda-java-core:1.2.3",
    "com.amazonaws:aws-lambda-java-events:3.11.4",
    "com.squareup.okhttp3:okhttp:4.12.0",
    "io.opentelemetry.javaagent:opentelemetry-javaagent:2.0.0"
)

javaPlatform {
    allowDependencies()
}

dependencies {
    for (bom in DEPENDENCY_BOMS) {
        api(platform(bom))
    }
    constraints {
        for (dependency in DEPENDENCIES) {
            api(dependency)
        }
    }
}

fun isNonStable(version: String): Boolean {
    val stableKeyword = listOf("RELEASE", "FINAL", "GA").any { version.uppercase().contains(it) }
    val regex = "^[0-9,.v-]+(-r)?$".toRegex()
    val isGuava = version.endsWith("-jre")
    val isStable = stableKeyword || regex.matches(version) || isGuava
    return isStable.not()
}

tasks {
    named<DependencyUpdatesTask>("dependencyUpdates") {
        revision = "release"
        checkConstraints = true

        rejectVersionIf {
            isNonStable(candidate.version)
        }
    }
}
