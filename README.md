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

Original project felt like an orphan and I needed some additional functionality. In addition latest PRs look unnecessary to me. Result should be fully compatible with previous releases - if not, please open issue here. 

Changes
----------------

* Code modernization/simplification/refactoring (go modules and latest compiler, etc).
* In order for "open" call to work properly when sending back file ("trans-localfile") additional information needs to be transferred because SSH port forwarding has been chosen for security and actual remote address is never available.

I attempted to support backward compatibility as much as I could leaving argument processing mostly unchanged. Everywhere possible I switched code to go stdlib minimizing external dependencies.
