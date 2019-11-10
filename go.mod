module github.com/ldcicconi/sfox-arbeur

go 1.13

require (
	github.com/ldcicconi/sfox-api-lib v0.0.0-20191105013720-071b742f3394
	github.com/ldcicconi/trading-common v0.0.0-20191106221151-6aebe24e685c
	github.com/ldcicconi/ws-contractor v0.0.0-20191104022244-6373f917fe44
	github.com/shopspring/decimal v0.0.0-20191009025716-f1972eb1d1f5
	github.com/valyala/fastjson v1.4.1
)

replace (
	github.com/ldcicconi/sfox-api-lib => ../sfox-api-lib
	github.com/ldcicconi/trading-common => ../trading-common
	github.com/ldcicconi/ws-contractor => ../ws-contractor
)
