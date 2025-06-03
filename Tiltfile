# Build
custom_build(
    # Name of the container image
    ref = 'perencanaan-service',

    # Command to build the container image
    command = 'docker build -t $EXPECTED_REF .',

    # Files to watch that trigger a new build
    deps = ['app', 'controller', 'middleware', 'model', 'repository', 'service', 'helper']
)

# Deploy
k8s_yaml(['k8s/deployment.yml', 'k8s/service.yml'])

# Manage
k8s_resource('perencanaan-service', port_forwards=['9002'])
