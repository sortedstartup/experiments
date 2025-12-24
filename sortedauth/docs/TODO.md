
1. the oidc.newprovider downloads config from `.well-known/openid-configuration`
`provider, err := oidc.NewProvider(ctx, pCfg.IssuerURL)`
So it may fail during first try (constructor) if internet is down or any other reason.
how to handle that gracefully ? shutdown service or just error later during callback and log it?


- metrics
- traces
- test cases
