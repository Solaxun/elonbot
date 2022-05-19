package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

const (
	elon          = 44196397
	michaeljburry = 412833880
	positive      = true
	negative      = false
)

var (
	secrets        map[string]string
	secretsFile, _ = ioutil.ReadFile("secrets.json")
	_              = json.Unmarshal(secretsFile, &secrets) // to initialize secrets before reading below
	config         = oauth1.NewConfig(secrets["TwitterAPIKey"], secrets["TwitterAPISecret"])
	token          = oauth1.NewToken(secrets["TwitterAccessToken"], secrets["TwitterSecret"])
	binanceKey     = secrets["BinanceAPIKey"]
	binanceSecret  = secrets["BinanceAPISecret"]
	httpClient     = config.Client(oauth1.NoContext, token)
	twitterClient  = twitter.NewClient(httpClient)
)

type strategy interface {
	start()
	stop()
}

type dogeElonStrategy struct {
	twitterchan   chan interface{}
	binancechan   chan *binance.WsTradeEvent
	lastprice     float64
	logfile       *os.File
	twitterClient *twitter.Client
	binanceClient *binance.Client
	stopprice     float64
	trailingpct   float64
	qtyheld       int
	tradeon       bool
	maxprice      float64
}

func (d *dogeElonStrategy) start() {
	for {
		select {
		case t := <-d.twitterchan:
			d.handleTwitterMsg(t)
		case b := <-d.binancechan:
			d.handlePriceUpdate(b)
		}
	}
}

func (d *dogeElonStrategy) handlePriceUpdate(b *binance.WsTradeEvent) {
	fmt.Println(b.Symbol, b.Price, b.Quantity, b.Time)
	curprice, err := strconv.ParseFloat(b.Price, 32)
	if err != nil {
		log.Println(err)
		return
	}

	d.lastprice = curprice
	// maxprice initially negative infinity, so first time around will just be curprice
	newmax := math.Max(curprice, d.maxprice)
	if newmax != d.maxprice {
		d.maxprice = newmax
		if d.tradeon {
			d.setTrailingStop(d.maxprice)
		}
	}
	// stop price initialized to negative inf, only updated once a buy completed
	if curprice <= d.stopprice && d.tradeon {
		res, err := d.sellAtMarket("DOGEUSDT", curprice, true)
		fmt.Println(curprice, d.stopprice, d.lastprice, "........")
		if err != nil {
			log.Println(err)
			return
		}
		if res != nil {
			log.Println(res) // TODO: update profit etc. with fills from the response
		}
	}
}

func (d *dogeElonStrategy) handleTwitterMsg(msg interface{}) {
	var tweetText string
	switch msg := msg.(type) {
	case *twitter.Tweet:
		media := msg.Entities.Media
		// fmt.Println(media, "****", msg)
		if extended := msg.ExtendedTweet; extended != nil {
			tweetText = extended.FullText
		} else {
			tweetText = msg.Text
		}
		fmt.Println(msg.User.ScreenName, msg.User.Name, tweetText)
		if len(media) > 0 {
			fmt.Println(media[0].MediaURL)
		}
		if msg.User.ID == elon {
			log.Println("Elon: ", tweetText)
			log.Println("DOGE Price: ", d.lastprice)
			if mentionsDoge(tweetText) && sentiment(tweetText) == positive && !d.tradeon {
				// buy doge with a stop loss equal to entry price and a
				// trailing take profit at max(0.95*curprice, entry price)
				d.yolo("DOGEUSDT", 5_000) //TODO: if this happens b4 1st px received, no qty set
			}

		} else if msg.User.ID == michaeljburry {
			log.Println("Burry: ", tweetText)
		}
	}
}

func mentionsDoge(text string) bool {
	match, _ := regexp.MatchString(`(?i)[Ðd](o|0)ge`, text)
	return match
}

//TODO: actual sentiment analysis, for now assume if he mentions it's good
func sentiment(text string) bool {
	return positive
}

// buy nearest whole integer of given ticker as dollaramt can purchase, rounded down
func (d *dogeElonStrategy) yolo(ticker string, dollaramt int) error {
	px := d.lastprice
	quantityTraded := int(float64(dollaramt) / px)
	// enter quantity as determined by dollaramt param and last price as market order
	res, err := d.buyAtMarket(ticker, px, quantityTraded, true)
	if err != nil {
		log.Println(err)
		return err
	}
	if res != nil {
		px, err = strconv.ParseFloat(res.Price, 64) // what we get filled at - use for profit calc
	}
	if err != nil {
		log.Println(err)
		return err
	}
	d.tradeon = true
	d.qtyheld = quantityTraded
	d.setTrailingStop(px) // only takes effect after buy if not already active
	return nil
}

func (d *dogeElonStrategy) buyAtMarket(ticker string, px float64, quantityTraded int, testorder bool) (*binance.CreateOrderResponse, error) {
	order := d.binanceClient.NewCreateOrderService().Symbol(ticker).
		Side(binance.SideTypeBuy).Type(binance.OrderTypeMarket).
		Quantity(strconv.Itoa(quantityTraded))
	if testorder == true {
		err := order.Test(context.Background())
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println("[Market Buy Completed] - ", quantityTraded, "Dogecoin at a price of ", px)
		return nil, nil
	}
	res, err := order.Do(context.Background())
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("[Market Buy Completed] - ", quantityTraded, "Dogecoin at a price of ", res.Price)
	return res, nil
}

// whenever price hits the trailing stop (trailpct * curprice), trigger a market sell
func (d *dogeElonStrategy) setTrailingStop(price float64) {
	trailprice := (1 - d.trailingpct) * price
	d.stopprice = trailprice
	// don't submit every time or the log will be inundated.. think about how to capture this, log less frequently? log sep file? etc
	fmt.Println("[Trailing Price Updated] - ", "Dogecoin with a trigger price of ", trailprice)
	log.Println("[Trailing Price Updated] - ", "Dogecoin with a trigger price of ", trailprice)

}

func (d *dogeElonStrategy) sellAtMarket(ticker string, px float64, testorder bool) (*binance.CreateOrderResponse, error) {
	order := d.binanceClient.NewCreateOrderService().Symbol(ticker).
		Side(binance.SideTypeSell).Type(binance.OrderTypeStopLossLimit).
		TimeInForce(binance.TimeInForceTypeGTC).
		StopPrice(fmt.Sprintf("%f", px)).
		Price(fmt.Sprintf("%f", px)).
		Quantity(strconv.Itoa(d.qtyheld))
	if testorder == true {
		err := order.Test(context.Background())
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Println("[Market Sell Complete] - ", d.qtyheld, "Dogecoin at a price of ", px, "stop was ", d.stopprice, "max was ", d.maxprice)
		d.tradeon = false
		fmt.Println("d.tradeon is set to: ", d.tradeon)
		d.maxprice, d.stopprice = math.Inf(-1), math.Inf(-1)
		return nil, nil
	}
	res, err := order.Do(context.Background())
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("[Market Sell Complete] - ", d.qtyheld, "Dogecoin at a price of ", px, "stop was ", d.stopprice, "max was ", d.maxprice)
	d.tradeon = false
	fmt.Println("d.tradeon is set to: ", d.tradeon)
	d.maxprice, d.stopprice = math.Inf(-1), math.Inf(-1)
	return res, nil
}

func dogeStreamPrice() (chan *binance.WsTradeEvent, error) {
	binancechan := make(chan *binance.WsTradeEvent)
	wsTradeHandler := func(event *binance.WsTradeEvent) {
		binancechan <- event
	}

	errHandler := func(err error) {
		fmt.Println(err)
		log.Println(err)
	}
	_, _, err := binance.WsTradeServe("DOGEUSDT", wsTradeHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		log.Println(err)
		return nil, err
	}
	return binancechan, nil

}

func main() {
	if len(secrets) == 0 {
		panic("`secrets.json` must exist in this directory.")
	}
	for _, key := range secrets {
		if key == "" {
			panic("some API keys not present in `secrets.json`")
		}
	}

	logfile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer logfile.Close()
	log.SetOutput(logfile)

	binanceClient := binance.NewClient(binanceKey, binanceSecret)
	binanceClient.BaseURL = "https://api.binance.us"

	params := &twitter.StreamFilterParams{
		Follow:        []string{strconv.Itoa(elon), strconv.Itoa(michaeljburry)},
		StallWarnings: twitter.Bool(true),
	}
	stream, err := twitterClient.Streams.Filter(params)
	if err != nil {
		fmt.Println("sum-ting-wong", err)
		return
	}
	binchan, err := dogeStreamPrice() // return chan that receives msgs from callback
	if err != nil {
		fmt.Println(err)
		return
	}

	strategy := &dogeElonStrategy{
		twitterchan:   stream.Messages,
		binancechan:   binchan,
		logfile:       logfile,
		twitterClient: twitterClient,
		binanceClient: binanceClient,
		stopprice:     math.Inf(-1),
		maxprice:      math.Inf(-1),
		trailingpct:   0.003, // decimal form percentage e.g. 0.05 = 5.0%
	}
	testelon := &twitter.Tweet{
		Text: "blah blah blah blah.... Doge",
		ExtendedTweet: &twitter.ExtendedTweet{
			FullText: "blah blah Dogecoin to the moon!!! blah blah",
		},
		Entities: &twitter.Entities{
			Media: []twitter.MediaEntity{},
		},
		User: &twitter.User{
			ScreenName: "@elonmusk", Name: "Fred Durst", ID: elon,
		},
	}
	go func() {
		time.Sleep(time.Duration(time.Second * 2))
		strategy.twitterchan <- testelon
		time.Sleep(time.Duration(time.Minute * 2))
		strategy.twitterchan <- testelon
	}()
	strategy.start()

}
