Lemonade
========

Lemonade is a remote utility tool - copy, paste and open browser over TCP.

Installation
------------

Configuration and examples
----------------

Please, see original project [here](https://github.com/lemonade-command/lemonade)

Reason for forking
----------------

Original project after v1.1.1 was mostly unmaintained.

Changes
----------------

* Code modernization/simplification/refactoring
* Several connection leaks have been detected and removed and various serving components changed slightly.
* In order for "open" call to work properly when sending back ("trans-localfile") actual file different information needs to be transferred because SSH port forwarding has been chosen for security and actual remote address is not available.

I left command line processing mostly unchanged to provide backward compatibility with old releases, only help printouts have been changed. Everywhere possible I switched code back to go stdlib minimizing dependencies.
