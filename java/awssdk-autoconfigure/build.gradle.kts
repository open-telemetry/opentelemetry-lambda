// TODO(anuraaga): Move this into instrumentation repo

plugins {
    `java-library`
}

base.archivesName = "opentelemetry-lambda-awsdk-autoconfigure"

dependencies {
    compileOnly("io.opentelemetry:opentelemetry-api")
    compileOnly("software.amazon.awssdk:aws-core")

    implementation("io.opentelemetry.instrumentation:opentelemetry-aws-sdk-2.2")
}
