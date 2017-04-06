
# bloodtales

## Running Locally

```sh
$ heroku local
```

Your app should now be running on [localhost:5000](http://localhost:5000/).

## Deploying to Heroku

```sh
$ heroku create
$ git push heroku master
$ heroku open
```



##########################################################
                         MAP
##########################################################
var romanNumeralDict = map[int]string{
  1000: "M",
  900 : "CM",
  500 : "D",
  400 : "CD",
  100 : "C",
  90  : "XC",
  50  : "L",
  40  : "XL",
  10  : "X",
  9   : "IX",
  5   : "V",
  4   : "IV",
  1   : "I",
}
##########################################################



##########################################################
                         REDIS
##########################################################
import 	"gopkg.in/redis.v3"

var (
	//Client for the database connection
	client *redis.Client
)

func connect() {
	var resolvedURL = os.Getenv("REDIS_URL")
	var password = ""
	if !strings.Contains(herokuURL, "localhost") {
		parsedURL, _ := url.Parse(herokuURL)
		password, _ = parsedURL.User.Password()
		resolvedURL = parsedURL.Host
	}
	fmt.Printf("connecting to %s", herokuURL)
	client = redis.NewClient(&redis.Options{
		Addr:     resolvedURL,
		Password: password,
		DB:       0, // use default DB
	})
}
##########################################################
