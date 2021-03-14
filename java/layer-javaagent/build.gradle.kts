plugins {
    `java-library`
}

configurations {
    val javaagent by creating {
        isCanBeConsumed = false
        isCanBeResolved = false
    }

    val javaagentClasspath by creating {
        extendsFrom(javaagent)
        isCanBeConsumed = false
        isCanBeResolved = true
        attributes {
            attribute(Bundling.BUNDLING_ATTRIBUTE, objects.named(Bundling::class.java, Bundling.SHADOWED))
        }
    }
}

dependencies {
    add("javaagent", "software.amazon.opentelemetry:aws-opentelemetry-agent")
}

tasks {
    val createLayer by registering(Zip::class) {
        archiveFileName.set("aws-opentelemetry-agent-layer.zip")
        destinationDirectory.set(file("$buildDir/distributions"))

        from(configurations["javaagentClasspath"]) {
            rename("aws-opentelemetry-agent-.*.jar", "aws-opentelemetry-agent.jar")
        }
    }

    val assemble by existing {
        dependsOn(createLayer)
    }
}
