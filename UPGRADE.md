# Upgrading

Please refer to [CHANGELOG.md](./CHANGELOG.md) for a full list of changes.

The intent of this document is to make migration of breaking changes as easy as possible. Please note that not all
breaking changes might be included here. Refer to refer to [CHANGELOG.md](./CHANGELOG.md) for a full list of changes
before finalizing the upgrade process.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# 0.1.0

* Refactored judge, jury, juror
  * Can add custom jurors
  * Also works with regular token introspection
    * But needs scope
    * TODO: Refresh token identification
* Compatibility with Keto
* How to make compatible with hydra 0.11.0
* needs sql migration

* rename the commands
  * management -> api
  * deprecated all (?)
    * MANAGEMENET_ and PROX_ deprecated due to all
    * Does not need database_url any more but ANONYMOUS_SUBJECT_ID

* Added ANONYMOUS_SUBJECT_ID

* Warden policies now possible with anonymous users

* Scopes are now being enforced in introspect