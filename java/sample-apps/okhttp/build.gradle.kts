plugins {
    java
    id("com.github.johnrengelman.shadow")
}

dependencies {
    implementation("com.amazonaws:aws-lambda-java-core")
    implementation("com.amazonaws:aws-lambda-java-events")
    implementation("com.squareup.okhttp3:okhttp")
    implementation("io.opentelemetry.instrumentation:opentelemetry-okhttp-3.0")
    implementation("org.apache.logging.log4j:log4j-core")

    runtimeOnly("org.apache.logging.log4j:log4j-slf4j-impl")
}

tasks {
    assemble {
        dependsOn("shadowJar")
    }
}
