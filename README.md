# `gopkgr` - Go Package Installer

by Tim Henderson

Packager and installer of source tarballs into a $GOPATH. BSD Licensed.

## Why?

The vision is a tool which can package, install, and manage go code -- both
binary and source. Currently, there is no standard way to package go projects
for distribution (either with dependencies or without). This project aims to
solve that problem. I do not aim at the moment to solve general dependency
management. Rather I hope that "godep" will provide a robust depency management
solution. This project will hopefully provide a solution for package,
installation and managing environments.

# Status

Currently `gopkgr` consists of two things. One is a very simple system to tarball
a src tree from a go path. It can also install such source trees (and remove
them if you have the original tarball). This is very much at the "proof of
concept" stage.

The second item is `goenv` which can manage a virtual environment for your
project. You will have trees in a project managed by `goenv`. The first is your
virtual environment where all your dependencies are installed. This goes under
`project/venv`. The second is your code which goes under `src` as usual. Thus
your environment variables will look like this:

    GOPATH=/path/to/project:/path/to/project/venv
    PATH=/path/to/project/bin:/path/to/project/venv/bin:...(other entries)
    GOENV=/path/to/project/env

The vision is ahead of the tool at the moment. But, I want to begin
experimenting immediately with this concept. For too long we have suffered
without a proper way to package our go code. Hopefully, this system will solve
that issue.

# Usage

    go get github.com/timtadh/gopkgr
    eval $(gopkgr --goenv-function) ## add this to your .bashrc

Making a virtual environment (or activating one)

    goenv activate /path/to/project

Deactivating the virtual environment

    goenv deactivate

Tarballing a source tree:

    gopkgr mkpkg -o mytree.tar.gz /path/to/gopath

Installing a tarball (2 ways):

    gopkgr install /path/to/gopath mytree.tar.gz
    goenv install mytree.tar.gz

Removing a tarball (2 ways):

    gopkgr remove /path/to/gopath mytree.tar.gz
    goenv remove mytree.tar.gz

Making a tarball from a go-gettable url

    goenv getpkg -o pkg.tar.gz github.com/username/repo
    tar tzf pkg.tar.gz ## check that it is what you want

Make and install a tarball from a go-gettable url. This will put the
tarball into `<projects>/deps`. It will be named
`echo <url>.tar.gz | sed 's/\//-/g'`

    goenv get github.com/username/repo

# Comments

Please let me know via email or github issues. There are many improvements I
want to make. In no particular order:

- track what files are associated with what package
- package manifests
- universal binary distributions (fully cross compiled)
- integration with `godep` or similar dependency tracking system
- getting dependencies that are not distributed in the package
- packages with seperate section for dependencies (not merged with project code)
- support both projects that are developed root at src/... and go gettable
  projects
- (your idea here!)

