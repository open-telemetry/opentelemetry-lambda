subprojects {
    plugins.withId("java") {
        configure<JavaPluginConvention> {
            sourceCompatibility = JavaVersion.VERSION_1_8
            targetCompatibility = JavaVersion.VERSION_1_8
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
