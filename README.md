# elonbot
This is a (mostly complete) bot which buys Dogecoin when Elon tweets about it.  

That trade was actually profitable (depending on your exit strategy) for a while, but it has since run it's course.  Since I don't think I will be making money off of it any time soon, I'm sharing it here for future reference or for anybody who might be interested in seeing how a simple bot like this could work.  There is still work to be done in terms of more robust error handling, logging, sentiment analysis etc.  The current version buys *whenever* Elon tweets about it, including if he were to say:

"I Just sold all of my Dogecoin"

Obviously that would be bad news if we bought.  

This repo is really meant as a starting point - more detailed logic should be added to determine if the tweet has positive or negative sentiment.  Additional improvements beyond sentiment analysis may include things like image processing to identify the presence of a Shiba Inu. In the past, Elon has tweeted memes with no actual text-based reference to Dogecoin, but the mere sight of a Shiba Inu would cause the Doge ~~clowns~~ investors to fevereshly slam the bid.

You must have a `secrets.json` file with Binance and Twitter API Keys in the same directory as the main file with the following key/value pairs:

```json
{
"TwitterAPIKey" : <twitterapikey>,
"TwitterAPISecret": <twitterapisecret>,
"TwitterAccessToken": <twitteraccesstoken>,
"TwitterSecret": <twittersecret>,
"BinanceAPIKey": <binancekey>,
"BinanceAPISecret": <binancesecret>
}
```

I strongly recommend you use the Binance test api before trying this for real.  To do so, get test api keys and replace the last two entries above with `BinanceTestAPIKey` and `BinanceTestAPISecret`.
