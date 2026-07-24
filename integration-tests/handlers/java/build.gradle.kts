plugins {
    java
    id("com.gradleup.shadow") version "9.6.1"
}

repositories {
    mavenCentral()
}

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(17)
    }
}

dependencies {
    implementation(platform("software.amazon.awssdk:bom:2.49.2"))
    implementation("software.amazon.awssdk:sts")
    implementation("com.amazonaws:aws-lambda-java-core:1.4.0")
    implementation("com.amazonaws:aws-lambda-java-events:3.16.1")
}

tasks {
    shadowJar {
        archiveBaseName = "handler"
        archiveClassifier = "all"
        archiveVersion = ""
    }

    assemble {
        dependsOn("shadowJar")
    }
}
