import com.github.benmanes.gradle.versions.updates.DependencyUpdatesTask

plugins {
    `java-platform`

    id("com.github.ben-manes.versions")
}

data class DependencySet(val group: String, val version: String, val modules: List<String>)

val DEPENDENCY_BOMS = listOf(
    "io.opentelemetry.instrumentation:opentelemetry-instrumentation-bom-alpha:1.10.1-alpha",
    "io.grpc:grpc-bom:1.42.1",
    "org.apache.logging.log4j:log4j-bom:2.17.1",
    "software.amazon.awssdk:bom:2.17.112"
)

val DEPENDENCIES = listOf(
    "com.amazonaws:aws-lambda-java-core:1.2.1",
    "com.amazonaws:aws-lambda-java-events:3.11.0",
    "com.squareup.okhttp3:okhttp:4.9.3",
    "io.opentelemetry.javaagent:opentelemetry-javaagent:1.10.1"
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
    val stableKeyword = listOf("RELEASE", "FINAL", "GA").any { version.toUpperCase().contains(it) }
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
