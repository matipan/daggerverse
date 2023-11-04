# Gradle

Dagger module for integrating with applications using gradle as a build system. The goal of the module is to provide the integrated gradle tasks (such as `build` and `test`) out of the box and to let users use the custom `task` function to call custom tasks, whether they are user defined or plugin based.

Features:
* Automatically mount .gradle caches
* Specify custom gradle image version

TODO:
* [ ] Validate gradle versions when specified in `with-version`
* [ ] Support for using `gradlew` when available instead of the gradle command. This would mean that `with-version` won't be necessary when using the wrapper. Maybe `with-wrapper` with a path to the wrapper and have `.` as a default?
* [ ] Potentially integrate tasks for known plugins such as `bootRun`? Not sure if a Dagger module should be that opinionated, but maybe there is a way of extending it a bit more. We could create a type for a `task` and have pre-defined values for `cmd` and `args`
