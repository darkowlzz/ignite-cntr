# ignite-cntr

Tool to build [ignite](https://github.com/weaveworks/ignite) VM image with a
conatiner-runtime and preloaded images. It also helps to run the container
application inside the VM.

By default, containerd is used as the container-runtime inside the VM. Support
for other container-runtime may be added in the future.

## Building VM Image

Build a VM image with preloaded container images:

```console
$ ignite-cntr image vm darkowlzz/ignite-etcd:test --image quay.io/coreos/etcd:v3.4.7
Started build container ignite-cntr-build-7198945983020451624
Starting containerd in the build container...
Creating containerd namespace: ignite...
Waiting for quay.io/coreos/etcd:v3.4.7 image pull to complete....

Created VM application image: darkowlzz/ignite-etcd:test (sha256:1eadb753b7ceabc68e3739bc1cdd32012ce18fb4c7fac94d86b316ee1e08d91a)
```

In the above example, the VM image is preloaded with an etcd container image.
The created image is an ignite compatible image.

Similarly, more images can be passed using the `--image` flag:

```console
$ ignite-cntr image vm darkowlzz/ignite-misc:test -i docker.io/library/busybox:latest -i docker.io/library/alpine:latest
Started build container ignite-cntr-build-2093644811210178439
Starting containerd in the build container...
Creating containerd namespace: ignite...
Waiting for docker.io/library/busybox:latest image pull to complete..
Waiting for docker.io/library/alpine:latest image pull to complete..

Created VM application image: darkowlzz/ignite-misc:test (sha256:8ce9bef42e5007ae5073d0c6d1b7b5c4eaa727104ddc8b425f7a9921bdcff83f)
```

By default, the VM image is based on `darkowlzz/ignite-cntr-base:dev` base
image. This is based on `weaveworks/ignite-ubuntu` with containerd installed.
To use a separate base image pass the image with `--baseImage` flag.

```console
$ ignite-cntr image vm darkowlzz/ignite-etcd:test --image quay.io/coreos/etcd:v3.4.7 --baseImage foo/bar:baz
...
```

## Building VM Base Image

VM base image can be built with the `image base` subcommand:

```console
$ ignite-cntr image base
Building image...
Base image built: darkowlzz/ignite-cntr-base:dev
```

By default, `darkowlzz/ignite-cntr-base:dev` base image is created. It can be
configured to build a different image by passing the name as an argument:

```console
$ ignite-cntr image base foo/bar:baz
Building image...
Base image built: foo/bar:baz
```

The base image of this VM base image can be passed using the `--baseImage` flag.

## Running the container application inside a VM

Before an application can be run, create an ignite VM that contains the
application container image preloaded:

```console
$ sudo ignite run darkowlzz/ignite-etcd:v0.0.1 --cpus 1 --memory 1GB --ssh --name=my-vm
INFO[0001] Created VM with ID "8e608e51e7cc0e92" and name "my-vm" 
INFO[0001] Networking is handled by "cni"               
INFO[0001] Started Firecracker VM "8e608e51e7cc0e92" in a container with ID "ignite-8e608e51e7cc0e92" 

$ sudo ignite ps
VM ID			IMAGE				KERNEL					SIZE	CPUS	MEMORY		CREATED	STATUS	IPS		PORTS	NAME
8e608e51e7cc0e92	darkowlzz/ignite-etcd:v0.0.1	weaveworks/ignite-kernel:4.19.47	4.0 GB	1	1024.0 MB	9s ago	Up 9s	10.61.0.54		my-vm
```

In the above, an ignite VM is created using an ignite-etcd VM image. To run etcd
inside the VM, run:

```console
$ sudo ignite-cntr run my-vm quay.io/coreos/etcd:v3.4.7 --env "ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379" --env "ETCD_ADVERTISE_CLIENT_URLS=http://10.61.0.52:2379" --net-host
CMD: ctr -n ignite container create quay.io/coreos/etcd:v3.4.7 container-app --env ETCD_ADVERTISE_CLIENT_URLS=http://10.61.0.54:2379 --env ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379 --net-host
```

__NOTE__: Like ignite, the `run` subcommand must be run with sudo.

In the above, etcd container is run with some environment variables and host
netorking enabled. The command output shows the command executed to create a
containerd container which is then run with a containerd task.

By default, a containerd namespace called `ignite` is created. The etcd
containerd task can be checked by logging into the VM.

```console
$ sudo ignite ssh my-vm
...
...
root@localhost:~# ctr -n ignite c ls
CONTAINER        IMAGE                         RUNTIME                  
container-app    quay.io/coreos/etcd:v3.4.7    io.containerd.runc.v2    

root@localhost:~# ctr -n ignite t ls
TASK             PID    STATUS    
container-app    969    RUNNING
```

### Container Environment Variables File

Passing environment variables file is supported. In the above example, the etcd
environment variables can be written into a file, say `etcd.env`:

```txt
ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379
ETCD_ADVERTISE_CLIENT_URLS=http://10.61.0.54:2379
```

This file can then be passed to the `run` subcommand as:

```console
$ sudo ignite-cntr run my-vm quay.io/coreos/etcd:v3.4.7 --env-file etcd.env --net-host
CMD: ctr -n ignite container create quay.io/coreos/etcd:v3.4.7 container-app --env ETCD_ADVERTISE_CLIENT_URLS=http://10.61.0.54:2379 --env ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379 --net-host
```

It also supports merging flag based env vars and file based env vars.
