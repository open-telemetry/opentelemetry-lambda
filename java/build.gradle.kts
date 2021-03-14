subprojects {
    plugins.withId("java") {
        configure<JavaPluginConvention> {
            sourceCompatibility = JavaVersion.VERSION_11
            targetCompatibility = JavaVersion.VERSION_11
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
