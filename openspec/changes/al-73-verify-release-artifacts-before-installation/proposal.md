# AL-73 release artifact verification

Installation should admit only the selected release bytes, and a release should pass local verification only when every selected platform has its archive, manager and engine. The [roadmap ticket](../../../alspaces/plan/to-do/AL-73%20verify%20release%20artifacts%20before%20installation.md) records the historical failures.

Current installers already request exact tags and check selected digests. This change closes gaps in bootstrap checksum failure handling, remote repository selection and release inventory verification. It also verifies that a failed installation reports the components actually installed. Both the contributor entry point and the installed setup skill retain the canonical downloader and native manager. The existing user command location, cache layout, platform defaults and hook assets stay supported. The selected behavior is defined in the [capability specification](specs/release-artifact-verification/spec.md).
