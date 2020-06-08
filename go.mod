module github.com/darkowlzz/ignite-cntr

go 1.13

require (
	github.com/docker/docker v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible // indirect
	github.com/fsouza/go-dockerclient v1.6.3
	github.com/joho/godotenv v1.3.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v0.0.7
	github.com/spf13/viper v1.6.2
	github.com/weaveworks/gitops-toolkit v0.0.0-20200410161308-f0fc148681c0
	github.com/weaveworks/ignite v0.6.3
	golang.org/x/crypto v0.0.0-20200406173513-056763e48d71
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20190711223531-1fb7fffdb266
	github.com/docker/docker => github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
	github.com/weaveworks/ignite => github.com/darkowlzz/ignite v0.6.1-0.20200608183646-d601c14af898
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c
)
