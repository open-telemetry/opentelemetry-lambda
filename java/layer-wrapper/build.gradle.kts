plugins {
    `java-library`
}

dependencies {
    runtimeOnly(project(":awssdk-autoconfigure"))

    runtimeOnly("io.opentelemetry.instrumentation:opentelemetry-aws-lambda-events-2.2")
    runtimeOnly("io.opentelemetry:opentelemetry-exporter-logging")
    runtimeOnly("io.opentelemetry:opentelemetry-exporter-otlp")
    runtimeOnly("io.opentelemetry:opentelemetry-extension-trace-propagators")
    runtimeOnly("io.opentelemetry:opentelemetry-sdk-extension-autoconfigure")
    runtimeOnly("io.opentelemetry.contrib:opentelemetry-aws-resources")
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

    named("assemble") {
        dependsOn(createLayer)
    }
}
