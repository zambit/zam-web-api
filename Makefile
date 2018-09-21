proj_pkg_path=git.zam.io/wallet-backend/web-api
proj_vendor_pkg_pref=$(proj_pkg_path)/vendor

gitver_pkg_path=git.zam.io/wallet-backend/common/pkg/gitversion
gitver_vendor_pkg_path=$(proj_vendor_pkg_pref)/$(gitver_pkg_path)

LDFLAGS = -X $(gitver_pkg_path).RawCommitSHA=$(shell git rev-parse HEAD)
LDFLAGS +=  -X $(gitver_vendor_pkg_path).RawCommitSHA=$(shell git rev-parse HEAD)

LDFLAGS +=  -X $(gitver_pkg_path).RawCommitRev=$(shell git describe --abbrev=0 --tags)
LDFLAGS +=  -X $(gitver_vendor_pkg_path).RawCommitRev=$(shell git describe --abbrev=0 --tags)

LDFLAGS +=  -X $(gitver_pkg_path).RawCommitBranch=$(shell git rev-parse --abbrev-ref HEAD)
LDFLAGS +=  -X $(gitver_vendor_pkg_path).RawCommitBranch=$(shell git rev-parse --abbrev-ref HEAD)

LDFLAGS +=  -X $(gitver_pkg_path).RawBuildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS +=  -X $(gitver_vendor_pkg_path).RawBuildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

LDFLAGS +=  -X $(gitver_pkg_path).RawOriginUrl=$(shell git config --get remote.origin.url)
LDFLAGS +=  -X $(gitver_vendor_pkg_path).RawOriginUrl=$(shell git config --get remote.origin.url)

ifeq (${CI_ENVIRONMENT_NAME},)
else
    env = $(CI_ENVIRONMENT_NAME)
    pipID = $(CI_PIPELINE_ID)

    LDFLAGS +=  -X $(gitver_pkg_path).RawCIEnv=$(env)
    LDFLAGS +=  -X $(gitver_vendor_pkg_path).RawCIEnv=$(env)
    LDFLAGS +=  -X $(gitver_pkg_path).RawCIPipID=$(pipID)
    LDFLAGS +=  -X $(gitver_vendor_pkg_path).RawCIPipID=$(pipID)
endif

ifdef out
else
    out=web-api
endif

# Build the project
all: test build

build:
	go build -ldflags="${LDFLAGS}" -o $(out) cmd/main/main.go

test:
	WA_MIGRATIONS_DIR=$(shell pwd)/db/migrations ginkgo -r .

clean:
	rm -f $BINARY

%:
	@:

.PHONY: build test clean
