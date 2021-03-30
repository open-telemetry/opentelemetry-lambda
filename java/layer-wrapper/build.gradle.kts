plugins {
    `java-library`
}

dependencies {
    runtimeOnly(project(":awssdk-autoconfigure"))

    runtimeOnly("io.opentelemetry.instrumentation:opentelemetry-aws-lambda-1.0")
    runtimeOnly("io.opentelemetry:opentelemetry-exporter-logging")
    runtimeOnly("io.opentelemetry:opentelemetry-exporter-otlp")
    runtimeOnly("io.opentelemetry:opentelemetry-extension-trace-propagators")
    runtimeOnly("io.opentelemetry:opentelemetry-sdk-extension-autoconfigure")
}

tasks {
    val createLayer by registering(Zip::class) {
        archiveFileName.set("opentelemetry-java-wrapper.zip")
        destinationDirectory.set(file("$buildDir/distributions"))

        from(configurations["runtimeClasspath"]) {
            into("java/lib")
        }

        from("scripts")
    }

    val assemble by existing {
        dependsOn(createLayer)
    }
}
