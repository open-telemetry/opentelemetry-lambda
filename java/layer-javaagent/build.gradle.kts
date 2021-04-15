plugins {
    `java-library`
}

dependencies {
    implementation("io.opentelemetry.javaagent", "opentelemetry-javaagent", classifier="all")
}

tasks {
    val createLayer by registering(Zip::class) {
        archiveFileName.set("opentelemetry-javaagent-layer.zip")
        destinationDirectory.set(file("$buildDir/distributions"))

        from(configurations["runtimeClasspath"]) {
            rename("opentelemetry-javaagent-.*.jar", "opentelemetry-javaagent.jar")
        }

        from("scripts")
    }

    val assemble by existing {
        dependsOn(createLayer)
    }
}
