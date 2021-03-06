# Gatekeeper
Authorization client and server for tidepool

## 0.4.3 - 2020-10-26
### Engineering
- YLP-241 Review openapi generation so we can serve it through a website

## 0.4.2 - 2020-09-29
### Engineering
- PT-1526 Base gatekeeper image on node:10-alpine

## 0.4.1 - 2020-09-21
### Engineering
- Fix security audit && update to mongo 4.2 

## 0.4.0
### Changed
- PT-1436 Make service start without MongoDb available

## 0.3.0 - 2020-08-04
### Changed
- PT-1277 Integrate Tidepool master for gatekeeper
### Fixed
- PT-1326 Gatekeeper crashes on first query after MongoDb restart
### Engineering
- Review the way the openapi doc build is triggered
- PT-1448 Generate SOUP list

## 0.2.2 - 2020-03-30
### Engineering
- PT-996 Document API as openapi

## 0.2.1 - 2019-11-29
- PT-89 Update dependencies and node version to fix security issues.
  Enable npm audit scan in travis.

## 0.2.0 - 2019-10-29
### Changed
- PT-735 Publish version on the status endpoint

## [0.1.0]
### Changed
- Integrate changes from tidepool [v0.8.0](https://github.com/tidepool-org/gatekeeper/releases/tag/v0.8.0)
- Increase node version to 10.15
