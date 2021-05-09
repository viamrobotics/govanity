# govanity

For nice vanity urls for go and private documentation.

Example @ http://go.viam.com/

## Setup

Specify repos in `etc/modules.txt`, e.g.:

```
vanity.domain.com/private-repo=github.com/org/private-repo
github.com/org/no-vanity-private-repro
```

## Static Doc Generation

Without vanity:

Run `make docs_static`.

With vanity:

Run `VANITY_HOST=<vanity_domain> make docs_static`.
