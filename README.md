OpenAPI2Siege
-------------

A simple tool to convert your OpenAPI spec(s) to one or more Siege URL files.

Features
========

- Parse an OpenAPI spec into a list of URLs, an accompanying cookie file, and a `siege.conf` meant to tie them together.
- Using a configuration file, users can override the parameters and payloads in the spec itself with their own values.
- Generate separate files per media type for use in separate runs (since Siege doesn't support per-URL media types).
- Verbose messages when something can't be converted to Siege's expectations, letting users adjust the results as needed.

Limitations
===========

- **NO OpenAPI v2 SUPPORT YET**
- Payloads are built in a best-effort fashion; it can probably improve
- Some features aren't available in Siege; these are generally flagged on stdout
- Paths are in alphabetical order to ensure they only appear once; this should probably be configurable

Testing
=======

Tested against the current Digital Ocean v2 documentation. YMMV.

Contribute
==========

Please tell us about any bugs or improvements you want to see!
