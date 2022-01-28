plugins {
    java
    id("com.github.johnrengelman.shadow")
}

dependencies {
    implementation("io.opentelemetry:opentelemetry-api")
    implementation("com.amazonaws:aws-lambda-java-core")
    implementation("com.amazonaws:aws-lambda-java-events")
    implementation("org.apache.logging.log4j:log4j-core")
    implementation("software.amazon.awssdk:s3")

    runtimeOnly("org.apache.logging.log4j:log4j-slf4j-impl")
}

tasks {
    assemble {
        dependsOn("shadowJar")
    }
}
