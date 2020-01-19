module gitlab.com/greut/eclint

go 1.13

replace github.com/editorconfig/editorconfig-core-go/v2 => github.com/greut/editorconfig-core-go/v2 v2.0.0-20200119154652-31109b0550dd

require (
	github.com/editorconfig/editorconfig-core-go/v2 v2.2.2
	github.com/go-logr/logr v0.1.0
	github.com/gogs/chardet v0.0.0-20191104214054-4b6791f73a28
	github.com/logrusorgru/aurora v0.0.0-20200102142835-e9ef32dff381
	github.com/mattn/go-colorable v0.1.4
	golang.org/x/crypto v0.0.0-20200117160349-530e935923ad
	k8s.io/klog/v2 v2.0.0-20200108022340-c4f748769d6e
)
