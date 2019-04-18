// Package coinmarketcap Coin Market Cap API client for Go
package coinmarketcap

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sort"

	"github.com/PuerkitoBio/goquery"
	"github.com/anaskhan96/soup"
	"github.com/hhh0pE/go-coinmarketcap/v2/types"
)

var (
	siteURL               = "https://coinmarketcap.com"
	baseURL               = "https://api.coinmarketcap.com/v2"
	coinGraphURL          = "https://graphs2.coinmarketcap.com/currencies"
	globalMarketGraphURL  = "https://graphs2.coinmarketcap.com/global/marketcap-total"
	altcoinMarketGraphURL = "https://graphs2.coinmarketcap.com/global/marketcap-altcoin"
)

// Interface interface
type Interface interface {
	Listings() ([]*types.Listing, error)
	Tickers(options *TickersOptions) ([]*types.Ticker, error)
	Ticker(options *TickerOptions) (*types.Ticker, error)
	TickerGraph(options *TickerGraphOptions) (*types.TickerGraph, error)
	GlobalMarket(options *GlobalMarketOptions) (*types.GlobalMarket, error)
	GlobalMarketGraph(options *GlobalMarketGraphOptions) (*types.MarketGraph, error)
	GlobalAltcoinMarketGraph(options *GlobalAltcoinMarketGraphOptions) (*types.MarketGraph, error)
	Markets(options *MarketsOptions) ([]*types.Market, error)
	Price(options *PriceOptions) (float64, error)
	CoinID(symbol string) (int, error)
	CoinSlug(symbol string) (string, error)
}

// listingsMedia listings response media
type listingsMedia struct {
	Data []*types.Listing `json:"data"`
}

// Listings gets all coin listings
func Listings() ([]*types.Listing, error) {
	url := fmt.Sprintf("%s/listings", baseURL)
	resp, err := makeReq(url)
	var body listingsMedia
	err = json.Unmarshal(resp, &body)
	if err != nil {
		return nil, err
	}
	return body.Data, nil
}

// TickersOptions options for tickers method
type TickersOptions struct {
	Start   int
	Limit   int
	Convert string
	Sort    string
}

// tickerMedia tickers response media
type tickersMedia struct {
	Data     map[string]*types.Ticker `json:"data,omitempty"`
	Metadata struct {
		Timestamp           int64
		NumCryptoCurrencies int    `json:"num_cryptocurrencies,omitempty"`
		Error               string `json:",omitempty"`
	}
}

// Tickers gets ticker information on coins
func Tickers(options *TickersOptions) ([]*types.Ticker, error) {
	var params []string
	if options.Start >= 0 {
		params = append(params, fmt.Sprintf("start=%v", options.Start))
	}
	if options.Limit >= 0 {
		params = append(params, fmt.Sprintf("limit=%v", options.Limit))
	}
	if options.Convert != "" {
		params = append(params, fmt.Sprintf("convert=%v", options.Convert))
	}
	if options.Sort != "" {
		params = append(params, fmt.Sprintf("sort=%v", options.Sort))
	}
	url := fmt.Sprintf("%s/ticker?%s", baseURL, strings.Join(params, "&"))
	resp, err := makeReq(url)
	var body tickersMedia
	err = json.Unmarshal(resp, &body)
	if err != nil {
		return nil, err
	}
	data := body.Data
	var tickers []*types.Ticker
	for _, v := range data {
		tickers = append(tickers, v)
	}

	if body.Metadata.Error != "" {
		return nil, errors.New(body.Metadata.Error)
	}

	sort.Slice(tickers, func(i, j int) bool {
		return tickers[i].Rank < tickers[j].Rank
	})

	return tickers, nil
}

// TickerOptions options for ticker method
type TickerOptions struct {
	Symbol  string
	Convert string
}

type tickerMedia struct {
	Data *types.Ticker `json:"data"`
}

// Ticker gets ticker information about a cryptocurrency
func Ticker(options *TickerOptions) (*types.Ticker, error) {
	var params []string
	if options.Convert != "" {
		params = append(params, fmt.Sprintf("convert=%v", options.Convert))
	}
	id, err := CoinID(options.Symbol)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/ticker/%v?%s", baseURL, id, strings.Join(params, "&"))
	resp, err := makeReq(url)
	if err != nil {
		return nil, err
	}
	var body tickerMedia
	err = json.Unmarshal(resp, &body)
	if err != nil {
		return nil, err
	}
	return body.Data, nil
}

// TickerGraphOptions options for ticker graph
type TickerGraphOptions struct {
	Symbol string
	Start  int64
	End    int64
}

// TickerGraph gets graph data points for a cryptocurrency
func TickerGraph(options *TickerGraphOptions) (*types.TickerGraph, error) {
	slug, err := CoinSlug(options.Symbol)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/%s/%d/%d", coinGraphURL, slug, options.Start*1000, options.End*1000)
	resp, err := makeReq(url)
	if err != nil {
		return nil, err
	}
	var data *types.TickerGraph
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// GlobalMarketOptions options for global data method
type GlobalMarketOptions struct {
	Convert string
}

// globalMedia global data response media
type globalMarketMedia struct {
	Data *types.GlobalMarket `json:"data"`
}

// GlobalMarket gets information about the global market of the cryptocurrencies
func GlobalMarket(options *GlobalMarketOptions) (*types.GlobalMarket, error) {
	var params []string
	if options.Convert != "" {
		params = append(params, fmt.Sprintf("convert=%v", options.Convert))
	}
	url := fmt.Sprintf("%s/global?%s", baseURL, strings.Join(params, "&"))
	resp, err := makeReq(url)
	var body globalMarketMedia
	err = json.Unmarshal(resp, &body)
	if err != nil {
		return nil, err
	}
	return body.Data, nil
}

// GlobalMarketGraphOptions options for global market graph method
type GlobalMarketGraphOptions struct {
	Start int64
	End   int64
}

// GlobalMarketGraph get graph data points of global market
func GlobalMarketGraph(options *GlobalMarketGraphOptions) (*types.MarketGraph, error) {
	url := fmt.Sprintf("%s/%d/%d", globalMarketGraphURL, options.Start*1000, options.End*1000)
	resp, err := makeReq(url)
	if err != nil {
		return nil, err
	}
	var data *types.MarketGraph
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// GlobalAltcoinMarketGraphOptions options for global altcoin market graph method
type GlobalAltcoinMarketGraphOptions struct {
	Start int64
	End   int64
}

// GlobalAltcoinMarketGraph gets graph data points of altcoin market
func GlobalAltcoinMarketGraph(options *GlobalAltcoinMarketGraphOptions) (*types.MarketGraph, error) {
	url := fmt.Sprintf("%s/%d/%d", altcoinMarketGraphURL, options.Start*1000, options.End*1000)
	resp, err := makeReq(url)
	if err != nil {
		return nil, err
	}
	var data *types.MarketGraph
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MarketsOptions options for markets method
type MarketsOptions struct {
	Symbol string
	Slug   string
}

// Markets get market data for a cryptocurrency
func Markets(options *MarketsOptions) ([]*types.Market, error) {
	if options == nil {
		return nil, errors.New("nil options")
	}
	if options.Slug == "" && options.Symbol == "" {
		return nil, errors.New("empty slug and Symbol")
	}

	slug := options.Slug
	if slug == "" {
		var slug_err error
		slug, slug_err = CoinSlug(options.Symbol)
		if slug_err != nil {
			return nil, slug_err
		}
	}
	url := fmt.Sprintf("%s/currencies/%s/#markets", siteURL, slug)
	var markets []*types.Market
	response, err := soup.Get(url)
	if err != nil {
		return nil, err
	}
	rows := soup.HTMLParse(response).Find("table", "id", "markets-table").Find("tbody").FindAll("tr")
	for _, row := range rows {
		var data []string
		var exchangeSlug string
		for ci, column := range row.FindAll("td") {
			attrs := column.Attrs()
			if attrs["data-sort"] != "" {
				data = append(data, attrs["data-sort"])
			} else {
				data = append(data, column.Text())
			}
			if ci == 1 { // td #2
				aAttrs := column.Find("a").Attrs()
				if link, exist := aAttrs["href"]; exist {
					link_parts := strings.Split(link, "/")
					exchangeSlug = link_parts[len(link_parts)-1]
				}
			}
		}
		markets = append(markets, &types.Market{
			Rank:          toInt(data[0]),
			Exchange:      data[1],
			ExchangeSlug:  exchangeSlug,
			Pair:          data[2],
			VolumeUSD:     toFloat(data[3]),
			Price:         toFloat(data[4]),
			VolumePercent: toFloat(data[5]),
			Category:      data[6],
			FeeType:       data[7],
			Updated:       data[8],
		})
	}
	return markets, nil
}

func Exchanges() ([]*types.Exchange, error) {
	var exchanges []*types.Exchange
	for i := 1; i > 0; i++ {
		pageExchanges, err := ExchangesByPage(i)
		if len(pageExchanges) == 0 {
			break
		}
		if err != nil {
			return nil, errors.New("Exchanges error: " + err.Error())
		}
		exchanges = append(exchanges, pageExchanges...)
	}
	return exchanges, nil
}

func ExchangesByPage(page int) ([]*types.Exchange, error) {
	if page == 0 {
		page = 1
	}
	url := fmt.Sprintf("%s/rankings/exchanges/%d", siteURL, page)
	var exchanges []*types.Exchange

	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}

	trElems := doc.Find("table#exchange-rankings tbody tr")
	for tri := 0; tri < trElems.Length(); tri++ {
		tr := trElems.Eq(tri)

		var data []string
		var exchangeSlug string
		var exchangeLogo string
		tr.Find("td").Each(func(tdi int, selection *goquery.Selection) {
			if val, exist := selection.Attr("data-sort"); exist {
				data = append(data, val)
			} else {
				data = append(data, selection.Text())
			}

			if tdi == 1 {
				if href, exist := selection.Find("a").Attr("href"); exist {
					link_parts := strings.Split(strings.TrimSuffix(href, "/"), "/")
					exchangeSlug = link_parts[len(link_parts)-1]
				}
				if src, exist := selection.Find("img").Attr("src"); exist {
					exchangeLogo = src
				}
			}
		})

		exchanges = append(exchanges, &types.Exchange{
			Rank:            toInt(data[0]),
			Name:            data[1],
			Slug:            exchangeSlug,
			LogoImg:         exchangeLogo,
			AdjustVolume24h: toFloat(data[2]),
			Volume24h:       toFloat(data[3]),
			Volume7d:        toFloat(data[4]),
			Volume30d:       toFloat(data[5]),
			MarketsNumber:   toInt(data[6]),
			Change24h:       toFloat(data[7]) / 100,
			LaunchedAt:      toDate(data[9]),
		})
	}

	return exchanges, nil
}

// PriceOptions options for price method
type PriceOptions struct {
	Symbol  string
	Convert string
}

// Price gets price of a cryptocurrency
func Price(options *PriceOptions) (float64, error) {
	coin, err := Ticker(&TickerOptions{
		Convert: options.Convert,
		Symbol:  options.Symbol,
	})
	if err != nil {
		return 0, err
	}

	if coin == nil {
		return 0, errors.New("coin not found")
	}
	return coin.Quotes[options.Convert].Price, nil
}

// CoinID gets the ID for the cryptocurrency
func CoinID(symbol string) (int, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	listings, err := Listings()
	if err != nil {
		return 0, err
	}

	for _, l := range listings {
		if l.Symbol == symbol {
			return l.ID, nil
		}
	}
	//returns error as default
	return 0, errors.New("coin not found")
}

// CoinSlug gets the slug for the cryptocurrency
func CoinSlug(symbol string) (string, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	coin, err := Ticker(&TickerOptions{
		Symbol: symbol,
	})
	if err != nil {
		return "", err
	}

	if coin == nil {
		return "", errors.New("coin not found")
	}
	return coin.Slug, nil
}

// toInt helper for parsing strings to int
func toInt(rawInt string) int {
	parsed, _ := strconv.Atoi(strings.Replace(strings.Replace(rawInt, "$", "", -1), ",", "", -1))
	return parsed
}

// toFloat helper for parsing strings to float
func toFloat(rawFloat string) float64 {
	parsed, _ := strconv.ParseFloat(strings.Replace(strings.Replace(strings.Replace(rawFloat, "$", "", -1), ",", "", -1), "%", "", -1), 64)
	return parsed
}

func toDate(rawTime string) time.Time {
	parsedTime, _ := time.Parse("2006-01-02", rawTime)
	return parsedTime
}

// doReq HTTP client
func doReq(req *http.Request) ([]byte, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if 200 != resp.StatusCode {
		return nil, fmt.Errorf("%s", body)
	}

	return body, nil
}

// makeReq HTTP request helper
func makeReq(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := doReq(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func fetch(url string) (io.Reader, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if 200 != resp.StatusCode {
		respBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Not 200 status(%d): %s", resp.StatusCode, string(respBytes))
	}
	return resp.Body, nil
}
