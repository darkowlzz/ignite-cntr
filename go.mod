module github.com/darkowlzz/ignite-cntr

go 1.13

require (
	github.com/fsouza/go-dockerclient v1.6.3
	github.com/joho/godotenv v1.3.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.6.2
	github.com/weaveworks/ignite v0.9.1-0.20210419164134-8b31ad7524bc
	github.com/weaveworks/libgitops v0.0.0-20200611103311-2c871bbbbf0c
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
)

replace github.com/docker/distribution => github.com/docker/distribution v0.0.0-20190711223531-1fb7fffdb266
