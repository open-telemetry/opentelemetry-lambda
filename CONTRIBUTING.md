# OpenTelemetry Contributor Guide

Welcome to OpenTelemetry! This document provides general guidance for contributing to 
this codebase. This is a multi-lingual codebase, so sub-directories like `collector` may
provide additional guidance in a CONTRIBUTING.md.
contribute to the code base. Feel free to browse the [open
issues](https://github.com/open-telemetry/opentelemetry-lambda/issues?q=is%3Aissue+is%3Aopen)
and file new ones, all feedback welcome!

## Before you get started

### Code of Conduct

Please make sure to read and observe our [Code of
Conduct](https://github.com/open-telemetry/community/blob/master/code-of-conduct.md).

### Sign the CLA

Before you can contribute, you will need to sign the [Contributor License
Agreement](https://docs.linuxfoundation.org/lfx/easycla/contributors).

### Code attribution

[License information](README.md#License) should be included in all source files where applicable.
Either full or short version of the header should be used as described at [apache.org](http://www.apache.org/foundation/license-faq.html#Apply-My-Software).
It is OK to exclude the year from the copyright notice. For the details on how to apply the copyright,
see the next section.

### Copyright Notices

OpenTelemetry follows [CNCF recommendations](https://github.com/cncf/foundation/blob/master/copyright-notices.md)
for copyright notices. We use "Copyright The OpenTelemetry Authors" notice form.

According to CNCF recommendations if you are contributing third-party code
you will need to [retain the original copyright notice](https://github.com/cncf/foundation/blob/master/copyright-notices.md#dont-change-someone-elses-notice-without-their-permission).

Any contributed third-party code must originally be Apache 2.0-Licensed or must
carry a permisive software license that is compatible when combining with
Apache 2.0 License. At this moment, BSD and MIT are the only
[OSI-approved licenses](https://opensource.org/licenses/alphabetical) known to be compatible.

If you make substantial changes to the third-party code, _prepend_ the contributed
third party file with OpenTelemetry's copyright notice.

If the contributed code is not third-party code and you are the author we
strongly encourage to avoid including your name in the notice and use the
generic "Copyright The OpenTelemetry Authors" notice. See rationale for this
recommendation [here](https://github.com/cncf/foundation/blob/master/copyright-notices.md#why-not-list-every-copyright-holder).

## Pre-requisites (TBD)

* List of tools, languages, and software required to work with the repository (e.g., programming languages, specific frameworks).
* Any version requirements or compatibility notes 

## Workflow (TBD)

* Explanation of PR title conventions
* Explanation of branch naming conventions and commit message formatting (if there is any)

### How To Get PRs Merged

A PR is considered to be **ready to merge** when:

- It has received approval from
  [Approvers](https://github.com/orgs/open-telemetry/teams/lambda-extension-approvers)
  /
  [Maintainers](https://github.com/orgs/open-telemetry/teams/lambda-extension-maintainers).
- Major feedbacks are resolved.
- It has been open for review for at least one working day. This gives people
  reasonable time to review.
- Trivial changes (typo, cosmetic, doc, etc.) don't have to wait for one day.

Any Maintainer can merge the PR once it is **ready to merge**. Note, that some
PRs may not be merged immediately if the repo is in the process of a release and
the maintainers decided to defer the PR to the next release train.

If a PR has been stuck (e.g. there are lots of debates and people couldn't agree
on each other), the owner should try to get people aligned by:

- Consolidating the perspectives and putting a summary in the PR. It is
  recommended to add a link into the PR description, which points to a comment
  with a summary in the PR conversation.
- Tagging subdomain experts (by looking at the change history) in the PR asking
  for suggestion.
- Reaching out to more people on the [CNCF OpenTelemetry Community Lambda Slack
  channel](TBD).
- Stepping back to see if it makes sense to narrow down the scope of the PR or
  split it up.

## Local Run/Build (TBD)

* How to set up and run the project locally.
* Commands for building the project and starting the application.
* Any important files that need to be created or modified.

## Testing (TBD)

* How to run the test suite for the repository.
* Explanation of different types of tests (e.g., unit, integration, or functional).
* Tools and frameworks used for testing

## Community Expectations and Roles

OpenTelemetry is a community project. Consequently, it is wholly dependent on
its community to provide a productive, friendly, and collaborative environment.

- See [Community
  Membership](https://github.com/open-telemetry/community/blob/master/community-membership.md)
  for a list the various responsibilities of contributor roles. You are
  encouraged to move up this contributor ladder as you gain experience.

## Further Help (TBD)

* Details on where contributors can seek assistance:
* Links to Slack, or other communication platforms.
