plugins {
    id("com.diffplug.spotless")
}

allprojects {
    group = "io.opentelemetry.lambda"

    plugins.apply("com.diffplug.spotless")

    plugins.withId("java") {
        configure<JavaPluginExtension> {
            sourceCompatibility = JavaVersion.VERSION_1_8
            targetCompatibility = JavaVersion.VERSION_1_8
        }

        spotless {
            java {
                googleJavaFormat()
            }
        }

        dependencies {
            afterEvaluate {
                configurations.configureEach {
                    if (!isCanBeResolved && !isCanBeConsumed) {
                        add(name, enforcedPlatform(project(":dependencyManagement")))
                    }
                }
            }
        }
    }
}
