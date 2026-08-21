plugins {
    `java-library`
}

dependencies {
    runtimeOnly(project(":awssdk-autoconfigure"))

    runtimeOnly("io.opentelemetry.instrumentation:opentelemetry-aws-lambda-events-3.11")
    runtimeOnly("io.opentelemetry:opentelemetry-exporter-logging")
    runtimeOnly("io.opentelemetry:opentelemetry-exporter-otlp")
    runtimeOnly("io.opentelemetry:opentelemetry-extension-trace-propagators")
    runtimeOnly("io.opentelemetry:opentelemetry-sdk-extension-autoconfigure")
    runtimeOnly("io.opentelemetry.contrib:opentelemetry-aws-resources")
}

tasks {
    val layerContents = copySpec {
        from(configurations["runtimeClasspath"]) {
            into("java/lib")
        }

        // Can be used by redistributions of the wrapper to add more libraries.
        from("build/extensions") {
            into("java/lib")
        }

        from("scripts") {
                filePermissions {
                        unix("755")
                }
        }
    }

    val createZipLayer = register<Zip>("createLayer") {
        archiveFileName.set("opentelemetry-javawrapper-layer.zip")
        destinationDirectory.set(file("$buildDir/distributions"))
        with(layerContents)
    }

    val createTarLayer = register<Tar>("createTarLayer") {
        archiveFileName.set("opentelemetry-javawrapper-layer.tar.gz")
        destinationDirectory.set(file("$buildDir/distributions"))
        compression = Compression.GZIP
        with(layerContents)
    }

    named("assemble") {
        dependsOn(createZipLayer, createTarLayer)
    }
}

tasks.register("printOtelJavaInstrumentationVersion") {
    doLast {
        println(configurations.named("runtimeClasspath").get().resolvedConfiguration.resolvedArtifacts.find {  it.name == "opentelemetry-aws-lambda-events-3.11" }?.moduleVersion?.id?.version)
    }
}
