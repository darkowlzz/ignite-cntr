module github.com/darkowlzz/ignite-cntr

go 1.13

require (
	github.com/docker/docker v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible // indirect
	github.com/fsouza/go-dockerclient v1.6.3
	github.com/joho/godotenv v1.3.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.6.2
	github.com/weaveworks/ignite v0.7.1-0.20201005162233-65c55dd258b7
	github.com/weaveworks/libgitops v0.0.0-20200611103311-2c871bbbbf0c
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20190711223531-1fb7fffdb266
	github.com/docker/docker => github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200121204235-bf4fb3bd569c
)
