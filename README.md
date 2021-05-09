# govanity

For nice vanity urls for go and private documentation.

## Setup

Specify repos in `etc/modules.txt`, e.g.:

```
vanity.domain.com/private-repo=github.com/org/private-repo
github.com/org/no-vanity-private-pro
```

## Static Doc Generation

Without vanity:
Run `make docs_static`.

With vanity:
Run `VANITY_HOST=<vanity_domain> make docs_static`.
