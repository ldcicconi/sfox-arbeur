package main

import (
	"net/url"
	"os"
	"strings"

	tc "github.com/ldcicconi/trading-common"
	"github.com/shopspring/decimal"
)

var (
	/*
		WS config
	*/
	SFOXURL = url.URL{
		Scheme: "wss",
		Host:   "ws.sfox.com",
		Path:   "ws",
	}

	/*
		Trader Configs
	*/
	// User-defined configs
	profitThresholdBps    = decimal.New(15, 0)
	USDQuotePairMaxAmount = decimal.New(25, 0)

	// SFOX-defined limits
	smartFee                = decimal.New(175, -1) // 17.5bps
	USDQuotePairMinQuantity = decimal.New(1, -3)   // 0.001
	USDQuotePairMinAmount   = decimal.New(5, 0)    // $5
	BTCQuotePairMinQuantity = decimal.New(1, -3)
	BTCQuotePairMinAmount   = decimal.New(1, -3)

	defaultConfigs = []TraderConfig{

		*NewTraderConfig(*tc.NewPair("btcusd"), TradeLimits{
			MinOrderQuantity:   USDQuotePairMinQuantity,
			MaxOrderQuantity:   decimal.New(1, 0),
			MinOrderAmount:     USDQuotePairMinAmount,
			MaxOrderAmount:     USDQuotePairMaxAmount,
			ProfitThresholdBps: profitThresholdBps,
			FeeRateBps:         smartFee,
		}),
		*NewTraderConfig(*tc.NewPair("etcusd"), TradeLimits{
			MinOrderQuantity:   USDQuotePairMinQuantity,
			MaxOrderQuantity:   decimal.New(100, 0),
			MinOrderAmount:     USDQuotePairMinAmount,
			MaxOrderAmount:     USDQuotePairMaxAmount,
			ProfitThresholdBps: profitThresholdBps,
			FeeRateBps:         smartFee,
		}),
		*NewTraderConfig(*tc.NewPair("ethusd"), TradeLimits{
			MinOrderQuantity:   USDQuotePairMinQuantity,
			MaxOrderQuantity:   decimal.New(100, 0),
			MinOrderAmount:     USDQuotePairMinAmount,
			MaxOrderAmount:     USDQuotePairMaxAmount,
			ProfitThresholdBps: profitThresholdBps,
			FeeRateBps:         smartFee,
		}),
		*NewTraderConfig(*tc.NewPair("ltcusd"), TradeLimits{
			MinOrderQuantity:   USDQuotePairMinQuantity,
			MaxOrderQuantity:   decimal.New(100, 0),
			MinOrderAmount:     USDQuotePairMinAmount,
			MaxOrderAmount:     USDQuotePairMaxAmount,
			ProfitThresholdBps: profitThresholdBps,
			FeeRateBps:         smartFee,
		}),
		*NewTraderConfig(*tc.NewPair("bchusd"), TradeLimits{
			MinOrderQuantity:   USDQuotePairMinQuantity,
			MaxOrderQuantity:   decimal.New(100, 0),
			MinOrderAmount:     USDQuotePairMinAmount,
			MaxOrderAmount:     USDQuotePairMaxAmount,
			ProfitThresholdBps: profitThresholdBps,
			FeeRateBps:         smartFee,
		}),
	}
)

func getAPIKeysFromEnv() []string {
	keysString := os.Getenv("SFOX_API_KEYS")
	return strings.Split(keysString, ",")
}

func main() {
	myApp := NewSFOXArbApp(defaultConfigs, getAPIKeysFromEnv())
	myApp.Start()
	forever := make(chan bool)
	<-forever
}

// {"sequence":5,"recipient":"orderbook.sfox.btcusd","timestamp":1572925018490611435,"payload":{"bids":[[9421.25,0.372592,"itbit"],[9420.71,0.10005447,"gemini"],[9420.69,0.01061493,"gemini"],[9420.53,0.02653778,"gemini"],[9420.52,0.0390851,"gemini"],[9420,0.48607006,"gemini"],[9419.79,2,"gemini"],[9419.5,0.00679645,"itbit"],[9419.37,0.771,"gemini"],[9418.5,0.01359435,"itbit"],[9418.07,0.2,"gemini"],[9418,0.09567346,"itbit"],[9416.793,0.15,"bittrex"],[9416.77,0.526,"gemini"],[9415.24,0.35216057,"market1"],[9414.11,1,"gemini"],[9413.5,2,"itbit"],[9413.41,0.01004082,"bitstamp"],[9413.02,0.2,"bitstamp"],[9413,0.92088618,"itbit"],[9412.87,0.00264272,"market1"],[9412.69,0.04,"market1"],[9412.49,3.7,"gemini"],[9412.44,0.01774418,"gemini"],[9412.28,0.74,"bitstamp"],[9412.075,0.39999998,"bittrex"],[9412.074,0.44449795,"bittrex"],[9412.073,5.2025194,"bittrex"],[9412.01,0.01,"market1"],[9411.804,0.37621591,"bittrex"],[9411.76,1,"gemini"],[9411.56,0.00531261,"gemini"],[9411.48,0.03312544,"bitstamp"],[9411.01,0.12,"market1"],[9411.01,0.28,"bitstamp"],[9411,0.5,"market1"],[9411,6.15310275,"bitstamp"],[9410.92,0.3071,"market1"],[9410.9,0.0026,"market1"],[9410.89,0.03,"gemini"],[9410.82,0.06,"market1"],[9410.635,0.2,"bittrex"],[9410.5,2,"itbit"],[9410.41,1,"market1"],[9410.36,2.827,"market1"],[9410,0.11,"market1"],[9409.85,0.08,"market1"],[9409.75,0.0441989,"itbit"],[9409.5,2.76278275,"itbit"],[9409.42,1,"gemini"],[9409.24,0.766,"market1"],[9409.02,1,"market1"],[9408.01,5.10171433,"bitstamp"],[9408,0.05,"bitstamp"],[9407.89,3.7,"bitstamp"],[9407.74,0.0026,"market1"],[9407.64,1,"gemini"],[9407.25,3,"itbit"],[9406.95,0.82,"market1"],[9406.77,1,"market1"],[9406.75,2,"itbit"],[9406.73,4.1042,"gemini"],[9406.01,1,"bitstamp"],[9406,1,"market1"],[9406,0.012,"itbit"],[9405.95,4.4,"market1"],[9405.7,7.4,"gemini"],[9405.35,3.4481745,"bitstamp"],[9405,0.1,"market1"],[9404.9,3,"bittrex"],[9404.6,0.57725983,"market1"],[9404.6,1,"bitstamp"],[9404.59,1.5,"market1"],[9404.58,0.0026,"market1"],[9403.46,0.18008821,"bittrex"],[9403.4,0.01,"market1"],[9403.26,3.585,"market1"],[9402.97,0.01025786,"gemini"],[9402.58,0.00442963,"bitstamp"],[9402.57,2,"bitstamp"],[9402.39,0.69432957,"market1"],[9402.37,1,"bitstamp"],[9402.03,3,"bittrex"],[9402,0.25,"market1"],[9402,1,"bitstamp"],[9401.891,5,"bittrex"],[9401.8,0.1,"market1"],[9401.75,2.7738,"gemini"],[9401.75,0.13264066,"itbit"],[9401.5,0.001,"market1"],[9401.5,5.319,"itbit"],[9401.42,0.0026,"market1"],[9401.42,1.2102,"bitstamp"],[9401.41,0.05,"bitstamp"],[9401.18,2.22,"market1"],[9401.08,5.07453177,"bitstamp"],[9401.01,0.00329751,"market1"],[9401,0.5,"market1"],[9400.73,0.793,"gemini"],[9400.7,1.39,"bitstamp"],[9400.691,0.19700806,"bittrex"],[9400.61,7.4,"bitstamp"],[9400.6,0.0025062,"market1"],[9400.56,5,"bitstamp"],[9400.34,0.05,"bitstamp"],[9400.33,0.00466189,"gemini"],[9400.1,0.01,"market1"],[9400.09,2.376,"bitstamp"],[9400.01,0.15,"market1"],[9400,2.87125197,"market1"],[9400,0.00053305,"bittrex"],[9400,0.3082351,"bitstamp"],[9400,1.94,"itbit"],[9399.95,1,"bitstamp"],[9399.45,0.66408304,"market1"],[9399,0.14,"bitstamp"],[9398.795,10,"bittrex"],[9398.79,25,"bitstamp"],[9398.73,0.3,"bitstamp"],[9398.71,0.05,"bitstamp"],[9398.62,1,"gemini"],[9398.62,1,"market1"],[9398.57,15,"bittrex"],[9398.5,0.01,"market1"],[9398.26,0.0026,"market1"],[9398,0.01,"market1"],[9397.745,0.1748,"bittrex"],[9397.64,3.06,"market1"],[9397.38,9.1,"gemini"],[9397.26,1.38311646,"bitstamp"],[9397.25,10,"bitstamp"],[9396.8,0.01,"market1"],[9396.5,0.4,"bitstamp"],[9396.44,0.1060614,"market1"],[9396.27,0.64152701,"market1"],[9396.18,3.65691772,"bitstamp"],[9396.02,1,"gemini"],[9396,0.02521792,"bittrex"],[9395.86,5.3172,"gemini"],[9395.75,2.89,"itbit"],[9395.52,1,"market1"],[9395.39,0.01,"market1"],[9395.1,0.0026,"market1"],[9395.1,1.452,"bitstamp"],[9394.91,0.0075186,"market1"],[9394.81,0.21288,"gemini"],[9394.69,0.73,"gemini"],[9394.61,0.64922,"market1"],[9394.36,1.1838,"bitstamp"],[9394.01,0.76111371,"market1"]],"asks":[[9415.25,5.22086259,"market1"],[9415.38,2,"market1"],[9415.85,0.8914117,"market1"],[9416.08,5,"market1"],[9417.8,0.7670324,"market1"],[9417.81,0.79213682,"market1"],[9418.53,0.27231076,"market1"],[9418.99,0.2,"market1"],[9419.248,0.02250009,"bittrex"],[9419.249,0.1747,"bittrex"],[9419.28,0.04,"market1"],[9419.31,0.00438205,"bittrex"],[9419.787,0.00151027,"bittrex"],[9420.7,0.06,"market1"],[9421.14,0.66438343,"market1"],[9421.47,3,"bittrex"],[9421.5,3.7403977,"itbit"],[9421.76,2.652,"market1"],[9421.89,1,"market1"],[9422.119,5,"bittrex"],[9422.12,0.174,"market1"],[9422.301,0.0044,"bittrex"],[9422.38,3.7,"market1"],[9422.79,8.9,"bitstamp"],[9422.8,1.98769633,"bitstamp"],[9423.01,0.00270614,"bitstamp"],[9423.02,0.08,"market1"],[9423.32,0.20014929,"gemini"],[9423.33,2,"gemini"],[9423.35,1.8767902,"gemini"],[9423.36,2,"gemini"],[9423.46,0.16296303,"market1"],[9423.47,1,"market1"],[9423.64,1.1552,"gemini"],[9423.66,0.62533129,"market1"],[9423.98,1,"bitstamp"],[9424.16,0.00270581,"bitstamp"],[9424.26,1.99911198,"market1"],[9424.742,0.1500203,"bittrex"],[9424.77,0.09796019,"gemini"],[9424.78,3.5,"gemini"],[9424.82,0.57724567,"market1"],[9424.87,5,"bitstamp"],[9425,0.01,"market1"],[9425.2,3.7,"gemini"],[9425.43,1,"gemini"],[9425.46,1,"market1"],[9425.53,3,"bittrex"],[9425.55,5.3077,"bitstamp"],[9425.78,0.00272311,"bitstamp"],[9426,1.01,"market1"],[9426.2,1.5,"market1"],[9426.213,10,"bittrex"],[9426.32,0.157,"bitstamp"],[9426.5,2,"itbit"],[9426.52,4,"market1"],[9426.63,0.87232241,"market1"],[9426.64,0.07213594,"bitstamp"],[9426.65,25,"bitstamp"],[9426.82,1,"gemini"],[9427.69,10,"bitstamp"],[9427.75,1.27345654,"itbit"],[9428.15,1,"gemini"],[9428.41,8.894,"market1"],[9428.53,0.63639,"market1"],[9428.68,0.143,"bitstamp"],[9429.5,2,"itbit"],[9429.72,2.86357933,"bitstamp"],[9430,0.02878814,"bitstamp"],[9430.05,0.70974247,"market1"],[9430.231,0.37621591,"bittrex"],[9430.56,1,"gemini"],[9430.69,3.85,"market1"],[9430.94,1,"gemini"],[9431.04,0.175,"bitstamp"],[9431.65,15,"bittrex"],[9431.72,0.2927,"market1"],[9431.76,7.4,"market1"],[9432.07,0.05,"bitstamp"],[9432.2,0.00777219,"bitstamp"],[9432.38,1,"gemini"],[9432.45,1,"bitstamp"],[9432.75,2,"itbit"],[9432.9,0.91468334,"market1"],[9433,0.01,"market1"],[9433.11,0.63606,"market1"],[9433.17,0.18080675,"market1"],[9433.4,0.178,"bitstamp"],[9433.67,0.63611208,"bitstamp"],[9433.86,0.05,"bitstamp"],[9434.27,0.05,"bitstamp"],[9434.31,1,"market1"],[9434.48,0.03519325,"bitstamp"],[9434.6,3.7,"market1"],[9434.61,0.05,"bitstamp"],[9434.978,0.2123494,"bittrex"],[9434.979,0.133,"bittrex"],[9434.98,15,"bittrex"],[9435,4.545,"itbit"],[9435.25,0.00678104,"itbit"],[9435.35,0.373,"gemini"],[9435.41,1,"bitstamp"],[9435.44,0.11803397,"gemini"],[9435.55,2.5,"gemini"],[9435.55,2.5,"bitstamp"],[9435.56,7.4,"gemini"],[9435.69,2.55,"gemini"],[9435.76,0.156,"bitstamp"],[9436,0.013561,"itbit"],[9436.01,0.73424318,"market1"],[9436.25,2.0001266,"itbit"],[9436.39,0.4,"bitstamp"],[9436.6,0.001,"market1"],[9436.856,2,"bittrex"],[9436.859,0.5374,"bittrex"],[9436.91,3.27,"bitstamp"],[9437,1.13268849,"itbit"],[9437.25,2.5,"market1"],[9437.34,0.63582222,"bitstamp"],[9437.69,0.63583,"market1"],[9437.95,0.736,"gemini"],[9438.12,0.208,"bitstamp"],[9438.31,1,"bitstamp"],[9438.82,0.01298029,"bitstamp"],[9438.9,0.146,"bitstamp"],[9439.24,1.2101,"bitstamp"],[9439.33,3.2,"market1"],[9439.37,0.70605932,"market1"],[9439.37,0.01341838,"bitstamp"],[9439.39,0.64922,"market1"],[9439.5,4.986,"gemini"],[9439.997,1.7724,"bittrex"],[9440,1.41559929,"gemini"],[9440,0.82570919,"market1"],[9440,0.53309449,"bitstamp"],[9440.01,6.4,"gemini"],[9440.25,3.0502,"itbit"],[9440.27,0.1221806,"market1"],[9440.48,0.161,"bitstamp"],[9440.8,3.1226,"gemini"],[9441,0.63556898,"bitstamp"],[9441.12,0.01479195,"bitstamp"],[9441.55,0.761,"market1"],[9441.78,1.54,"bitstamp"],[9442,0.6,"gemini"],[9442.01,1.5,"gemini"],[9442.23,2.2895,"bitstamp"],[9442.25,0.0068,"itbit"],[9442.27,0.63548,"market1"],[9442.47,0.01,"market1"]],"market_making":{"bids":[[9413.41,0.01004082,"bitstamp"],[9415.24,0.35216057,"market1"],[9416.793,0.15,"bittrex"],[9420.71,0.10005447,"gemini"],[9421.25,0.372592,"itbit"]],"asks":[[9423.32,0.20014929,"gemini"],[9422.79,8.9,"bitstamp"],[9421.5,3.7403977,"itbit"],[9419.248,0.02250009,"bittrex"],[9415.25,5.22086259,"market1"]]},"timestamps":{"gemini":[1572925018164,1572925018166],"bittrex":[1572925018115,1572925018116],"bitstamp":[1572925017968,1572925017968],"itbit":[1572925017486,1572925017630],"market1":[1572925017637,1572925017638]},"lastupdated":1572925018406,"pair":"btcusd","currency":"usd","lastpublished":1572925018452}}
