#!/usr/bin/env python


# http://stackoverflow.com/a/14050282
def check_antipackage():
    from sys import version_info
    sys_version = version_info[:2]
    found = True
    if sys_version < (3, 0):
        # 'python 2'
        from pkgutil import find_loader
        found = find_loader('antipackage') is not None
    elif sys_version <= (3, 3):
        # 'python <= 3.3'
        from importlib import find_loader
        found = find_loader('antipackage') is not None
    else:
        # 'python >= 3.4'
        from importlib import util
        found = util.find_spec('antipackage') is not None
    if not found:
        print('Install missing package "antipackage"')
        print('Example: pip install git+https://github.com/ellisonbg/antipackage.git#egg=antipackage')
        from sys import exit
        exit(1)
check_antipackage()

# ref: https://github.com/ellisonbg/antipackage
import antipackage
from github.appscode.libbuild import libbuild, pydotenv

import os
import os.path
import subprocess
import sys
import yaml
from os.path import expandvars, join, dirname

libbuild.REPO_ROOT = libbuild.GOPATH + '/src/github.com/appscode/guard'
BUILD_METADATA = libbuild.metadata(libbuild.REPO_ROOT)
libbuild.BIN_MATRIX = {
    'guard': {
        'type': 'go',
        'go_version': True,
        'distro': {
            'alpine': ['amd64'],
            'darwin': ['amd64'],
            'linux': ['amd64'],
        }
    }
}
if libbuild.ENV not in ['prod']:
    libbuild.BIN_MATRIX['guard']['distro'] = {
        'alpine': ['amd64']
    }
libbuild.BUCKET_MATRIX = {
    'prod': 'gs://appscode-cdn',
    'dev': 'gs://appscode-dev'
}


def call(cmd, stdin=None, cwd=libbuild.REPO_ROOT):
    print(cmd)
    return subprocess.call([expandvars(cmd)], shell=True, stdin=stdin, cwd=cwd)


def die(status):
    if status:
        sys.exit(status)


def check_output(cmd, stdin=None, cwd=libbuild.REPO_ROOT):
    print(cmd)
    return subprocess.check_output([expandvars(cmd)], shell=True, stdin=stdin, cwd=cwd)


def version():
    # json.dump(BUILD_METADATA, sys.stdout, sort_keys=True, indent=2)
    for k in sorted(BUILD_METADATA):
        print(k + '=' + BUILD_METADATA[k])


def fmt():
    libbuild.ungroup_go_imports('*.go', 'auth', 'commands', 'docs', 'installer', 'server', 'test', 'util')
    die(call('goimports -w *.go auth commands docs installer server test util'))
    call('gofmt -s -w *.go auth commands docs installer server test util')


def vet():
    call('go vet *.go ./...')


def lint():
    call('golint *.go ./...')


def gen():
    die(call('go generate auth/providers/ldap/authchoice.go'))


def build_cmd(name):
    cfg = libbuild.BIN_MATRIX[name]
    if cfg['type'] == 'go':
        if 'distro' in cfg:
            for goos, archs in cfg['distro'].items():
                for goarch in archs:
                    libbuild.go_build(name, goos, goarch, main='*.go')
        else:
            libbuild.go_build(name, libbuild.GOHOSTOS, libbuild.GOHOSTARCH, main='*.go')


def build_cmds():
    gen()
    for name in libbuild.BIN_MATRIX:
        build_cmd(name)


def build(name=None):
    if name:
        cfg = libbuild.BIN_MATRIX[name]
        if cfg['type'] == 'go':
            gen()
            build_cmd(name)
    else:
        build_cmds()


def install():
    die(call('GO15VENDOREXPERIMENT=1 ' + libbuild.GOC + ' install ./...'))


def default():
    gen()
    fmt()
    die(call('GO15VENDOREXPERIMENT=1 ' + libbuild.GOC + ' install ./...'))


def test(type, *args):
    die(call(libbuild.GOC + ' install ./...'))

    if os.path.exists(libbuild.REPO_ROOT + "/hack/configs/.env"):
        print 'Loading env file'
        pydotenv.load_dotenv(libbuild.REPO_ROOT + "/hack/configs/.env")

    if type == 'unit':
        unit_test(args)
    elif type == 'e2e':
        e2e_test(args)
    else:
        print '{test unit|e2e}'

def unit_test(args):
    st = ' '.join(args)
    die(call(libbuild.GOC + ' test -v . ./auth/... ./commands/... ./installer/... ./server/... ./util/...' + st))

def e2e_test(args):
    st = ' '.join(args)
    die(call(libbuild.GOC + ' test -v ./test/e2e/... -timeout 10h -args -ginkgo.v -ginkgo.progress -ginkgo.trace -v=3 ' + st))


if __name__ == "__main__":
    if len(sys.argv) > 1:
        # http://stackoverflow.com/a/834451
        # http://stackoverflow.com/a/817296
        globals()[sys.argv[1]](*sys.argv[2:])
    else:
        default()
