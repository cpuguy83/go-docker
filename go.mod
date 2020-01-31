module github.com/cpuguy83/go-docker

go 1.13

require (
	github.com/docker/go-units v0.4.0
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/pkg/errors v0.9.1
	gotest.tools v2.2.0+incompatible
)

replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20191113042239-ea84732a7725
