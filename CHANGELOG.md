Changelog
=========

## 0.0.5 in-progress

*   Refactor logging a bit
*   Move encoding functionality into a submodule to reduce import size (and
    thus resultant binary size) for url-tool

## 0.0.4 2012-09-02

*   Refactor Stats code out of camoproxy
*   Make stats an optional flag in go-camo
*   Minor documentation cleanup
*   Clean up excessive logging on client initiated broken pipe

## 0.0.3 2012-08-05

*   organize and clean up code
*   make header filters exported 
*   start filtering response headers
*   add default Server name
*   fix bug dealing with header filtering logic
*   add cli utility to encode/decode urls for testing, etc.
*   change repo layout to be friendlier for Go development/building
*   timeout flag is now a duration (15s, 10m, 1h, etc)
*   X-Forwarded-For support
*   Added more info to readme


## 0.0.2 2012-07-12

*   documentation cleanup
*   code reorganization
*   some cleanup of command flag text
*   logging code simplification


## 0.0.1 2012-07-07

*   initial release
