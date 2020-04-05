module github.com/darkowlzz/ignite-cntr

go 1.13

require (
	github.com/containerd/containerd v1.3.3 // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible // indirect
	github.com/fsouza/go-dockerclient v1.6.3
	github.com/mitchellh/go-homedir v1.1.0
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/spf13/cobra v0.0.7
	github.com/spf13/viper v1.6.2
	github.com/weaveworks/gitops-toolkit v0.0.0-20190830163251-b6682e98e2fa
	github.com/weaveworks/ignite v0.6.3
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.3.0
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20190711223531-1fb7fffdb266
	github.com/weaveworks/ignite => github.com/darkowlzz/ignite v0.6.1-0.20200323175308-232484ec80ee
)
