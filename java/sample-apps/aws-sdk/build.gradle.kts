plugins {
    java
    id("com.github.johnrengelman.shadow")
}

dependencies {
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
