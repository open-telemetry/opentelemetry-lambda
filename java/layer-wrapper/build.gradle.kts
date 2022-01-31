plugins {
    `java-library`
}


dependencies {
    runtimeOnly(project(":awssdk-autoconfigure"))

    // TODO: Remove this when fix released upstream
    // See here: https://github.com/open-telemetry/opentelemetry-java-instrumentation/pull/5284
    runtimeOnly("com.fasterxml.jackson.core:jackson-core")

    runtimeOnly("io.grpc:grpc-netty-shaded")
    runtimeOnly("io.opentelemetry.instrumentation:opentelemetry-aws-lambda-1.0")
    runtimeOnly("io.opentelemetry:opentelemetry-exporter-logging")
    runtimeOnly("io.opentelemetry:opentelemetry-exporter-otlp")
    runtimeOnly("io.opentelemetry:opentelemetry-exporter-otlp-metrics")
    runtimeOnly("io.opentelemetry:opentelemetry-extension-trace-propagators")
    runtimeOnly("io.opentelemetry:opentelemetry-sdk-extension-autoconfigure")
    runtimeOnly("io.opentelemetry:opentelemetry-sdk-extension-aws")
}

tasks {
    val createLayer by registering(Zip::class) {
        archiveFileName.set("opentelemetry-java-wrapper.zip")
        destinationDirectory.set(file("$buildDir/distributions"))

        from(configurations["runtimeClasspath"]) {
            into("java/lib")
        }

        // Can be used by redistributions of the wrapper to add more libraries.
        from("build/extensions") {
            into("java/lib")
        }

        from("scripts")
    }

    val assemble by existing {
        dependsOn(createLayer)
    }
}
