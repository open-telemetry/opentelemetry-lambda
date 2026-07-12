plugins {
    `java-library`
}

val agentClasspath = configurations.create("agentClasspath") {
    extendsFrom(configurations["implementation"])
    isCanBeConsumed = false
    isCanBeResolved = true
    attributes {
        attribute(Bundling.BUNDLING_ATTRIBUTE, objects.named(Bundling::class.java, Bundling.SHADOWED))
    }
}

dependencies {
    // version set in dependencyManagement/build.gradle.kts
    implementation("io.opentelemetry.javaagent:opentelemetry-javaagent")
}

tasks {
    val createLayer = register<Zip>("createLayer") {
        archiveFileName.set("opentelemetry-javaagent-layer.zip")
        destinationDirectory.set(file("$buildDir/distributions"))

        from(agentClasspath) {
            rename(".*.jar", "opentelemetry-javaagent.jar")
        }

        from("scripts")
    }

    named("assemble") {
        dependsOn(createLayer)
    }
}
