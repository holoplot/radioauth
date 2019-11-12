module github.com/holoplot/sw__radioauth

go 1.13

require (
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/stretchr/testify v1.4.0 // indirect
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413 // indirect
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6
	gopkg.in/square/go-jose.v2 v2.4.0 // indirect
	layeh.com/radius v0.0.0-20190322222518-890bc1058917
)

replace layeh.com/radius => github.com/zonque/radius v0.0.0-20191208182337-a8c0895345cb
