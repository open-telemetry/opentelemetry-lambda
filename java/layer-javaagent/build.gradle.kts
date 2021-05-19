plugins {
    `java-library`
}

val agentDependency: String? = rootProject.findProperty("otel.lambda.javaagent.dependency") as String?

val agentClasspath by configurations.creating {
    extendsFrom(configurations["implementation"])
    isCanBeConsumed = false
    isCanBeResolved = true
    attributes {
        attribute(Bundling.BUNDLING_ATTRIBUTE, objects.named(Bundling::class.java, Bundling.SHADOWED))
    }
}

dependencies {
    if (agentDependency != null) {
        implementation(agentDependency)
    } else {
        implementation("io.opentelemetry.javaagent", "opentelemetry-javaagent", classifier = "all")
    }
}

tasks {
    val createLayer by registering(Zip::class) {
        archiveFileName.set("opentelemetry-javaagent-layer.zip")
        destinationDirectory.set(file("$buildDir/distributions"))

        from(agentClasspath) {
            rename(".*.jar", "opentelemetry-javaagent.jar")
        }

        from("scripts")
    }

    val assemble by existing {
        dependsOn(createLayer)
    }
}
