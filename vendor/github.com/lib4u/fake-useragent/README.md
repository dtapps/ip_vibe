# fake-useragent
<p align="center">
  <img width="400" height="250" src="https://github.com/user-attachments/assets/4a28c3fe-4e9d-4a4f-a677-9ea454fcdd34">
</p>


Up-to-date simple useragent faker with real world database in Golang

## Features

- Data is pre-downloaded & post-processed from [Intoli LLC](https://github.com/intoli/user-agents/tree/main/src) and the data is part of the package itself
- The most up-to-date database if user-agents
- The data consists of a wide range of browser agents and various browsers
- Retrieves user-agent strings (both of type: `desktop`, `tablet` and/or `mobile` UAs)
- Retrieve user-agent fron JSON file in Go struct, with fields like `Useragent`, `Percent`, `Type`, `DeviceBrand`, `Browser`, `BrowserVersion`, `Os`, `OsVersion` and `Platform`
- This Golang package has the same functionality as popular Python libray "Fake-useragent" and based on same Json user-agents database
- Compatible fake user-agent with Colly Scraping Framework and other.
- Compatible fake user-agent with net/http standard library.
 ```sh
go get github.com/lib4u/fake-useragent
```

### Usage

Simple usage examples


```go
//...
import (
	"fmt"
	app "github.com/lib4u/fake-useragent"
)
//....
ua, err := app.New()
	if err != nil {
		fmt.Println(err)
	}
// Get random user-agent in string
fmt.Println(ua.GetRandom())  // Mozilla/5.0 (iPhone; CPU iPhone OS 18_1_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1.1 Mobile/15E148 Safari/604.1

// Get user-agent string from a specific browser
fmt.Println(ua.Filter().Chrome().Get())
// Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Mobile Safari/537.36

fmt.Println(ua.Filter().Firefox().Get())
//Mozilla/5.0 (Android 14; Mobile; rv:133.0) Gecko/133.0 Firefox/133.0

fmt.Println(ua.Filter().Safari().Get())
//Mozilla/5.0 (iPhone; CPU iPhone OS 18_1_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1.1 Mobile/15E148 Safari/604.1

fmt.Println(ua.Filter().Opera().Get())
//Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36 OPR/114.0.0.0

fmt.Println(ua.Filter().Edge().Get())
//Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0
```

### Advanced Use

You can specify additional user-agent filters
#### Supported next Browsers, OS and platforms in fake-useragent
```go
// Browsers
const (
	Google           = "Google"
	Chrome           = "Chrome"
	Firefox          = "Firefox"
	Edge             = "Edge"
	Opera            = "Opera"
	Safari           = "Safari"
	AndroidBrowser   = "Android"
	YandexBrowser    = "Yandex Browser"
	SamsungInternet  = "Samsung Internet"
	OperaMobile      = "Opera Mobile"
	MobileSafari     = "Mobile Safari"
	FirefoxMobile    = "Firefox Mobile"
	FirefoxiOS       = "Firefox iOS"
	ChromeMobile     = "Chrome Mobile"
	ChromeMobileiOS  = "Chrome Mobile iOS"
	MobileSafariUIWK = "Mobile Safari UI/WKWebView"
	EdgeMobile       = "Edge Mobile"
	DuckDuckGoMobile = "DuckDuckGo Mobile"
	MiuiBrowser      = "MiuiBrowser"
	Whale            = "Whale"
	Twitter          = "Twitter"
	Facebook         = "Facebook"
	AmazonSilk       = "Amazon Silk"
)

// OS
const (
	Windows  = "Windows"
	Linux    = "Linux"
	Ubuntu   = "Ubuntu"
	ChromeOS = "Chrome OS"
	MacOSX   = "Mac OS X"
	Android  = "Android"
	IOS      = "iOS"
)

// Platforms
const (
	Mobile  = "mobile"
	Tablet  = "tablet"
	Desktop = "desktop"
)
```
If you want to specify your own browser list, you can do that via the browsers argument.
This example will only return random user-agents from Firefox and Chrome:

```go
ua, err := app.New()
	if err != nil {
		fmt.Println(err)
	}

// Use filters by browser
fmt.Println(ua.Filter().Browser(app.Firefox, app.Chrome).Get())

```
Also you can use filter by Platform and OS
```go
ua, err := app.New()
	if err != nil {
		fmt.Println(err)
	}
// Random user-agent from a desktop device
fmt.Println(ua.Filter().Platform(app.Desktop).Get())

// Use filters by platform Tablet and Chrome browser
fmt.Println(ua.Filter().Chrome().Platform(app.Tablet).Get())

// Use filters by OS
fmt.Println(ua.Filter().Chrome().Os(app.IOS).Get())

// Use filters by IOS And Windows
fmt.Println(ua.Filter().Chrome().Os(app.IOS, app.Windows).Get())
```

If you need, you can access the entire structure, for example if you need to get the BrowserVersion value or Percent

```go
ua, err := app.New()
	if err != nil {
		fmt.Println(err)
	}

// Get Percent display
fmt.Println(ua.Filter().Browser(app.Firefox, app.Chrome).GetRaw().Percent)

// Get all Go Struct
fmt.Println(ua.Filter().Browser(app.Firefox, app.Chrome).GetRaw())

```

You can set the fallback string using the SetFallback function, in much rare cases something failed:

```go
ua, err := app.New()
	if err != nil {
		fmt.Println(err)
	}
 // Set Fallback string
 ua.SetFallback("user favorite user-agent")
// Use filters by browser
fmt.Println(ua.Filter().Browser(app.Firefox, app.Chrome).Get())

```
