module bldy.build/build

require (
	bitbucket.org/pkg/inflect v0.0.0-20130829110746-8961c3750a47
	github.com/BurntSushi/toml v0.3.1
	github.com/Microsoft/go-winio v0.4.13-0.20190321200717-c599b533b43b
	github.com/Microsoft/hcsshim v0.8.7-0.20190325164909-8abdbb8205e4
	github.com/bazelbuild/bazel-gazelle v0.0.0-20190402225339-e530fae7ce5c
	github.com/beorn7/perks v0.0.0-20160804104726-4c0e84591b9a
	github.com/blang/semver v3.1.0+incompatible
	github.com/cenkalti/backoff v2.1.1+incompatible

	github.com/containerd/aufs v0.0.0-20181026202817-da3cf16bfbe6
	github.com/containerd/btrfs v0.0.0-20180306195803-2e1aa0ddf94f
	github.com/containerd/cgroups v0.0.0-20190328223300-4994991857f9
	github.com/containerd/console v0.0.0-20180822173158-c12b1e7919c1
	github.com/containerd/containerd v1.2.5
	github.com/containerd/continuity v0.0.0-20181001140422-bd77b46c8352

	github.com/containerd/cri v1.11.1-0.20190125013620-4dd6735020f5
	github.com/containerd/fifo v0.0.0-20180307165137-3d5202aec260
	github.com/containerd/go-cni v0.0.0-20181011142537-40bcf8ec8acd
	github.com/containerd/go-runc v0.0.0-20180907222934-5a6d9f37cfa3
	github.com/containerd/ttrpc v0.0.0-20190107211005-f02858b1457c
	github.com/containerd/typeurl v0.0.0-20180627222232-a93fcdb778cd

	github.com/containerd/zfs v0.0.0-20181026174235-9f6ef3b1fe51
	github.com/containernetworking/cni v0.6.0
	github.com/containernetworking/plugins v0.7.0
	github.com/coreos/go-systemd v0.0.0-20161114122254-48702e0da86b
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker v0.7.3-0.20171019062838-86f080cff091
	github.com/docker/go-events v0.0.0-20170721190031-9461782956ad
	github.com/docker/go-metrics v0.0.0-20180131145841-4ea375f7759c
	github.com/docker/go-units v0.3.1
	github.com/docker/spdystream v0.0.0-20160310174837-449fdfce4d96
	github.com/emicklei/go-restful v2.2.1+incompatible
	github.com/godbus/dbus v0.0.0-20151105175453-c7fdd8b5cd55
	github.com/gofrs/flock v0.7.1
	github.com/gogo/googleapis v1.1.0
	github.com/gogo/protobuf v1.2.1
	github.com/golang/mock v1.1.1
	github.com/golang/protobuf v1.3.1
	github.com/google/btree v1.0.0
	github.com/google/go-cmp v0.1.0
	github.com/google/gofuzz v0.0.0-20161122191042-44d81051d367

	github.com/google/skylark v0.0.0-20181101142754-a5f7082aabed
	github.com/google/subcommands v0.0.0-20181012225330-46f0354f6315
	github.com/google/uuid v1.1.1
	github.com/grpc-ecosystem/go-grpc-prometheus v0.0.0-20160910222444-6b7015e65d36
	github.com/hashicorp/errwrap v0.0.0-20141028054710-7554cd9344ce
	github.com/hashicorp/go-multierror v0.0.0-20161216184304-ed905158d874
	github.com/josharian/impl v0.0.0-20180228163738-3d0f908298c4 // indirect
	github.com/json-iterator/go v1.1.5
	github.com/konsorten/go-windows-terminal-sequences v1.0.1
	github.com/kr/pty v1.1.1
	github.com/mattes/migrate v3.0.1+incompatible
	github.com/matttproud/golang_protobuf_extensions v1.0.0
	github.com/mistifyio/go-zfs v2.1.2-0.20171122051224-166add352731+incompatible
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742
	github.com/opencontainers/go-digest v1.0.0-rc1.0.20180430190053-c9281466c8b2
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v1.0.0-rc7
	github.com/opencontainers/runtime-spec v1.0.1
	github.com/opencontainers/runtime-tools v0.6.0
	github.com/opencontainers/selinux v1.0.0-rc1.0.20180628160156-b6fa367ed7f5
	github.com/pborman/uuid v0.0.0-20180122190007-c65b2f87fee3
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.0-pre1.0.20180131142826-f4fb1b73fb09
	github.com/prometheus/client_model v0.0.0-20171117100541-99fa1f4be8e5
	github.com/prometheus/common v0.0.0-20180110214958-89604d197083
	github.com/prometheus/procfs v0.0.0-20180125133057-cb4147076ac7
	github.com/seccomp/libseccomp-golang v0.0.0-20160531183505-32f571b70023
	github.com/sirupsen/logrus v1.4.1
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2
	github.com/tchap/go-patricia v2.2.6+incompatible
	github.com/urfave/cli v1.20.1-0.20171014202726-7bc6a0acffa5
	github.com/vaughan0/go-ini v0.0.0-20130923145212-a98ad7ee00ec
	github.com/vishvananda/netlink v1.0.0
	github.com/vishvananda/netns v0.0.0-20180720170159-13995c7128cc // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415
	github.com/xeipuuv/gojsonschema v0.0.0-20180618132009-1d523034197f
	go.etcd.io/bbolt v1.3.2
	go.starlark.net v0.0.0-20190402030956-cb523dafc375
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net v0.0.0-20190403144856-b630fd6fe46b
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
	golang.org/x/sync v0.0.0-20190227155943-e225da77a7e6
	golang.org/x/sys v0.0.0-20190303122642-d455e41777fc
	golang.org/x/text v0.3.0
	golang.org/x/time v0.0.0-20180412165947-fbb02b2291d2
	google.golang.org/genproto v0.0.0-20180817151627-c66870c02cf8
	google.golang.org/grpc v1.19.1
	gopkg.in/inf.v0 v0.9.0
	gopkg.in/yaml.v2 v2.2.2
	gotest.tools v2.1.0+incompatible
	gvisor.googlesource.com/gvisor v0.0.0-00010101000000-000000000000
	k8s.io/api v0.0.0-20181204000039-89a74a8d264d
	k8s.io/apimachinery v0.0.0-20181127025237-2b1284ed4c93
	k8s.io/apiserver v0.0.0-20181204001702-9caa0299108f
	k8s.io/client-go v10.0.0+incompatible
	k8s.io/klog v0.0.0-20181108234604-8139d8cb77af
	k8s.io/kubernetes v1.13.0
	k8s.io/utils v0.0.0-20181115163542-0d26856f57b3
	sevki.org/pqueue v0.0.0-20180411183211-725269ba7143
	sevki.org/x v1.0.0
	sigs.k8s.io/yaml v1.1.0
)

replace gvisor.googlesource.com/gvisor => ./third_party/gvisor
