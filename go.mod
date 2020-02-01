module gitlab.com/greut/eclint

go 1.13

replace github.com/editorconfig/editorconfig-core-go/v2 => github.com/greut/editorconfig-core-go/v2 v2.0.0-20200201103819-93b20009932a

require (
	github.com/editorconfig/editorconfig-core-go/v2 v2.2.2
	github.com/go-logr/logr v0.1.0
	github.com/gogs/chardet v0.0.0-20191104214054-4b6791f73a28
	github.com/logrusorgru/aurora v0.0.0-20200102142835-e9ef32dff381
	github.com/mattn/go-colorable v0.1.4
	golang.org/x/crypto v0.0.0-20200128174031-69ecbb4d6d5d
	k8s.io/klog/v2 v2.0.0-20200127113903-12be8a0d907a
)
