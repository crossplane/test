# test

## Overview 

This is the home for automated end-to-end tests for the Crossplane ecosystem. It
contains logic and configuration files for the testing of Crossplane and its
providers. 

It uses Go and Github actions primarily to run tests and automated workflows. To
create a new workflow place a workflow file in the `.github` directory of the
repository. 

Currently, this repo holds the following tests that run on a daily basis:
- [e2e tests against the latest stable release of
  Crossplane](https://github.com/crossplane/test/blob/master/.github/workflows/periodic.yml)
- [provider upgrade
  tests](https://github.com/crossplane/test/blob/master/.github/workflows/provider-upgrade.yml)
- [core crossplane upgrade
  test](https://github.com/crossplane/test/blob/master/.github/workflows/crossplane-upgrade.yml)

You can have a look at actions tab to see all the automated workflow runs and
their status.

### Repo Structure:

- `.github` - contains all the scheduled Github workflows for testing crossplane
  components. 
- `apis` - contains all the API types useful for running Go tests.
- `config` - contains YAML files that use the above mentioned API types.
- `test` - contains all the test scripts with subdirectory name denoting the
  component under consideration. 

## Background and Purpose

As Crossplane and its providers have grown and evolved, the surface area for
potential bugs has increased as well. Unit tests and automated integration tests
at the provider level are effective for determining if the provider can
successfully be installed into a Crossplane Kubernetes cluster, but they do not
go so far as to test any of the controllers beyond that they start. This raised
the need for an automated end-to-end test infrastructure. 

The primary objective of this repo is to ensure stable behaviour across
crossplane and to verify how it behaves with other parts of its stack. It is
also important to ensure that upgrades of Crossplane and its providers from
stable versions to latest builds are not broken. 

## Contributing

test is a community driven project and we welcome contributions. See the
Crossplane
[Contributing](https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md)
guidelines to get started.

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please
open an [issue](https://github.com/crossplane/test/issues).

## Contact

Please use the following to reach members of the community:

- Slack: Join our [slack channel](https://slack.crossplane.io)
- Forums:
  [crossplane-dev](https://groups.google.com/forum/#!forum/crossplane-dev)
- Twitter: [@crossplane_io](https://twitter.com/crossplane_io)
- Email: [info@crossplane.io](mailto:info@crossplane.io)

## Governance and Owners

test is run according to the same
[Governance](https://github.com/crossplane/crossplane/blob/master/GOVERNANCE.md)
and [Ownership](https://github.com/crossplane/crossplane/blob/master/OWNERS.md)
structure as the core Crossplane project.

## Code of Conduct

test adheres to the same [Code of
Conduct](https://github.com/crossplane/crossplane/blob/master/CODE_OF_CONDUCT.md)
as the core Crossplane project.

## Licensing

test is under the Apache 2.0 license.