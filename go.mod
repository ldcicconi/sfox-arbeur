module github.com/ldcicconi/sfox-arbeur

go 1.13

require (
	github.com/ldcicconi/sfox-api-lib v0.0.0-20191215225313-fb927637e16d
	github.com/ldcicconi/trading-common v0.0.0-20191215231423-972966632a40
	github.com/ldcicconi/ws-contractor v0.0.0-20191110170019-88afc346ecef
	github.com/shopspring/decimal v0.0.0-20191130220710-360f2bc03045
	github.com/valyala/fastjson v1.4.1
)

replace (
	github.com/ldcicconi/sfox-api-lib => ../sfox-api-lib
	github.com/ldcicconi/trading-common => ../trading-common
	github.com/ldcicconi/ws-contractor => ../ws-contractor
)
