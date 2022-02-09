# -*- mode: Python -*-

# For more on Extensions, see: https://docs.tilt.dev/extensions.html
load('ext://restart_process', 'docker_build_with_restart')
load('ext://local_output', 'local_output')

default_registry('ghcr.io/vasu1124')
allow_k8s_contexts(['colima', 'Default'])

gitVersion   = local_output('cat introspect.VERSION')
gitCommit    = local_output('git rev-parse --verify HEAD')
gitTreeState = local_output('[ -z git status --porcelain 2>/dev/null ] && echo clean || echo dirty')
buildDate    = local_output('date --rfc-3339=seconds | sed "s/ /T/"')

LDFLAGS = """ \
-ldflags '\
-X github.com/vasu1124/introspect/pkg/version.gitVersion=%s \
-X github.com/vasu1124/introspect/pkg/version.gitCommit=%s \
-X github.com/vasu1124/introspect/pkg/version.gitTreeState=%s \
-X github.com/vasu1124/introspect/pkg/version.buildDate=%s' \
""" % (gitVersion, gitCommit, gitTreeState, buildDate)
compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ' + LDFLAGS + ' -o introspect-linux-amd64 ./cmd/'

print(compile_cmd)

local_resource(
  'introspect-compile',
  compile_cmd,
  deps=['./cmd', './pkg', './vendor'],
  labels=['introspect']
)

docker_build_with_restart(
  'ghcr.io/vasu1124/introspect',
  '.',
  entrypoint=['/introspect-linux-amd64'],
  dockerfile='docker/Dockerfile.alpine',
  only=[
    './introspect-linux-amd64',
    './css', 
    './tmpl',
  ],
  live_update=[
    sync('./css', '/css'),
    sync('./tmpl', '/tmpl'),
  ],
)

k8s_yaml(kustomize('./kubernetes/all-in-one'))
k8s_resource(
  'introspect', 
  port_forwards=[9090],  
  resource_deps=['introspect-compile'], 
  labels=['introspect']
)

k8s_resource(workload='introspect', objects=[
  'introspect-config:configmap',
  'introspect-tls:secret',
  'introspect-validationwebook:validatingwebhookconfiguration',
  'uselessmachines.introspect.actvirtual.com:customresourcedefinition',
  'introspect-lease:role',
  'introspect-role:clusterrole',
  'introspect-lease-rolebinding:rolebinding',
  'introspect-rolebinding:clusterrolebinding',],
  labels=['introspect']
)
k8s_resource(workload='mongodb', objects=[
  'mongodb:persistentvolumeclaim',
  'mongodb-secret:secret'],
  labels=['introspect']
)

v1alpha1.extension_repo(name='tilt-extensions', url='https://github.com/tilt-dev/tilt-extensions')
v1alpha1.extension(
  name='ngrok', 
  repo_name='tilt-extensions', 
  repo_path='ngrok',
)
