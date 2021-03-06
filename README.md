<p align="center">
    <h1 align="center">lemonade</h1>
    <p align="center">
		Lemonade is a remote utility tool - copy, paste and open browser over TCP.
    </p>
    <p align="center">
        <a href="https://godoc.org/github.com/rupor-github/lemonade"><img alt="GoDoc" src="https://img.shields.io/badge/godoc-reference-blue.svg" /></a>
        <a href="https://goreportcard.com/report/github.com/rupor-github/lemonade"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/rupor-github/lemonade" /></a>
    </p>
    <hr>
</p>

Installation
------------

### From binaries

Download from the [releases page](https://github.com/rupor-github/lemonade/releases) and unpack it in a convenient location.

### From source

I am using `cmake` to build it from sources, included `CMakeList.txt` expects host platform to be up to date `linux` with go 1.13 or later installed. See `build-all.sh` for example of cross compile for all supported platforms.

Configuration and examples
----------------

Please, see original project [here](https://github.com/lemonade-command/lemonade)

Reason for forking
----------------

Original project felt like an orphan and I needed some additional functionality. In addition latest PRs look unnecessary to me. Result should be fully compatible with previous releases - if not, please open issue here.

For Windows users: lemonade "server" backend is included as part of [wsl-ssh-agent](https://github.com/rupor-github/wsl-ssh-agent) and could be run as an icon in the taskbar notification area.

Changes
----------------

* Code modernization/simplification/refactoring (go modules and latest compiler, etc).
* In order for "open" call to work properly when sending back file ("trans-localfile") additional address information needs to be transferred because SSH port forwarding has been chosen for security and actual remote address is never available.
* In order to avoid ssh channel errors when dynamic port forwarding is used to get file we need to handle multiple browser connections from "server" end.
* The idea of local fallback is really unclear to me, since settings default to localhost anyways.

I attempted to support **backward compatibility** as much as I could, leaving argument processing unchanged (just adding some new aguments with defaults). Everywhere possible I switched code to go stdlib trying to minimize dependencies.

## Credit

* Thanks to [Masataka Pocke Kuwabara](https://github.com/pocke) for original [lemonade](https://github.com/lemonade-command/lemonade)

---------------------------------------------------------------------------------------------------------------------------------------

Licensed under MIT license.

This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
