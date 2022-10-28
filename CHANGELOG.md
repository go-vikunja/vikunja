# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

All releases can be found on https://code.vikunja.io/api/releases.

## [0.20.0] - 2022-10-28

### Bug Fixes

* *(caldav)* Make sure duration and due date follow rfc5545
* *(caldav)* No failed login emails for tokens (#1252)
* *(ci)* Make sure release zip files have a .zip ending
* *(ci)* Make sure release os packages are properly named
* *(docs)* Clarify using port 25 as mail port when mail does not work
* *(docs)* Document pnpm instead of yarn
* *(docs)* Fix redirect_url example (#50)
* *(lists)* Return correct max right for lists where the user has created the namespace
* *(mail)* Pass mail server timeout (#1253)
* *(migration)* Properly parse duration
* *(migration)* Expose ticktick migrator to /info
* *(migration)* Make sure importing works when the csv file has errors and don't try to parse empty values as dates
* *(namespaces)* Add list subscriptions (#1254)
* *(todoist)* Properly import all done tasks* Properly log extra message ([c194797](c19479757a20d72484b4e071b45055746ff2b67e))
* Don't try to compress riscv64 binaries in releases ([d8f387f](d8f387f7967ffb94035de2fcfc4578247ae1023e))
* Preserve dates for repeating tasks (#47) ([090c671](090c67138a16258480b866b05c6fdc2e02d12c89))
* Tasks with the same assignee as doer should not appear twice in overdue task mails ([45defeb](45defebcf435cade4b72763236e1e2dfdac770cc))
* Don't allow setting a list namespace to 0 ([96ed1e3](96ed1e33e38beec1bb1ab0813074b035dd02fade))
* Make sure pseudo namespaces and lists always have the current user as owner ([878d19b](878d19beb81869392e33a8ffc1ec247d1cf1e4d6))
* Use connection string for postgres ([fcb205a](fcb205a842a4e828e6e933339b23f5aa8b297125))
* Make sure user searches are always case-insensitive ([c076f73](c076f73a87bc9b39b17389e25d0186ab71aa24bf))
* Make cover image id actually updatable ([0e1904d](0e1904d50b8576a2e9ea5812314aa3c8f304edb5))
* Make cover image id actually updatable ([0eb4709](0eb47096db02ceb5032c7439b3b901fbadd0d1bb))
* Make sure a user can only be assigned once to a task ([008908e](008908eb49eeb50a554c416422feb3b465efa165))
* Make sure list subscriptions are set correctly when their namespace has a subscription already ([2fc690a](2fc690a783f5b702fad71da627aa616017727f56))


### Dependencies

* *(deps)* Update klakegg/hugo docker tag to v0.101.0
* *(deps)* Update golang.org/x/sync digest to 8fcdb60
* *(deps)* Update golang.org/x/oauth2 digest to f213421
* *(deps)* Update module src.techknowlogick.com/xgo to v1.5.0+1.19
* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.4.0
* *(deps)* Update golang.org/x/image digest to e7cb969
* *(deps)* Update golang.org/x/term digest to 7a66f97
* *(deps)* Update module github.com/lib/pq to v1.10.7
* *(deps)* Update module github.com/spf13/viper to v1.13.0 (#1260)
* *(deps)* Update dependency golang to v1.19 (#1228)
* *(deps)* Update module github.com/wneessen/go-mail to v0.2.8 (#1258)
* *(deps)* Update module github.com/yuin/goldmark to v1.5.2 (#1261)
* *(deps)* Update module src.techknowlogick.com/xormigrate to v1.5.0 (#1262)
* *(deps)* Update module github.com/magefile/mage to v1.14.0 (#1259)
* *(deps)* Update module github.com/swaggo/swag to v1.8.6 (#1243)
* *(deps)* Update module github.com/wneessen/go-mail to v0.2.9 (#1264)
* *(deps)* Update dependency klakegg/hugo to v0.102.3 (#1265)
* *(deps)* Update module github.com/getsentry/sentry-go to v0.14.0 (#1266)
* *(deps)* Update module github.com/labstack/gommon to v0.4.0 (#1269)
* *(deps)* Update golang.org/x/crypto digest to 4161e89 (#1268)
* *(deps)* Update golang.org/x/oauth2 digest to b44042a (#1270)
* *(deps)* Update golang.org/x/sys digest to 84dc82d (#1271)
* *(deps)* Update dependency klakegg/hugo to v0.104.2 (#1267)
* *(deps)* Update golang.org/x/crypto digest to d6f0a8c (#1275)
* *(deps)* Update golang.org/x/sys digest to 090e330 (#1276)
* *(deps)* Update module github.com/spf13/cobra to v1.6.0 (#1277)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.0 (#1278)
* *(deps)* Update golang.org/x/crypto digest to 56aed06 (#1280)
* *(deps)* Update golang.org/x/text to v0.3.8
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.1 (#1281)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.9.1 (#1282)
* *(deps)* Update golang.org/x/sys digest to 95e765b (#1283)
* *(deps)* Update golang.org/x/oauth2 digest to 6fdb5e3 (#1284)
* *(deps)* Update golang.org/x/image digest to ffcb3fe (#1288)
* *(deps)* Update module golang.org/x/sync to v0.1.0 (#1291)
* *(deps)* Update module github.com/swaggo/swag to v1.8.7 (#1290)
* *(deps)* Update golang.org/x/term digest to 8365914 (#1289)
* *(deps)* Update module github.com/coreos/go-systemd/v22 to v22.4.0 (#1287)
* *(deps)* Update module golang.org/x/oauth2 to v0.1.0 (#1296)
* *(deps)* Update module golang.org/x/crypto to v0.1.0 (#1295)
* *(deps)* Update module golang.org/x/image to v0.1.0 (#1293)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.2 (#1297)
* *(deps)* Update module github.com/stretchr/testify to v1.8.1 (#1298)
* *(deps)* Update module github.com/spf13/cobra to v1.6.1 (#1299)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.3 (#1300)
* *(deps)* Update module github.com/wneessen/go-mail to v0.3.4 (#1302)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.16 (#1301)

### Features

* *(docs)* Add docs about how to deploy Vikunja in a subdirectory
* *(docs)* Document pnpm (#1251)
* *(migration)* Add TickTick migrator
* *(migration)* Add routes for TickTick migrator
* *(migration)* Generate swagger docs
* *(task)* Add cover image attachment id property
* *(task)* Add cover image attachment id property (#1263)* Add sponsor to readme (relm) ([f814dd0](f814dd03eb7f1ae08ea67ae0e3e89b8b4e684ce3))
* Upgrade xorm ([b1fd13b](b1fd13bbcbc551d1bbfe78d91fe6209369709df5))
* Upgrade xorm ([4323803](4323803fd6801e21121eac0d9f9cd62879f090f7))
* Upgrade xorm (#1197) ([5341918](53419180be386d675b4513e7ec70aca85b5ac99b))
* Add github issue templates ([9c4bb5a](9c4bb5a24429dec686e3ccdcd2b920ce5528031c))
* Remove gitea issue template so that only the form is used ([ce621ee](ce621ee5d6b47a0776628073bbd53312a97d116b))
* Add gitea issue template ([0612f4d](0612f4d0e03fbe85018f51056c4833557e78cd3f))
* Provide default user settings for new users via config ([5a40100](5a40100ac5be33d2cbce3c25e355d4036b9b4d3f))
* Add proper checks and errors to see if an attachment belongs to the task it's being used as cover image in ([631a265](631a265d2de9a6196faf28574023fc3cdcc0bfc7))
* Allow a user to remove themselves from a team ([b8769c7](b8769c746ceddc9818f91d6a8a404293ea2e837e))
* TickTick migrator (#1273) ([df2e36c](df2e36c2a378d4bd1b81d959da180b6e9b9a37b9))


### Miscellaneous Tasks

* Upgrade echo ([86ee827](86ee8273bce36c7b4767a34e0d878d63b37ea1b4))
* Go mod tidy ([903b8ff](903b8ff43871234f41f706d571ee2caaba5f4232))
* Generate swagger docs ([e113fe3](e113fe34d074f698f4b0cb237821f359976daa5c))
* Remove unused dependencies ([f5fd849](f5fd849a0b93ff3bba53ac4907bb3fb04fa8692b))

## [0.19.2] - 2022-08-17

### Bug Fixes

* Don't fail a migration if there is no filter saved ([10ded56](10ded56f6697ef47910ec68d37f26ed47cbe9180))
* Don't override saved filters ([beb4d07](beb4d07cf95fc25f7cc5f7471b46bdab49f95fe0))

## [0.19.1] - 2022-08-17

### Bug Fixes

* Prevent moving a list into a pseudo namespace ([3ccc636](3ccc6365a6892f37ee54b0750a34a61e52f6dba1))
* Make sure generating blur hashes for bmp, tiff and webp images works ([8bf0f8b](8bf0f8bb571ddff69a7142be1acaa2e4e0c38e3b))
* Add debian-based docker image for arm 32 builds ([c9e044b](c9e044b3ad60d25e9641d22d84571a7db83a26ac))
* Only list all users when allowed ([9ddd7f4](9ddd7f48895f508539d591aeebde450a86987024))
* Lint ([0c8bed4](0c8bed4054649de8510e5a636d1a14b65d52c402))

### Dependencies

* *(deps)* Update golang.org/x/sys digest to 6e608f9 (#1229)
* *(deps)* Update golang.org/x/sync digest to 886fb93 (#1221)
* *(deps)* Update golang.org/x/sys digest to 8e32c04 (#1230)
* *(deps)* Update golang.org/x/term digest to a9ba230 (#1222)
* *(deps)* Update module github.com/prometheus/client_golang to v1.13.0
* *(deps)* Update module github.com/prometheus/client_golang to v1.13.0 (#1231)
* *(deps)* Update golang.org/x/sys digest to 1c4a2a7
* *(deps)* Update golang.org/x/oauth2 digest to 128564f (#1220)
* *(deps)* Update golang.org/x/image digest to 062f8c9 (#1219)
* *(deps)* Update golang.org/x/crypto digest to 630584e (#1218)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.8.0 (#1233)
* *(deps)* Update golang.org/x/sys digest to fbc7d0a (#1234)
* *(deps)* Update module github.com/wneessen/go-mail to v0.2.6 (#1235)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.15 (#1238)

### Features

* *(docs)* Add k8s docs* Add openid examples ([dbb0f54](dbb0f5473269fb29c4a484cd233a5b76484c4ca7))
* Search by assignee username instead of id ([7f28865](7f28865903740d6dde15ee005323fbdee3072166))
* Add migration to change user ids to usernames in saved filters ([3047ccf](3047ccfd4af8fee55d9ebff49138911ab80cb3d2))

## [0.19.0] - 2022-08-03

### Bug Fixes

* *(caldav)* Make sure the caldav tokens of non-local accounts are properly checked
* *(caldav)* Properly parse durations when returning VTODOs
* *(caldav)* Make sure description is parsed correctly when multiline
* *(ci)* Sign drone config
* *(ci)* Make sure the linter actually runs
* *(ci)* Install git in lint step
* *(docker)* Switch to debian base image
* *(docker)* Use official go image instead of our own to build
* *(docs)* Update minimum required go version
* *(docs)* Use up-to-date hugo image for building
* *(docs)* Don't use cannonify url
* *(docs)* Image urls in synology setup explanation
* *(docs)* Clarify frontend requirements to use Vikunja
* *(dump)* Don't try to save a config file if none was provided and dump vikunja env variables
* *(mage)* Handle different types of errors
* *(mail)* Don't set a username by default
* *(mail)* Don't try to authenticate against the mail server when no credentials are provided
* *(mail)* Set server name in tls config so that sending mail works with skipTlsVerify set to false
* *(restore)* Properly decode notifications json data
* *(restore)* Make sure to reset sequences after importing a dump when using postgres
* *(restore)* Use the correct initial migration* Generate swagger docs ([4de8ec5](4de8ec56a62caef22c2061376383de1fe53ca4c3))
* Make sure the full task is available in notifications ([c2b6119](c2b6119434e6e806785d2c259c3ca3d25496ec75))
* Don't try to load the namespace of a list if it is a shared list ([d7e47a2](d7e47a28d4bb04d4c7c3ed85a263134180da447a))
* Correctly load and pass the user when deleting it ([50b65a5](50b65a517da6869dc6a48fec40323e254ba4c032))
* Updating a list might remove its background ([cf05de1](cf05de19b317bd99c30de4c6a149a0d8a4ff4f49))
* Sorting for saved filters ([57e5d10](57e5d10eee4c45a04e9e1aaeaf41dd44eb8ce788))
* Importing trello attachments ([c3e0e64](c3e0e6405a634894a30dbf9c0506d1691ae4d443))
* Lint ([0b77625](0b7762590f6a0a82090ef74e9e7e32b37142d343))
* Deleting users with no namespaces ([f8a0a7e](f8a0a7e9539a44b2f790a08eb1b03028b56eaac3))
* Importing tasks from todoist without a due time set ([fd0d462](fd0d462bf4dd8225c67ba34958e5148f6167d264))
* User deletion never happens ([72d3c54](72d3c54efd3dda6ae846a069415688391cb1c9ae))
* User deletion reminder emails counting up ([f581885](f581885e65ada15439ec02f1d18d825b03581523))
* User not actually deleted ([70e005e](70e005e7ce5cf1dd25ec9ddfde3cfbbd258fadb6))
* User deletion schedule ([5c88dfe](5c88dfe88eab442724f22c3b29741e78939deae2))
* Friendly name not getting synced on first login from openid ([190a9f2](190a9f2a4c1a59bc68b839c465bb2536532c0e96))
* Importing archived lists or namespaces ([8bb3f8d](8bb3f8d37c78dc704ff4316c750e143528151b48))
* Lint ([a31086a](a31086a7a9ca7723f61a826bccbea125243478f1))
* Microsoft todo migration not importing all tasks ([43f1daf](43f1daf40c388a0aa40f7fd6a8db4c78308d4efd))
* Clarify which config file is used on startup ([44aaf0a](44aaf0a4eccebb1d1a25f5563e928bd1bb82d351))
* Disabling logging completely now works ([22e3f24](22e3f242a396aa9cf54e9426077816f97a0da36f))
* Restoring dumps with no config file saved in them ([8bf2254](8bf2254f4b87446ab0a39080cb0b7d32ccec7c0a))
* Validate email address when creating a user via cli ([75f74b4](75f74b429eea7ae3a75cb10def1ca658af35086a))
* Checking for error types ([ac6818a](ac6818a4769a162c458553944509fe64357370f9))
* Lint ([7fa0865](7fa086518800243385d8cc4696eeea9bf093e5b3))
* Return BlurHash in unsplash search results ([6b51fae](6b51fae0931308464038f55b25e81e68d014c49c))
* Go mod tidy ([e19ad11](e19ad1184662dc9ac9aa89a44abdffc091e2a1b8))
* Decoding images for blurHash generation ([d3bdafb](d3bdafb717b1ad3e2165097ef0b0c2dd47e1502e))
* Lint ([de97fcb](de97fcbd121b1d56b74175fd79ef594ef34e71c8))
* Broken link (#27) ([96e519e](96e519ea96c9537222d0b455037e11fbe9660c31))
* Add more methods to figure out the current binary location ([9845fcc](9845fcc1708431f8f736d36e7e19a1067b0e0e52))
* Set derived default values only after reading config from file or env ([f5ebada](f5ebada91351faf1e5602f0260908defaaabd810))
* Sort tasks logically and consistent across dbms (#1177) ([e52c45d](e52c45d5aabb74ea7b472e8d5b44491cdd7e9489))
* VIKUNJA_SERVICE_JWT_SECRET should be VIKUNJA_SERVICE_JWTSECRET (#1184) ([172a621](172a6214d7c30278017129b950339c78a6ddb7bc))
* Add missing migration ([d837f8a](d837f8a6248b5ff2700a4bfc300d7f9d466cb918))
* Revert renaming Attachments to Embeds everywhere ([c62e26b](c62e26b6fe9d9f362fcfb1df2d5664d7f6854c31))
* Set the correct go version in go.mod ([bc7f6a8](bc7f6a858693b0e61fff7d03b5c2b40b6ae1a55d))
* Go mod tidy ([7a30294](7a30294407843693f6c3a7414b3b9d7093359194))
* Tests ([d0e09d6](d0e09d69d048e62ee7c5b666c2f56761b03e68e6))
* Go mod tidy ([951d74b](951d74b272b1e881faa10095f47b6598bb076273))
* Prevent logging openid provider errors twice ([25ffa1b](25ffa1bc2e2f1108f20b0336708d2410bb61c9e1))
* Remove credential escaping for postgres connections to allow for passwords with special characters ([230478a](230478aae947c86f4c6f1f251dcb30aeb1293283))
* Cycles in tasks array when memory caching was enabled ([f5a4c13](f5a4c136fbca6fc5770476e6de8d81173f007df2))
* Add missing error check ([5cc4927](5cc4927b9ef97667bf763772beb36225fdbeded8))
* Properly set tls config for mailer ([5743a4a](5743a4afe51de221beeeabe66552ae4d92eed1a6))
* Return 9:00 as default time for reminders if none was set ([79b3167](79b31673e2a79eaa124976840e85757d2bebb887))
* Reset id sequence when importing a dump from postgres ([0f555b7](0f555b7ec74ad493d2f70a4f4040db333943dc1c))
* Add validation for negative repeat after values ([dd46174](dd461746a655d716ef142d96a2bcef5615de3dd9))
* Lint ([1feb62c](1feb62cc458e939d46d16d24347557e7959ddfb9))
* Make sure to use user discoverability settings when searching list users ([382a788](382a7884be1f37da5c8f657c4b17316d8691dd59))
* Properly decode params in url ([8f27e7e](8f27e7e619ac73716211d838f52c73d7d97aead5))
* Return all users on a list when no search param was provided ([c51ee94](c51ee94ad1d552d69c71adfc2180c7ad0d23235d))
* Don't return email addresses from user search results ([3688bbd](3688bbde20e989397353ea4f7e872b00a53099c2))
* Lint ([77fafd5](77fafd5dc32aee464961be40d5d0ccf82490d02a))
* Increase test timeout ([26e2d0b](26e2d0bddeaea902dba055baf7a4c866a44ba7f1))
* Switch to buster for build image ([59796fd](59796fd4905fca74d26c5541878379cda143a30e))
* Use our own build image as base build image ([b6d7323](b6d7323cdfac958c9740feba1342114ab13a0afd))
* Use golang build image to test migrations ([84bcdbf](84bcdbf937c3be7823fcf8d5fef52e3cbb1c9bde))
* Switch back to alpine for everything, disable arm 32 docker builds ([7ffe9b6](7ffe9b625e441202a704db2774dd66fc38244c6d))


### Dependencies

* *(deps)* Update golang.org/x/sys commit hash to a851e7d (#972)
* *(deps)* Update golang.org/x/sys commit hash to aa78b53 (#973)
* *(deps)* Update golang.org/x/sys commit hash to 528a39c (#974)
* *(deps)* Update golang.org/x/sys commit hash to 437939a (#975)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.1 (#976)
* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.1.0 (#985)
* *(deps)* Update module github.com/spf13/viper to v1.9.0 (#987)
* *(deps)* Update golang.org/x/crypto commit hash to 089bfa5 (#979)
* *(deps)* Update golang.org/x/term commit hash to 140adaa (#983)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.6.0 (#988)
* *(deps)* Update golang.org/x/sys commit hash to b8560ed (#989)
* *(deps)* Update module github.com/golang-jwt/jwt/v4 to v4.1.0 (#991)
* *(deps)* Update golang.org/x/sys commit hash to 92d5a99 (#992)
* *(deps)* Update module github.com/swaggo/swag to v1.7.3 (#990)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.6.1 (#993)
* *(deps)* Update golang.org/x/sys commit hash to 1cf2251 (#994)
* *(deps)* Update golang.org/x/sys commit hash to 39ccf1d (#995)
* *(deps)* Update golang.org/x/term commit hash to 03fcf44 (#996)
* *(deps)* Update golang.org/x/oauth2 commit hash to 6b3c2da (#1000)
* *(deps)* Update golang.org/x/sys commit hash to 69063c4 (#1001)
* *(deps)* Update module github.com/gabriel-vasile/mimetype to v1.4.0 (#1004)
* *(deps)* Update postgres docker tag to v14 (#1005)
* *(deps)* Update module github.com/go-redis/redis/v8 to v8.11.4 (#1003)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.9 (#1008)
* *(deps)* Update golang.org/x/sys commit hash to 9d821ac (#1009)
* *(deps)* Update golang.org/x/sys commit hash to 0ec99a6 (#1010)
* *(deps)* Update golang.org/x/sys commit hash to 9d61738 (#1011)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.2 (#1012)
* *(deps)* Update golang.org/x/sys commit hash to 8e51046 (#1016)
* *(deps)* Update golang.org/x/sys commit hash to d6a326f (#1017)
* *(deps)* Update module github.com/swaggo/swag to v1.7.4 (#1018)
* *(deps)* Update golang.org/x/sys commit hash to 711f33c (#1019)
* *(deps)* Update golang.org/x/sys commit hash to 69cdffd (#1020)
* *(deps)* Update golang.org/x/oauth2 commit hash to ba495a6 (#1022)
* *(deps)* Update golang.org/x/image commit hash to 6944b10 (#1023)
* *(deps)* Update golang.org/x/sys commit hash to 6e78728 (#1024)
* *(deps)* Update golang.org/x/sys commit hash to b3129d9 (#1025)
* *(deps)* Update golang.org/x/sys commit hash to 611d5d6 (#1026)
* *(deps)* Update golang.org/x/sys commit hash to 39c9dd3 (#1027)
* *(deps)* Update golang.org/x/sys commit hash to a2f17f7 (#1028)
* *(deps)* Update golang.org/x/sys commit hash to 4dd7244 (#1029)
* *(deps)* Update golang.org/x/sys commit hash to ae416a5 (#1030)
* *(deps)* Update golang.org/x/sys commit hash to 7861aae (#1031)
* *(deps)* Update golang.org/x/oauth2 commit hash to d3ed0bb (#1032)
* *(deps)* Update module github.com/labstack/gommon to v0.3.1 (#1033)
* *(deps)* Update golang.org/x/sys commit hash to c75c477 (#1034)
* *(deps)* Update golang.org/x/sys commit hash to ebca88c (#1035)
* *(deps)* Update golang.org/x/sys commit hash to e0b2ad0 (#1037)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.3 (#1038)
* *(deps)* Update golang.org/x/crypto commit hash to ceb1ce7 (#1041)
* *(deps)* Update module github.com/lib/pq to v1.10.4 (#1040)
* *(deps)* Update golang.org/x/sys commit hash to 51b60fd (#1042)
* *(deps)* Update golang.org/x/sys commit hash to 99a5385 (#1043)
* *(deps)* Update golang.org/x/sys commit hash to f221eed (#1044)
* *(deps)* Update golang.org/x/sys commit hash to 0c823b9 (#1045)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.4 (#1046)
* *(deps)* Update golang.org/x/sys commit hash to 0a5406a (#1048)
* *(deps)* Update golang.org/x/crypto commit hash to b4de73f (#1047)
* *(deps)* Update module github.com/ulule/limiter/v3 to v3.9.0 (#1049)
* *(deps)* Update golang.org/x/crypto commit hash to ae814b3 (#1050)
* *(deps)* Update golang.org/x/sys commit hash to dee7805 (#1051)
* *(deps)* Update golang.org/x/sys commit hash to ef496fb (#1052)
* *(deps)* Update golang.org/x/sys commit hash to fe61309 (#1054)
* *(deps)* Update module github.com/swaggo/swag to v1.7.6 (#1055)
* *(deps)* Update golang.org/x/crypto commit hash to 5770296 (#1056)
* *(deps)* Update module github.com/golang-jwt/jwt/v4 to v4.2.0 (#1057)
* *(deps)* Update golang.org/x/sys commit hash to 94396e4 (#1058)
* *(deps)* Update golang.org/x/sys commit hash to 97ca703 (#1059)
* *(deps)* Update golang.org/x/crypto commit hash to 4570a08 (#1062)
* *(deps)* Update golang.org/x/sys commit hash to 798191b (#1061)
* *(deps)* Update golang.org/x/sys commit hash to af8b642 (#1063)
* *(deps)* Update module github.com/spf13/viper to v1.10.0 (#1064)
* *(deps)* Update golang.org/x/sys commit hash to 03aa0b5 (#1067)
* *(deps)* Update golang.org/x/sys commit hash to 3b038e5 (#1068)
* *(deps)* Update module github.com/spf13/cobra to v1.3.0 (#1070)
* *(deps)* Update golang.org/x/sys commit hash to 4825e8c (#1071)
* *(deps)* Update module github.com/spf13/viper to v1.10.1 (#1072)
* *(deps)* Update golang.org/x/crypto commit hash to e495a2d (#1073)
* *(deps)* Update golang.org/x/sys commit hash to 4abf325 (#1074)
* *(deps)* Update golang.org/x/sys commit hash to 1d35b9e (#1075)
* *(deps)* Update module github.com/magefile/mage to v1.12.0 (#1076)
* *(deps)* Update module github.com/magefile/mage to v1.12.1 (#1077)
* *(deps)* Update module github.com/getsentry/sentry-go to v0.12.0 (#1079)
* *(deps)* Update module github.com/swaggo/swag to v1.7.8 (#1080)
* *(deps)* Update module github.com/spf13/afero to v1.7.0 (#1078)
* *(deps)* Update module github.com/spf13/afero to v1.7.1 (#1081)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.10 (#1082)
* *(deps)* Update module github.com/spf13/afero to v1.8.0 (#1083)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.6.2 (#1084)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.6.3 (#1089)
* *(deps)* Update golang.org/x/sys commit hash to a018aaa (#1088)
* *(deps)* Update golang.org/x/sys commit hash to 5a964db (#1090)
* *(deps)* Update golang.org/x/crypto commit hash to 5e0467b (#1091)
* *(deps)* Update golang.org/x/sys commit hash to da31bd3 (#1093)
* *(deps)* Update module github.com/prometheus/client_golang to v1.12.0 (#1094)
* *(deps)* Update golang.org/x/crypto commit hash to e04a857 (#1097)
* *(deps)* Update golang.org/x/crypto commit hash to aa10faf (#1098)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.11 (#1099)
* *(deps)* Update golang.org/x/crypto commit hash to 198e437 (#1100)
* *(deps)* Update golang.org/x/sys commit hash to 99c3d69 (#1101)
* *(deps)* Update module github.com/prometheus/client_golang to v1.12.1 (#1102)
* *(deps)* Update klakegg/hugo docker tag to v0.92.0 (#1103)
* *(deps)* Update klakegg/hugo docker tag to v0.92.1 (#1104)
* *(deps)* Update golang.org/x/crypto commit hash to 30dcbda (#1105)
* *(deps)* Update module github.com/swaggo/swag to v1.7.9 (#1106)
* *(deps)* Update golang.org/x/sys commit hash to 1c1b9b1 (#1107)
* *(deps)* Update module github.com/spf13/afero to v1.8.1 (#1108)
* *(deps)* Update golang.org/x/sys commit hash to 5739886 (#1110)
* *(deps)* Update golang.org/x/crypto commit hash to 20e1d8d (#1111)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.5 (#1112)
* *(deps)* Update golang.org/x/crypto commit hash to bba287d (#1113)
* *(deps)* Update golang.org/x/crypto commit hash to dad3315 (#1114)
* *(deps)* Update module github.com/golang-jwt/jwt/v4 to v4.3.0 (#1117)
* *(deps)* Update golang.org/x/sys commit hash to 3681064 (#1116)
* *(deps)* Update golang.org/x/crypto commit hash to db63837 (#1115)
* *(deps)* Update golang.org/x/crypto commit hash to f4118a5 (#1118)
* *(deps)* Update golang.org/x/crypto commit hash to 8634188 (#1121)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.6 (#1122)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.7 (#1123)
* *(deps)* Update module github.com/swaggo/swag to v1.8.0 (#1124)
* *(deps)* Update golang.org/x/sys commit hash to 0005352 (#1125)
* *(deps)* Update golang.org/x/sys commit hash to f242548 (#1126)
* *(deps)* Update klakegg/hugo docker tag to v0.92.2 (#1127)
* *(deps)* Update golang.org/x/sys commit hash to dbe011f (#1129)
* *(deps)* Update golang.org/x/sys commit hash to 95c6836 (#1130)
* *(deps)* Update golang.org/x/oauth2 commit hash to ee48083 (#1128)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.12 (#1132)
* *(deps)* Update golang.org/x/sys commit hash to 4e6760a (#1131)
* *(deps)* Update golang.org/x/image commit hash to 723b81c (#1133)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.7.0 (#1134)
* *(deps)* Update klakegg/hugo docker tag to v0.93.0 (#1135)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.8 (#1136)
* *(deps)* Update klakegg/hugo docker tag to v0.93.2 (#1137)
* *(deps)* Update golang.org/x/sys commit hash to 22a9840 (#1138)
* *(deps)* Update golang.org/x/crypto commit hash to efcb850 (#1139)
* *(deps)* Update golang.org/x/oauth2 commit hash to 6242fa9 (#1140)
* *(deps)* Update golang.org/x/sys commit hash to b874c99 (#1141)
* *(deps)* Update klakegg/hugo docker tag to v0.93.3 (#1142)
* *(deps)* Update module github.com/labstack/echo/v4 to v4.7.1 (#1146)
* *(deps)* Update module github.com/stretchr/testify to v1.7.1 (#1148)
* *(deps)* Update module github.com/swaggo/swag to v1.8.1 (#1156)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.11 (#1143)
* *(deps)* Update module github.com/spf13/cobra to v1.4.0 (#1145)
* *(deps)* Update module github.com/lib/pq to v1.10.5 (#1157)
* *(deps)* Update module github.com/spf13/viper to v1.11.0 (#1159)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.12 (#1162)
* *(deps)* Update module github.com/prometheus/client_golang to v1.12.2 (#1166)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.13 (#1165)
* *(deps)* Update module github.com/coreos/go-oidc/v3 to v3.2.0 (#1164)
* *(deps)* Update module github.com/swaggo/swag to v1.8.2 (#1167)
* *(deps)* Update module github.com/lib/pq to v1.10.6 (#1169)
* *(deps)* Update module gopkg.in/yaml.v3 to v3.0.1 (#1179)
* *(deps)* Update module github.com/imdario/mergo to v0.3.13 (#1178)
* *(deps)* Update module github.com/stretchr/testify to v1.7.2 (#1182)
* *(deps)* Update module github.com/swaggo/swag to v1.8.3 (#1185)
* *(deps)* Update module github.com/spf13/cobra to v1.5.0 (#1192)
* *(deps)* Update module github.com/golang-jwt/jwt/v4 to v4.4.2 (#1193)
* *(deps)* Update module github.com/stretchr/testify to v1.8.0 (#1191)
* *(deps)* Update module github.com/go-testfixtures/testfixtures/v3 to v3.8.0 (#1168)
* *(deps)* Update module github.com/mattn/go-sqlite3 to v1.14.14 (#1194)
* *(deps)* Update golang.org/x/term digest to 065cf7b (#1207)
* *(deps)* Update golang.org/x/image digest to 41969df (#1203)
* *(deps)* Update module github.com/yuin/goldmark to v1.4.13 (#1209)
* *(deps)* Update golang.org/x/crypto digest to 0559593 (#1202)
* *(deps)* Update module github.com/spf13/afero to v1.9.0 (#1210)
* *(deps)* Update module github.com/gabriel-vasile/mimetype to v1.4.1 (#1208)
* *(deps)* Update golang.org/x/sync digest to 0de741c (#1205)
* *(deps)* Update github.com/c2h5oh/datasize digest to 859f65c (#1201)
* *(deps)* Update golang.org/x/oauth2 digest to 2104d58 (#1204)
* *(deps)* Update golang.org/x/sys digest to c0bba94 (#1206)
* *(deps)* Update golang.org/x/oauth2 digest to c8730f7 (#1214)
* *(deps)* Update module github.com/spf13/afero to v1.9.2 (#1215)
* *(deps)* Update module github.com/swaggo/swag to v1.8.4 (#1216)
* *(deps)* Update module github.com/spf13/viper to v1.12.0 (#1180)
* *(deps)* Update golang.org/x/sys digest to 1609e55 (#1217)
* *(deps)* Update module github.com/go-testfixtures/testfixtures/v3 to v3.8.1 (#1226)
* *(deps)* Update module go to 1.18 (#1225)

### Documentation
* Add docker-compose example with no proxy ([4255bc3](4255bc3a945b6fe4314e3cd3f62908dd1be1ff4a))
* Add another youtube tutorial ([dbd6f36](dbd6f36da6e56355993cc1411379997e26c88b36))
* Fix api url in docker examples without a proxy ([68998e9](68998e90a446569869fb150bd5fc0739f496b066))
* Make sure all links to vikunja pages are https ([cc612d5](cc612d505f22e5d895b6ebda61fe62498634cec5))
* Update backup instructions ([4829c89](4829c899400544ad27cacfb7d19b40988302a413))
* Add postgres to docker-compose examples ([2aea169](2aea1691cf33b7d9e03fbe2c711af7d8f76d9724))
* Improve development docs ([9bf32aa](9bf32aae99a7e69cce0cd4477e8fc8ddcaea25ea))
* Add another tutorial link ([1fa74cb](1fa74cba6407c2b694b14f8439f1492476433d62))
* Improve wording for systemd ([13561f2](13561f211493903b17c856b3010345ea9df725d4))
* Update testing ([da318e3](da318e3db15121ba864db8450a76ba9ed18b9fd5))
* Add guide for Synology NAS ([049ae39](049ae39c62079f77921b7a9fad5023b2c1c0c1c5))


### Features

* *(docs)* Add details of using NGINX Proxy Manager to the Reverse Proxy docs (#13)
* *(docs)* Add versions explanation
* *(mail)* Don't try to authenticate when no username and password was provided* Add better error logs for mage commands ([bb086eb](bb086eb9f87669f844c283d42ea9ca9f3f5a7877))
* Expose if task comments are enabled or not in /info ([ae8db17](ae8db176db57fa6176e00b87924f70352332ca66))
* Improve account deletion email grammar (#1006) ([dcb52c0](dcb52c00f1c6b3217e2b508d7799fc83adb3b055))
* Add more debug logging when deleting users ([8f55af0](8f55af07c936218487ec94e65c6673fbddd0cdb5))
* Don't require a password for data export from users authenticated with third-party auth ([9eca971](9eca971c938699d481915fb6e14c765aea1fa3b5))
* Expose if a user is a local user through its jwt token ([516c812](516c812043e77be7f834ae1326d13d39e156ef77))
* Expose if a user is a local user through the /user endpoint ([2683ef2](2683ef23d538eb846d5d799798fa82cca70dc017))
* Enable rate limit for unauthenticated routes ([093d0c6](093d0c65ca6338358dbd1df904daadd7808f2817))
* Use wallpaper topic for default unsplash background list ([88a2ced](88a2cede19f1844814530af948c3cc5a0b026419))
* Gravatar - Lowercase emails before MD5 hash (#10) ([36bf3d2](36bf3d216a7be28e917e2816a9e5da43439f2c20))
* Add marble avatar (#1060) ([73ee696](73ee696fc3cf941af2d2c2cf81224aa01f93234e))
* Save user language in the settings ([a98119f](a98119f2d670a11efab6008129b767f9208f8113))
* Add time zone setting for reminders (#1092) ([61d49c3](61d49c3a56a59e52ce407b858ddd4aa573dbee9d))
* Add long-lived api tokens (#1085) ([1322cb1](1322cb16d76a40ad90631e3e091da0f0d44957a9))
* Upgrade golangci-lint to 1.45.2 ([5cf263a](5cf263a86f954a38cbfafb6b0857bf591f82a811))
* Add date math for filters (#1086) ([0a1d8c9](0a1d8c940410b03a78016ac6110883ca05484816))
* Add migration to create BlurHash strings for all list backgrounds ([362706b](362706b38d52720b5a1615e185a985b7708168f7))
* Generate a BlurHash when uploading a new image ([f83b09a](f83b09af59ed25425a16824ccf48d903c81e861a))
* Save BlurHash from unsplash when selecting a photo from unsplash ([2ec7d7a](2ec7d7a8a85cc12c07d20cfab9b90a78a7857eb6))
* Return BlurHash for unsplash search results ([6df8658](6df865876df961f2bec476126bf6e7fbe5d43e0e))
* Add caldav tokens (#1065) ([e4b50e8](e4b50e84a44f809cc829c2fdb6f52b03b40a367b))
* Ability to serve static files (#1174) ([acaa850](acaa85083f2bebbc67608ae0f454ed5e9a3ef8a0))
* Restrict max avatar size ([2f25b48](2f25b48869f59256bf7d692c4486c64c30b85e5e))
* Send overdue tasks email notification at 9:00 in the user's time zone ([7eb3b96](7eb3b96a4465ca6648572b07c506c06f2c28c375))
* Add setting to change overdue tasks reminder email time ([8869adf](8869adfc276f674b686bf68f949d7efbb417e55b))
* Allow only the authors of task comments to edit them ([01271c4](01271c4c0111b3b040dcb9a0d502d31078ad6d4b))
* Migrate away from gomail ([30e0e98](30e0e98f7738e36698990523377f47edcbf6806c))
* Embed the vikunja logo as inline attachment ([f4f8450](f4f8450d166f1a836eea202dd0340d2156d3dfe9))
* Use embed fs directly to embed the logo in mails ([73c4c39](73c4c399e5d610bb713f1e9feab543e0425ee959))
* Use actual uuids for tasks ([62325de](62325de9cd5da5b70987081956a28e7baa907081))
* Add issue template ([117f6b3](117f6b38e1d35c09f2657975ea75dcfedcd8425d))


### Miscellaneous Tasks

* *(ci)* Use latest version of s3 plugin
* *(ci)* Sign drone config
* *(docs)* Update docs about compiling from source
* *(docs)* Redirect properly from /docs/docs
* *(docs)* Add new mailer option to docs
* *(docs)* Clarify openid setup with environment variables
* *(docs)* Add frontendurl to all example configs
* *(mage)* Don't set api packages when they are not used* Sign drone config ([1d8d0f1](1d8d0f140e4f2a59947167bd597e5f12b84b009d))
* Cleanup namespace creation ([b60c69c](b60c69c5a8c004a780b989cf0bb8ab6455086b0f))
* Generate swagger docs ([ba2bdff](ba2bdff39109db9ecc4b525e39e2642b41ac03b8))
* Go mod tidy ([726a517](726a517bec731f1af8e3186e280718fef02cadf7))
* Upgrade trello api wrapper and remove fork ([7e99618](7e99618319547c7e7dfa2cc063f654300f7074fb))
* Use our custom build image to build docker image ([251b877](251b877015761fdd2b8dbd18cd8ec696dc374103))
* Update golangci-lint ([430057a](430057a404b04e75c62a15693f479c6fc8e63189))


### Other

* *(other)* Healthcheck endpoint (#998)
* *(other)* Added the ability to configure the JWT expiry date using a new server.jwtttl config parameter. (#999)
* *(other)* Enable a list to be moved across namespaces (#1096)
* *(other)* A bunch of dependency updates at once (#1155)
* *(other)* Add client-cert parameters of the Go pq driver to the Vikunja config (#1161)
* *(other)* Add exec to run script to run app as PID 1 (#1200)

## [0.18.1] - 2021-09-08

### Fixed

* Docs: Add another third-party tutorial link
* Don't try to export items which do not have a parent
* fix(deps): update golang.org/x/sys commit hash to 6f6e228 (#970)
* fix(deps): update golang.org/x/sys commit hash to c212e73 (#971)
* Fix exporting tasks from archived lists
* Fix lint
* Fix tasks not exported
* Fix tmp export file created in the wrong path

## [0.18.0] - 2021-09-05

### Added

* Add default list setting (#875)
* Add menu link to Vikunja Cloud in docs
* Add more logging and better error messages for openid authentication + clarify docs
* Add more logging for test data api endpoint
* Add searching for tasks by index
* Add setting for first day of the week
* Add support of Unix socket (#912)
* Add truncate parameter to test fixtures setup
* Notify the user after three failed login attempts
* Reorder tasks, lists and kanban buckets (#923)
* Send a notification on failed TOTP
* Task mentions (#926)
* Try to get more information about the user when authenticating with openid
* User account deletion (#937)
* User Data Export and import (#967)

### Changed

* Allow running migration 20210711173657 multiple times to fix issues when it didn't completely run through previously
* Better logging for errors while importing a bunch of tasks
* Change task title to TEXT instead of varchar(250) to allow for longer task titles
* Disable the user account after 10 failed password attempts
* Docs: Add a note about default password
* Docs: Add another youtube tutorial
* Docs: Add ios to the list of not working caldav clients
* Docs: Add k8s-at-home Helm Chart for Vikunja
* Docs: Add other installation resources
* Docs: Add translation docs
* Docs: Fix rewrite rules in apache example configs
* Docs: Translation now happening at crowdin
* Docs: Update translation guidelines
* Don't fail when removing the last bucket in migration from other services
* Don't notify the user who created the team
* Don't use the mariadb root user in docker-compose examples
* Ensure case insensitive search on postgres (#927)
* Increase test timeout
* Only filter out failing openid providers if multiple are configured and one of them failed
* Only send an email about failed totp after three failed attempts
* Rearrange setting frontend url in config
* Refactor user email confirmation + password reset handling (#919)
* Rename and sign drone config
* Replace jwt-go with github.com/golang-jwt/jwt
* Reset failed totp attempts when logging in successfully
* Save user tokens as text and not varchar
* Save user tokens as varchar(450) and not text to fix mysql indexing issues
* Set todoist migration redirect url to the frontend url by default
* Show config full paths and env variables with config options
* Switch the :latest docker image tag to contain the latest release instead of the latest unstable
* Tune test db server settings to speed up tests (#939)

### Fixed

* Fix authentication callback
* Fix duplicating empty lists
* Fix error handling when deleting an attachment file
* Fix error when searching for a namespace returned no results
* Fix error when searching for a namespace with subscribers
* Fix goimports
* Fix importing archived projects and done items from todoist
* Fix jwt middleware
* Fix lint
* Fix mapping task priorities from Vikunja to calDAV
* Fix moving the done bucket around
* Fix old references to master in docs
* Fix panic on invalid smtp config
* Fix parsing openid config when using a json config file
* Fix saving pointer values to memory keyvalue
* Fix saving reminders of repeating tasks
* Fix setting a saved filter as favorite
* Fix setting task favorite status of related tasks
* Fix setting up keyvalue storage in tests
* Fix swagger docs for create requests
* Fix task relations not getting properly cleaned up when deleting them
* Fix tests & lint
* Make sure a bucket exists or use the default bucket when importing tasks
* Make sure all associated entities of a task are deleted when the task is deleted
* Make sure list / task favorites are set per user, not per entity (#915)
* Make sure the configured frontend url always has a / at the end
* Refactor & fix storing struct-values in redis keyvalue
* Todoist migration: don't panic if no reminder was found for task

### Dependency updates

* fix(deps): update golang.org/x/sys commit hash to 63515b4 (#959)
* fix(deps): update golang.org/x/sys commit hash to 97244b9 (#965)
* fix(deps): update golang.org/x/sys commit hash to f475640 (#962)
* fix(deps): update golang.org/x/sys commit hash to f4d4317 (#961)
* fix(deps): update module github.com/lib/pq to v1.10.3 (#963)
* Update alpine Docker tag to v3.13 (#884)
* Update alpine Docker tag to v3.14 (#889)
* Update golang.org/x/crypto commit hash to 0a44fdf (#944)
* Update golang.org/x/crypto commit hash to 0ba0e8f (#943)
* Update golang.org/x/crypto commit hash to 32db794 (#949)
* Update golang.org/x/crypto commit hash to 5ff15b2 (#891)
* Update golang.org/x/crypto commit hash to a769d52 (#916)
* Update golang.org/x/image commit hash to 775e3b0 (#880)
* Update golang.org/x/image commit hash to a66eb64 (#900)
* Update golang.org/x/image commit hash to e6eecd4 (#893)
* Update golang.org/x/net commit hash to 37e1c6af
* Update golang.org/x/oauth2 commit hash to 14747e6 (#894)
* Update golang.org/x/oauth2 commit hash to 2bc19b1 (#955)
* Update golang.org/x/oauth2 commit hash to 6f1e639 (#931)
* Update golang.org/x/oauth2 commit hash to 7df4dd6 (#952)
* Update golang.org/x/oauth2 commit hash to a41e5a7 (#902)
* Update golang.org/x/oauth2 commit hash to a8dc77f (#896)
* Update golang.org/x/oauth2 commit hash to bce0382 (#895)
* Update golang.org/x/oauth2 commit hash to d040287 (#888)
* Update golang.org/x/oauth2 commit hash to f6687ab (#862)
* Update golang.org/x/oauth2 commit hash to faf39c7 (#935)
* Update golang.org/x/sys commit hash to 00dd8d7 (#953)
* Update golang.org/x/sys commit hash to 15123e1 (#946)
* Update golang.org/x/sys commit hash to 1e6c022 (#947)
* Update golang.org/x/sys commit hash to 30e4713 (#945)
* Update golang.org/x/sys commit hash to 41cdb87 (#956)
* Update golang.org/x/sys commit hash to 7d9622a (#948)
* Update golang.org/x/sys commit hash to bfb29a6 (#951)
* Update golang.org/x/sys commit hash to d867a43 (#934)
* Update golang.org/x/sys commit hash to e5e7981 (#933)
* Update golang.org/x/sys commit hash to f52c844 (#954)
* Update golang.org/x/term commit hash to 6886f2d (#887)
* Update module getsentry/sentry-go to v0.11.0 (#869)
* Update module github.com/coreos/go-oidc to v3 (#885)
* Update module github.com/gabriel-vasile/mimetype to v1.3.1 (#904)
* Update module github.com/golang-jwt/jwt to v3.2.2 (#928)
* Update module github.com/golang-jwt/jwt to v4 (#930)
* Update module github.com/go-redis/redis/v8 to v8.11.0 (#903)
* Update module github.com/go-redis/redis/v8 to v8.11.1 (#925)
* Update module github.com/go-redis/redis/v8 to v8.11.2 (#932)
* Update module github.com/go-redis/redis/v8 to v8.11.3 (#942)
* Update module github.com/iancoleman/strcase to v0.2.0 (#918)
* Update module github.com/labstack/echo/v4 to v4.4.0 (#917)
* Update module github.com/labstack/echo/v4 to v4.5.0 (#929)
* Update module github.com/mattn/go-sqlite3 to v1.14.8 (#921)
* Update module github.com/spf13/cobra to v1.2.0 (#905)
* Update module github.com/spf13/cobra to v1.2.1 (#906)
* Update module github.com/spf13/viper to v1.8.0 (#890)
* Update module github.com/spf13/viper to v1.8.1 (#899)
* Update module github.com/swaggo/swag to v1.7.1 (#936)
* Update module github.com/yuin/goldmark to v1.3.8 (#892)
* Update module github.com/yuin/goldmark to v1.3.9 (#901)
* Update module github.com/yuin/goldmark to v1.4.0 (#908)
* Update module go-redis/redis/v8 to v8.10.0 (#882)
* Update module go-redis/redis/v8 to v8.7.1 (#807)
* Update module go-testfixtures/testfixtures/v3 to v3.6.1 (#868)
* Update module lib/pq to v1.10.2 (#865)
* Update module prometheus/client_golang to v1.11.0 (#879)
* Update module yuin/goldmark to v1.3.6 (#863)
* Update module yuin/goldmark to v1.3.7 (#867)
* Update monachus/hugo Docker tag to v0.75.1 (#940)

## [0.17.1] - 2021-06-09

### Fixed

* Fix parsing openid config when using a json config file

## [0.17.0] - 2021-05-14

### Added

* Add a "done" option to kanban buckets (#821)
* Add arm64 builds
* Add basic auth for metrics endpoint
* Add bucket limit validation
* Add crud endpoints for notifications (#801)
* Add endpoint to remove a list background
* Add events (#777)
* Add github funding link
* Add link share password authentication (#831)
* Add names for link shares (#829)
* Add notifications package for easy sending of notifications (#779)
* Add reminders for overdue tasks (#832)
* Add repeat monthly setting for tasks (#834)
* Add security information to readme
* Add separate docker manifest file for latest docker images
* Add systemd service file to linux packages
* Add test for moving a task to another list
* Enable searching users by full email or name
* Expose tls parameter of Go MySQL driver to config file (#855)
* Pagingation for tasks in kanban buckets (#805)

### Changed

* Change keyvalue.Get to return if a value exists or not instead of an error
* Change main branch to main
* Change test file names to unstable
* Change the name of the newly created bucket from "New Bucket" to "Backlog"
* Change unstable versions in migration tests
* Check if we're on main and change the version name accordingly if that's the case
* Cleanup listener names
* Cleanup old docs themes submodule
* Disable deb repo in drone
* Don't keep old releases from os packages when releasing for master
* Don't try to get users for tasks if no tasks were found when looking for reminders
* Explicitly add docker build step for latest
* Explicitly check if there are Ids before trying to get items by a list of Ids
* Improve duration format of overdue tasks in reminders
* Improve loading labels performance (#824)
* Improve sending overdue task reminders by only sending one for all overdue tasks
* Make sure all tables are properly pluralized
* Only send reminders for undone tasks
* Re-Enable migration test steps in pipeline
* Refactor getting all namespaces
* Remove unused tools from tools.go
* Run all lint checks at once
* Send a notification to the user when they are added to the list
* Show empty avatar when the user was not found
* Subscribe a user to a task when they are assigned to it
* Subscriptions and notifications for namespaces, tasks and lists (#786)
* Switch building the docs to download the theme instead of building
* Switch telegram notifications to matrix notifications
* Temporarily disable migration step
* Temporary build fix
* Update changelog
* Update copyright year
* Update README (#858)
* Use golang's tzdata package to handle time zones

### Fixed

* Explicitly set darwin-10.15 when building binaries
* Fix build
* Fix checking list rights when accessing a bucket
* Fix /dav/principals/*/ throwing a server error when accessed with GET instead of PROPFIND (#769)
* Fix deleting task relations
* Fix docs
* Fix drone file
* Fix due dates with times when migrating from todoist
* Fix event error handler retrying infinitely
* Fix filter for task index
* Fix getting lists for shared, favorite and saved lists namespace
* Fix getting user info from /user endpoint for link shares
* Fix IncrBy and DecrBy in memory keyvalue implementation if there was no value set previously
* Fix lint
* Fix matrix notify room id
* Fix moving repeating tasks to the done bucket
* Fix multiarch docker image building
* Fix not able to make saved filters favorite
* Fix notifications table not being created on initial setup
* Fix resetting the bucket limit
* Fix retrieving over openid providers if there are none
* Fix sending notifications to users if the user object didn't have an email
* Fix setting the user in created_by when uploading an attachment
* Fix shared lists showing up twice
* Fix tests
* Fix the shared lists pseudo namespace containing owned lists
* Fix unstable version build file names
* Fix user uploaded avatars
* Pin golang alpine builder image to 3.12 to fix builds on arm
* Revert "Update alpine Docker tag to v3.13 (#768)"

### Dependency Updates

* Update alpine Docker tag to v3.13 (#768)
* Update github.com/gordonklaus/ineffassign commit hash to 2e10b26 (#803)
* Update github.com/gordonklaus/ineffassign commit hash to d0e41b2 (#780)
* Update golang.org/x/crypto commit hash to 0c34fe9 (#822)
* Update golang.org/x/crypto commit hash to 3497b51 (#853)
* Update golang.org/x/crypto commit hash to 38f3c27 (#854)
* Update golang.org/x/crypto commit hash to 4f45737 (#836)
* Update golang.org/x/crypto commit hash to 513c2a4 (#817)
* Update golang.org/x/crypto commit hash to 5bf0f12 (#839)
* Update golang.org/x/crypto commit hash to 5ea612d (#797)
* Update golang.org/x/crypto commit hash to 83a5a9b (#840)
* Update golang.org/x/crypto commit hash to b8e89b7 (#793)
* Update golang.org/x/crypto commit hash to c07d793 (#861)
* Update golang.org/x/crypto commit hash to cd7d49e (#860)
* Update golang.org/x/crypto commit hash to e6e6c4f (#816)
* Update golang.org/x/crypto commit hash to e9a3299 (#851)
* Update golang.org/x/image commit hash to 4410531 (#788)
* Update golang.org/x/image commit hash to 55ae14f (#787)
* Update golang.org/x/image commit hash to 7319ad4 (#852)
* Update golang.org/x/image commit hash to ac19c3e (#798)
* Update golang.org/x/oauth2 commit hash to 0101308 (#776)
* Update golang.org/x/oauth2 commit hash to 01de73c (#762)
* Update golang.org/x/oauth2 commit hash to 16ff188 (#789)
* Update golang.org/x/oauth2 commit hash to 22b0ada (#823)
* Update golang.org/x/oauth2 commit hash to 2e8d934 (#827)
* Update golang.org/x/oauth2 commit hash to 5366d9d (#813)
* Update golang.org/x/oauth2 commit hash to 5e61552 (#833)
* Update golang.org/x/oauth2 commit hash to 6667018 (#783)
* Update golang.org/x/oauth2 commit hash to 81ed05c (#848)
* Update golang.org/x/oauth2 commit hash to 8b1d76f (#764)
* Update golang.org/x/oauth2 commit hash to 9bb9049 (#796)
* Update golang.org/x/oauth2 commit hash to af13f52 (#773)
* Update golang.org/x/oauth2 commit hash to ba52d33 (#794)
* Update golang.org/x/oauth2 commit hash to cd4f82c (#815)
* Update golang.org/x/oauth2 commit hash to d3ed898 (#765)
* Update golang.org/x/oauth2 commit hash to f9ce19e (#775)
* Update golang.org/x/sync commit hash to 036812b (#799)
* Update golang.org/x/term commit hash to 6a3ed07 (#800)
* Update golang.org/x/term commit hash to 72f3dc4 (#828)
* Update golang.org/x/term commit hash to a79de54 (#850)
* Update golang.org/x/term commit hash to b80969c (#843)
* Update golang.org/x/term commit hash to c04ba85 (#849)
* Update golang.org/x/term commit hash to de623e6 (#818)
* Update golang.org/x/term commit hash to f5beecf (#845)
* Update module adlio/trello to v1.9.0 (#825)
* Update module coreos/go-oidc to v3 (#760)
* Update module gabriel-vasile/mimetype to v1.2.0 (#812)
* Update module gabriel-vasile/mimetype to v1.3.0 (#857)
* Update module getsentry/sentry-go to v0.10.0 (#792)
* Update module go-redis/redis/v8 to v8.4.10 (#771)
* Update module go-redis/redis/v8 to v8.4.11 (#774)
* Update module go-redis/redis/v8 to v8.4.9 (#770)
* Update module go-redis/redis/v8 to v8.5.0 (#778)
* Update module go-redis/redis/v8 to v8.6.0 (#795)
* Update module go-sql-driver/mysql to v1.6.0 (#826)
* Update module go-testfixtures/testfixtures/v3 to v3.5.0 (#761)
* Update module go-testfixtures/testfixtures/v3 to v3.6.0 (#838)
* Update module iancoleman/strcase to v0.1.3 (#766)
* Update module imdario/mergo to v0.3.12 (#811)
* Update module jgautheron/goconst to v1 (#804)
* Update module labstack/echo/v4 to v4.2.0 (#785)
* Update module labstack/echo/v4 to v4.2.1 (#810)
* Update module labstack/echo/v4 to v4.2.2 (#830)
* Update module labstack/echo/v4 to v4.3.0 (#856)
* Update module lib/pq to v1.10.0 (#809)
* Update module lib/pq to v1.10.1 (#841)
* Update module mattn/go-sqlite3 to v1.14.7 (#835)
* Update module olekukonko/tablewriter to v0.0.5 (#782)
* Update module prometheus/client_golang to v1.10.0 (#819)
* Update module spf13/afero to v1.6.0 (#820)
* Update module spf13/cobra to v1.1.2 (#781)
* Update module spf13/cobra to v1.1.3 (#784)
* Update module src.techknowlogick.com/xgo to v1.3.0+1.16.0 (#791)
* Update module src.techknowlogick.com/xgo to v1.4.0+1.16.2 (#814)
* Update module stretchr/testify to v1.7.0 (#763)

## [0.16.1] - 2021-04-22

### Fixed

* Fix checking list rights when accessing a bucket
* Remove old deb-structure ci step
* Fix docker from

## [0.16.0] - 2021-01-10

### Added

* Add colors for caldav (#738)
* Add email reminders (#743)
* Add "like" filter comparator
* Add login via email (#740)
* Add Microsoft Todo migration (#737)
* Add name field to users
* Add support for migrating todoist boards (#732)
* Add task filter for assignees (#746)
* Add task filter for labels (#747)
* Add task filter for lists and namespaces (#748)
* Add task filter for reminders (#745)
* Add task filters for kanban
* Add testing endpoint to reset db tables (#716)
* Add tests for sending task reminders (#757)
* Add trello migration (#734)
* Authentication with OpenID Connect providers (#713)

### Fixed

* Fix completion status in DAV for OpenTasks and multiline descriptions (#697)
* Fix docs about caldav tasks.org
* Fix drone badge in README
* Fix getting current user when updating avatar or user name
* Fix go header lint
* Fix /info endpoint 500 error when no openid providers were configured
* Fix missing auto increments from b0d4902406 on mysql
* Fix not possible to create tasks if metrics were enabled
* Fix password reset without a reseet token
* Fix task updated timestamp not being updated in the response after updating a task

### Changed

* Change avatar endpoint
* Change license to AGPLv3
* Clarify docs about cors configuration
* Don't create a list identifier by default
* Make sure all int64 db fields are using bigint when actually storing the data (#741)
* Make sure a password reset token can be used only once
* Make the debian repo structure for buster instead of strech
* Refactor adding more details to tasks (#739)
* Simplify updating task reminders
* Update code header template
* Update github.com/gordonklaus/ineffassign commit hash to 3b93a88 (#701)
* Update github.com/gordonklaus/ineffassign commit hash to 8eed68e (#755)
* Update github.com/jgautheron/goconst commit hash to b58d7cf (#702)
* Update github.com/jgautheron/goconst commit hash to ccae5bf (#712)
* Update github.com/jgautheron/goconst commit hash to f8e4fe8 (#703)
* Update golang.org/x/crypto commit hash to 0c6587e (#706)
* Update golang.org/x/crypto commit hash to 5f87f34 (#729)
* Update golang.org/x/crypto commit hash to 8b5274c (#733)
* Update golang.org/x/crypto commit hash to 9d13527 (#736)
* Update golang.org/x/crypto commit hash to be400ae (#719)
* Update golang.org/x/crypto commit hash to c8d3bf9 (#710)
* Update golang.org/x/crypto commit hash to eec23a3 (#749)
* Update golang.org/x/image commit hash to 35266b9 (#727)
* Update golang.org/x/lint commit hash to 83fdc39 (#728)
* Update golang.org/x/oauth2 commit hash to 08078c5 (#722)
* Update golang.org/x/oauth2 commit hash to 0b49973 (#718)
* Update golang.org/x/oauth2 commit hash to 9fd6049 (#714)
* Update golang.org/x/sync commit hash to 09787c9 (#725)
* Update golang.org/x/sync commit hash to 67f06af (#695)
* Update golang.org/x/term commit hash to 2321bbc (#731)
* Update golang.org/x/term commit hash to ee85cb9 (#726)
* Update module cweill/gotests to v1.6.0 (#752)
* Update module fzipp/gocyclo to v0.3.1 (#696)
* Update module gabriel-vasile/mimetype to v1.1.2 (#708)
* Update module getsentry/sentry-go to v0.8.0 (#709)
* Update module getsentry/sentry-go to v0.9.0 (#723)
* Update module go-redis/redis/v8 to v8.4.4 (#742)
* Update module go-redis/redis/v8 to v8.4.6 (#756)
* Update module go-redis/redis/v8 to v8.4.7 (#758)
* Update module go-redis/redis/v8 to v8.4.8 (#759)
* Update module lib/pq to v1.9.0 (#717)
* Update module magefile/mage to v1.11.0 (#754)
* Update module mattn/go-sqlite3 to v1.14.5 (#711)
* Update module mattn/go-sqlite3 to v1.14.6 (#751)
* Update module pquerna/otp to v1.3.0 (#705)
* Update module prometheus/client_golang to v1.9.0 (#735)
* Update module spf13/afero to v1.5.0 (#724)
* Update module spf13/afero to v1.5.1 (#730)
* Update module src.techknowlogick.com/xgo to v1.2.0+1.15.6 (#720)
* Update module src.techknowlogick.com/xormigrate to v1.4.0 (#700)
* Update module swaggo/swag to v1.6.9 (#694)
* Update module swaggo/swag to v1.7.0 (#721)
* Update module ulule/limiter/v3 to v3.8.0 (#699)
* Update nfpm config for nfpm v2
* Use db sessions everywere (#750)

## [0.15.1] - 2020-10-20

### Fixed

* Fix not possible to create tasks if metrics were enabled

## [0.15.0] - 2020-10-19

### Added

* Add app support info for DAV (#692)
* Add better tests for namespaces
* Add caldav enabled/disabled to /info endpoint
* Add checks if tasks exist in maps before trying to access them
* Add config option to force ssl connections to connect with the mailer
* Add dav proxy directions to example proxy configurations
* Add docs about using vikunja with utf-8 characters
* Add FreeBSD guide to installation docs
* Add github sponsor link
* Add Golangci Lint (#676)
* Add mage command to create a new migration
* Add option to configure legal urls
* Add rootpath to deb command to not include everything in the deb file
* Add toc to docs
* Add update route to toggle team member admin status
* Add util function to move files
* Disable gocyclo for migration modules
* Favorite lists (#654)
* Favorite tasks (#653)
* Generate config docs from sample config (#684)
* Kanban bucket limits (#652)
* Key-Value Storages (#674)
* Manage users via cli (#632)
* Mention client_max_body_size in nginx proxy settings
* More avatar providers (#622)
* Return rights when reading a single item (#626)
* Saved filters (#655)

### Fixed

* Cleanup references to make
* Don't add a subtask to the top level of tasks to not add it twice in the list
* Fetch tasks for caldav lists (#641)
* Fix building for darwin with mage
* Fix creating lists with non ascii characters (#607)
* Fix decoding active users from redis
* Fix dockerimage build
* Fix docs index links
* Fix duplicating a list with background
* "Fix" gocyclo
* Fix loading list background information for uploaded backgrounds
* Fix migrating items with large items from todoist
* Fix nfpm command in drone
* Fix parsing todoist reminder dates
* Fix reading passwords on windows
* Fix release commands in drone
* Fix release trigger
* Fix release trigger in drone
* Fix token renew for link shares
* Fix trigger for pushing release artifacts to drone
* Fix updating team admin status
* Fix upload avatar not working
* Fix users with disabled totp but not enrolled being unable to login
* Makefile: make add EXTRA_GOFLAG to GOFLAGS (#605)
* Make sure built binary files are executable when compressing with upx
* Make sure lists which would have a duplicate identifier can still be duplicated
* Make sure the metrics map accesses only happen explicitly
* Make sure to copy the permissions as well when moving files
* Make sure to only initialize all variables when needed
* Make sure to require admin rights when modifying list/namespace users to be consistent with teams
* Make sure we have git installed when building os packages
* Make sure we have go installed when building os packages (for build step dependencies)
* Only check if a bucket limit is exceeded when moving a task between buckets
* Only try to download attachments from todoist when there is a url
* Pin telegram notification plugin in drone
* Regenerate swagger docs
* Skip directories when moving build release artefacts in drone
* Support absolute iCal timestamps in CalDAV requests (#691)
* Work around tasks with attachments not being duplicated

### Changed

* Replace renovate tokens with env
* Switch s3 release bucket to scaleway
* Switch to mage (#651)
* Testing improvements (#666)
* Update docs with testmail command + reorder
* Update github.com/asaskevich/govalidator commit hash to 29e1ff8 (#639)
* Update github.com/asaskevich/govalidator commit hash to 50839af (#637)
* Update github.com/asaskevich/govalidator commit hash to 7a23bdc (#657)
* Update github.com/asaskevich/govalidator commit hash to df4adff (#552)
* Update github.com/c2h5oh/datasize commit hash to 48ed595 (#644)
* Update github.com/gordonklaus/ineffassign commit hash to e36bfde (#625)
* Update github.com/jgautheron/goconst commit hash to 8f5268c (#658)
* Update github.com/shurcooL/vfsgen commit hash to 0d455de (#642)
* Update golang.org/x/crypto commit hash to 123391f (#619)
* Update golang.org/x/crypto commit hash to 5c72a88 (#640)
* Update golang.org/x/crypto commit hash to 7f63de1 (#672)
* Update golang.org/x/crypto commit hash to 84dcc77 (#678)
* Update golang.org/x/crypto commit hash to 948cd5f (#609)
* Update golang.org/x/crypto commit hash to 9e8e0b3 (#685)
* Update golang.org/x/crypto commit hash to ab33eee (#608)
* Update golang.org/x/crypto commit hash to afb6bcd (#668)
* Update golang.org/x/crypto commit hash to c90954c (#671)
* Update golang.org/x/crypto commit hash to eb9a90e (#669)
* Update golang.org/x/image commit hash to 4578eab (#663)
* Update golang.org/x/image commit hash to a67d67e (#664)
* Update golang.org/x/image commit hash to e162460 (#665)
* Update golang.org/x/image commit hash to e59bae6 (#659)
* Update golang.org/x/sync commit hash to 3042136 (#667)
* Update golang.org/x/sync commit hash to b3e1573 (#675)
* Update module 4d63.com/tz to v1.2.0 (#631)
* Update module fzipp/gocyclo to v0.2.0 (#686)
* Update module fzipp/gocyclo to v0.3.0 (#687)
* Update module getsentry/sentry-go to v0.7.0 (#617)
* Update module go-errors/errors to v1.1.1 (#677)
* Update module go-testfixtures/testfixtures/v3 to v3.4.0 (#627)
* Update module go-testfixtures/testfixtures/v3 to v3.4.1 (#693)
* Update module iancoleman/strcase to v0.1.0 (#636)
* Update module iancoleman/strcase to v0.1.1 (#645)
* Update module iancoleman/strcase to v0.1.2 (#660)
* Update module imdario/mergo to v0.3.10 (#615)
* Update module imdario/mergo to v0.3.11 (#629)
* Update module labstack/echo/v4 to v4.1.17 (#646)
* Update module lib/pq to v1.7.1 (#616)
* Update module lib/pq to v1.8.0 (#618)
* Update module mattn/go-sqlite3 to v1.14.1 (#638)
* Update module mattn/go-sqlite3 to v1.14.2 (#647)
* Update module mattn/go-sqlite3 to v1.14.3 (#661)
* Update module mattn/go-sqlite3 to v1.14.4 (#670)
* Update module prometheus/client_golang to v1.8.0 (#681)
* Update module spf13/afero to v1.3.2 (#610)
* Update module spf13/afero to v1.3.3 (#623)
* Update module spf13/afero to v1.3.4 (#628)
* Update module spf13/afero to v1.3.5 (#650)
* Update module spf13/afero to v1.4.0 (#662)
* Update module spf13/afero to v1.4.1 (#673)
* Update module spf13/cobra to v1.1.0 (#679)
* Update module spf13/cobra to v1.1.1 (#690)
* Update module spf13/viper to v1.7.1 (#620)
* Update module src.techknowlogick.com/xgo to v1.1.0+1.15.0 (#630)
* Update module src.techknowlogick.com/xgo to v1 (#613)
* Update module swaggo/swag to v1.6.8 (#680)
* Update renovate token
* Update src.techknowlogick.com/xgo commit hash to 7c2e3c9 (#611)
* Update src.techknowlogick.com/xgo commit hash to 96de19c (#612)
* update theme
* Update xgo to v1.0.0+1.14.6
* Use db sessions for task-related things (#621)
* Use nfpm to build deb, rpm and apk packages (#689)

## [0.14.1] - 2020-07-07

### Fixed

* Fix creating lists with non ascii characters (#607)
* Fix decoding active users from redis
* Fix parsing todoist reminder dates
* Make sure the metrics map accesses only happen explicitly

### Changed

* Update docs theme

## [0.14.0] - 2020-07-01

### Added

* Add ability to run the docker container with configurable user and group ids
* Add better errors if the sqlite db file is not writable
* Add cache for initial unsplash collection
* Add docker setup guide from start to finish
* Add docs for restore
* Add dump command (#592)
* Add section to full-docker-example.md for Caddy v2 (#595)
* Add go version to version command
* Add list background information when getting all lists
* Add logging if downloading an image from unsplash fails
* Add migration test in drone (#585)
* Add option to disable totp for everyone
* Add plausible to docs
* Add restarting commands to all example docker compose files
* Add seperate docker pipeline for amd64 and arm
* Add test mail command (#571)
* Add todoist migrator to available migrators in info endpoint if it is enabled
* Add unsplash image proxy for images and thumbnails
* Add returning unsplash info when searching
* Don't return all tasks when a user has no lists
* Duplicate Lists (#603)
* Enable upload backgrounds by default
* Generate a random list identifier based on the list title
* List Backgrounds (#568)
* List Background upload (#582)
* Repeat tasks after completion (#587)
* Restore command (#593)
* Sentry integration (#591)
* Todoist Migration (#566)

### Fixed

* Ensure consistent naming of title fields (#528)
* Ensure task dates are in the future if a task has a repeating interval (#586)
* Fix caching of initial unsplash results per page
* Fix case-insensitive task search for postgresql (#524)
* Fix docker manifest build
* Fix docker multiarch build
* Fix docs theme build
* Fix getting unsplash thumbnails for non "photo-*" urls
* Fix migration 20200425182634
* Fix migration 20200516123847
* Fix migration to add position to task
* Fix misspell
* Fix namespace title not being updated
* Fix not loading timezones on all operating systems
* Fix proxying unsplash images (security)
* Fix removing existing sqlite files
* Fix resetting list, label & namespace colors
* Fix searching for unsplash pictures with words that contain a space
* Fix setting a list identifier to empty
* Fix sqlite db not working when creating a new one
* Fix sqlite path in default config
* Fix swagger docs
* Fix updating the index when moving a task
* Prevent crashing when trying to register with an empty payload
* Properly ping unsplash when using unsplash images
* Return errors when dumping
* Set the list identifier when creating a new task

### Changed

* Expose namespace id when querying lists
* Improve getting all namespaces performance (#526)
* Improve memory usage of dump by not loading all files in memory prior to adding them to the zip
* Improve metrics performance
* Load the list when setting a background
* Make the db timezone migration mysql compatible
* Make the `_unix` suffix optional when sorting tasks
* Migrate all timestamps to real iso dates (#594)
* Make sure docker images are only built when tests pass
* Remove build date from binary
* Remove dependencies on build step to speed up test pipeline (#521)
* Remove go mod vendor todo from pr template now that we don't keep dependencies in the repo anymore
* Remove migration dependency to models
* Remove min length for labels, lists, namespaces, tasks and teams
* Remove vendored dependencies
* Reorganize cmd init functions
* Set unsplash empty collection caching to one hour
* Simplify pipeline & add docker manifest step
* Update alpine Docker tag to v3.12 (#573)
* Update and fix staticcheck
* Update dependency github.com/mattn/go-sqlite3 to v1.14.0
* Update github.com/shurcooL/vfsgen commit hash to 92b8a71 (#599)
* Update golang.org/x/crypto commit hash to 279210d (#577)
* Update golang.org/x/crypto commit hash to 70a84ac (#578)
* Update golang.org/x/crypto commit hash to 75b2880 (#596)
* Update module go-redis/redis/v7 to v7.3.0 (#565)
* Update module go-redis/redis/v7 to v7.4.0 (#579)
* Update module go-testfixtures/testfixtures/v3 to v3.3.0 (#600)
* Update module lib/pq to v1.6.0 (#572)
* Update module lib/pq to v1.7.0 (#581)
* Update module prometheus/client_golang to v1.7.0 (#589)
* Update module prometheus/client_golang to v1.7.1 (#597)
* Update module spf13/afero to v1.3.0 (#588)
* Update module spf13/afero to v1.3.1 (#602)
* Update module spf13/cobra to v1 (#511)
* Update module src.techknowlogick.com/xormigrate to v1.2.1 (#574)
* Update module src.techknowlogick.com/xormigrate to v1.3.0 (#590)
* Update module stretchr/testify to v1.6.0 (#570)
* Update module stretchr/testify to v1.6.1 (#580)
* Update module swaggo/swag to v1.6.7 (#601)
* Update src.techknowlogick.com/xgo commit hash to 209a5cf (#523)
* Update src.techknowlogick.com/xgo commit hash to a09175e (#576)
* Update src.techknowlogick.com/xgo commit hash to eeb7c0a (#575)
* update theme
* Update theme
* Update web handler
* Update xorm.io/xorm 1.0.1 -> 1.0.2
* Use the db logger instance for logging migration related stuff

## [0.13.1] - 2020-05-19

### Fixed

* Don't get all tasks if a user has no lists

## [0.13] - 2020-05-12

#### Added

* Add 2fa for authentification (#383)
* Add categories to error docs
* Add changing email for users
* Add community link
* Add configuration options for log level
* Add creating a new first bucket when creating a new list
* Add docs for changing frontend url
* Add endpoint to disable totp auth
* Add endpoint to get the current users totp status
* Add explanation to docs about cors
* Add github token for renovate (#164)
* Add gosec static analysis
* Add moving tasks between lists (#389)
* Add real buckets for tasks which don't have one (#446)
* Add traefik 2 example configuration
* Configure Renovate (#159)
* Kanban (#393)
* Task filters (#243)
* Task Position (#412)

#### Fixed

* Add checking and logging when trying to put a task into a nonexisting bucket
* Fix bucket ID being reset with no need to do so
* Fix creating new things with a link share auth
* Fix dependencies
* Fix gosec in drone
* Fix link share creation & creating admin link shares without admin rights
* Fix moving tasks back into the empty (ID: 0) bucket
* Fix moving tasks in buckets
* Fix not moving its bucket when moving a task between lists
* Fix pagination count for task collection
* Fix parsing array style comparators by query param
* Fix reference to reverse proxies in docs
* Fix removing the last bucket
* Fix replace statements for tail
* Fix team rights not updating for namespace rights
* Fix tests after renaming json fields to snake_case
* Fix total label count when getting all labels (#477)
* Remove setting task bucket to 0
* Task Filter Fixes (#495)

#### Changed

* Change all json fields to snake_case
* Change totp secret datatype from varchar to text
* Update alpine Docker tag to v3.11 (#160)
* Update docs theme
* Update github.com/c2h5oh/datasize commit hash to 28bbd47 (#212)
* Update github.com/gordonklaus/ineffassign commit hash to 7953dde (#233)
* Update github.com/jgautheron/goconst commit hash to cda7ea3 (#228)
* Update github.com/shurcooL/httpfs commit hash to 8d4bc4b (#229)
* Update golang.org/x/crypto commit hash to 056763e (#222)
* Update golang.org/x/crypto commit hash to 06a226f (#504)
* Update golang.org/x/crypto commit hash to 0848c95 (#371)
* Update golang.org/x/crypto commit hash to 3c4aac8 (#419)
* Update golang.org/x/crypto commit hash to 44a6062 (#429)
* Update golang.org/x/crypto commit hash to 4b2356b (#475)
* Update golang.org/x/crypto commit hash to 4bdfaf4 (#438)
* Update golang.org/x/crypto commit hash to 729f1e8 (#458)
* Update golang.org/x/crypto commit hash to a76a400 (#411)
* Update golang.org/x/lint commit hash to 738671d (#223)
* Update module go-redis/redis to v6.15.7 (#234)
* Update module go-redis/redis to v6.15.7 (#290)
* Update module go-redis/redis to v7 (#277)
* Update module go-redis/redis to v7 (#309)
* Update module go-testfixtures/testfixtures/v3 to v3.1.2 (#457)
* Update module go-testfixtures/testfixtures/v3 to v3.2.0 (#505)
* Update module imdario/mergo to v0.3.9 (#238)
* Update module labstack/echo/v4 to v4.1.16 (#241)
* Update module lib/pq to v1.4.0 (#428)
* Update module lib/pq to v1.5.0 (#476)
* Update module lib/pq to v1.5.1 (#485)
* Update module lib/pq to v1.5.2 (#491)
* Update module olekukonko/tablewriter to v0.0.4 (#240)
* Update module prometheus/client_golang to v0.9.4 (#245)
* Update module prometheus/client_golang to v1
* Update module prometheus/client_golang to v1.6.0 (#463)
* Update module spf13/cobra to v0.0.7 (#271)
* Update module spf13/viper to v1.6.2 (#272)
* Update module spf13/viper to v1.6.3 (#291)
* Update module spf13/viper to v1.7.0 (#494)
* Update module stretchr/testify to v1.5.1 (#274)
* Update Renovate Configuration (#161)
* Update src.techknowlogick.com/xgo commit hash to bb0faa3 (#279)
* Update src.techknowlogick.com/xgo commit hash to c43d4c4 (#224)
* Update xorm redis cacher to use the xorm logger instead of a special seperate one
* Update xorm to v1 (#323)

## [0.12] - 2020-04-04

#### Added

* Add support for archiving lists and namespaces (#152)
* Colors for lists and namespaces (#155)
* Add build time to compile flags
* Add proxying gravatar requests for user avatars (#148)
* Add empty avatar provider (#149)
* expand relative path ~/.config/vikunja to $HOME/.config/vikunja **WINDOWS** (#147)
* Show lists as archived if their namespace is archived

#### Fixed

* Workaround for timezones on windows (#151)
* Fix getting one namespace
* Fix getting the authenticated user with caldav
* Fix searching for config in home directories
* Fix updating lists with an identifier

#### Changed

* Change release bucket

## [0.11] - 2020-03-01

### Added

* Add config options for cors handling (#124)
* Add config options for task attachments (#125)
* Add generate as a make dependency for make build
* Add logging for invalid model errors (#126)
* Add more logging to web handler methods
* Add postgres support (#135)
* Add rate limit by ip for non-authenticated routes (#127)
* Better efficency for loading teams (#128)
* Expand relative path ~/.config/vikunja to $HOME/.config/vikunja (#146)
* Task Comments (#138)

### Fixed

* Fix typo in docker-compose example (#140)
* Fix frontend url for wunderlist migration in docs
* Fix inserting task structure with related tasks (#142)
* Fix time zone settings not working in Docker
* Fix updating dates when marking a task as done (#145)
* Make sure the author is returned when creating a new comment
* Remove double user field

### Changed

* Explicitly disable wunderlist migration by default (#141)
* Migration Improvements (#122)
* Refactor User and DB handling (#123)
* Return iso dates for everything date related from the api (#130)
* Update copyright header
* Update theme
* Update xorm to use the new import path (#133)
* Use relative url in .gitmodules (#132)

## [0.10] - 2020-01-19

### Added

* Migration (#120)
* Endpoint to get tasks on a list (#108)
* Sort Order for tasks (#110)
* Add files volume to docker compose docs
* Add motd config option to docs
* Add option to disable registration (#117)
* Add task identifier (#115)
* Add tests for md5 generation (#111)
* Add user token renew (#113)

### Fixed

* Fix new tasks not getting a new task index (#116)
* Fix owner field being null for user shared namespaces (#119)
* Fix passing sort_by and order_by as query path arrays
* Fix sorting tasks by bool values
* Fix task collection tests
* Consistent copyright text in file headers (#112)

### Changed

* Task collection improvements (#109)
* Update copyright year (#118)
* Update docs with a traefik configuration
* Use redis INCRBY and DECRBY when updating metrics values (#121)
* Use utf8mb4 instead of plain utf8 (#114)
* Update docs theme

## [0.9] - 2019-11-24

### Added

* Task Attachments (#104)
* Task Relations (#103)
* Add endpoint to get a single task (#106)
* Add file volume to the docker image
* Added extra depth to logging to correctly show the functions calling the logger in logs
* Added more infos to a link share auth (#98)
* Added percent done to tasks (#102)

### Fixed

* Fix default logging settings (#107)
* Fixed a bug where adding assignees or reminders via an update would re-create them and not respect already inserted
  ones, leaving a lot of garbage
* Fixed a bug where deleting an attachment would cause a nil panic
* Fixed building docs theme
* Fixed error when setting max file size on 32-Bit systems
* Fixed labels being displayed multiple times if they were associated with more than one task (#99)
* Fixed metrics on/off setting
* Fixed migration for task relations
* Fixed not getting all labels when retrieving a list with all tasks
* Fixed panic when using link share and metrics
* Fixed rate limit panic when authenticating with a link share auth token (#97)
* Fixed removing reminders
* Small link share fixes (#96)

### Changed

* Improve pagination (#105)
* Moved `teams_{namespace|list}_*` to `{namespace|list}_teams_*` for better consistency (#101)
* Refactored getting all lists for a namespace (#100)
* Refactored getting task IDs for labels
* Switched default logger to stdout instead of stderr
* update docs theme

### Misc

* Move from markdown lists to Vikunja for roadmap

## [0.8] - 2019-09-01

### Added

* Better Caldav support (#73)
* Added settings for max open/idle connections and max connection lifetime (#74)
* /info endpoint (#85)
* Added http endpoint to list all users on a list (#87)
* Rate limits (#91)
* Sharing of lists via public links (#94)

### Changed

* Reminders now use an extra table (#75)
* Use the username instead of a full user object when adding a user to a team or giving it rights (#76)
* Add the md5-hashed user email to user objects for use with gravatar (#78)
* Use the auth methods to get IDs to avoid unneeded casts
* Better config handling with constants (#83)
* Statically compile templates in the final binary (#84)
* Use longtext instead of varchar(1000) on description fields (#88)
* Logger refactoring (#90)

### Fixed

* Fixed `listID` not being returned in tasks
* Fixed tests (#72)
* Fixed metrics endpoint not working
* Fixed check if the user really exists before updating/deleting its rights (#77)
* Fixed duedate spelling issue (#79)

### Misc

* Integration tests (#71)
* Make sure the version works when building in drone
* Switched to another version of xgo
* Simplified the docker image (#80)
* Update echo (#82)
* Compress binaries after building them (#81)
* Simplify structure by having less files (#86)
* Limit the test pipeline to run only on pull requests (#89)
* GetUser now returns a pointer (#93)
* Refactor ListTask to Task (#92)

## [0.7] - 2019-04-05

### Added

* DB migrations (#67)
* More cli options for Vikunja (#66 #68)
* Use query params to sort tasks instead of url params (#61)
* More config paths (#55)

### Fixed

* Fixed Priority not updating when setting it to 0
* Fixed getting lists by namespace
* Fixed rights check (#70 #62)
* Fixed labels not being queried correctly on tasks
* Fixed bulk update label tasks

### Changed

* Hide a user's email address everywhere (#69)
* Refactored `canRead()` to get the list before checking rights #65
* Let rights methods return errors (#64 #63)
* Improved Swagger docs for label tasks
* Docs improvements (#58)
* Logging Handling (#57)
* Rights performance improvements (#54)

### Misc

* Releases also as Debian packages (#56)

## [0.6] - 2019-01-16

### Added

* Added prometheus endpoint to get metrics (#33)
* More unit tests (#34)
* Tests can now use config files (#36)
* Redoc for swagger ui (#39, #46)
* Start and end dates for tasks (#40)
* Get tasks between a date range (#41)
* Bulk edit for tasks (#42)
* More ci checks (#43)
* Task assignees (#44, #47)
* Task labels (#45, #48)

### Fixed

* Fixed path to get all tasks (echo bug)
* Explicitly get the peudonamespace with all shared lists (#32)
* Properly init tabels Redis
* unexpected EOF when using metrics (#35)
* Task sorting in lists (#36)
* Various user fixes (#38)
* Fixed a bug where updating a list would update it with the same values it had

### Changed

* Simplified list rights check (#50)
* Refactored some structs to not expose unneded values via json (#52)

### Misc

* Updated libraries
* Updated drone to version 1
* Releases are now signed with our pgp key (more info about this
  on [the download page](https://vikunja.io/en/download/)).

## [0.5] - 2018-12-02

### Added

* Shared lists are now shown in a pseudonamespace with all other namespaces, has the ID -1
* Tasks can have multiple reminders
* Tasks can have subtasks. Subtasks are fully-fleged tasks, but not shown in the task list of a list.
* Tasks can have priorities

### Changed

* Validation not so verbose anymore
* [License](https://git.kolaente.de/vikunja/api/src/branch/master/LICENSE) is now GPLv3
* The crudhandler now has its [own repo](https://git.kolaente.de/vikunja/web) - you can use it in your own projects!

## [0.4] - 2018-11-16

#### Added

* Get all tasks for the authenticated user sorted by their due date
* CalDAV support
* Pagination for everything which returns an array
* Search all the things
* More validation for most of the structs
* Improved Swagger docs (available on `/api/v1/swagger`)

## [0.3] - 2018-11-02

### Added

* Password reset
* Email verification when registering

Misc bugfixes and improvements to the build process

## [0.2] - 2018-10-17

## [0.1] - 2018-09-20

