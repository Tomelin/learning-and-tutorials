package stockanalyzer

import (
	"github.com/markcheno/go-quote"
)

// FetchHistoricalData fetches historical stock data for a given symbol from Tiingo.
// It requires the stock symbol (e.g., "AAPL"), start date ("YYYY-MM-DD"),
// end date ("YYYY-MM-DD"), and a valid Tiingo API key.
//
// Parameters:
//   symbol: The stock ticker symbol.
//   startDate: The start date for the historical data.
//   endDate: The end date for the historical data.
//   tiingoAPIKey: The user's Tiingo API key for authenticating requests.
//
// Returns:
//   quote.Quote: A struct containing the historical stock data (OHLC, Volume, etc.).
//                The `quote.Quote` type is from the "github.com/markcheno/go-quote" library.
//                It typically includes fields like `Date`, `Open`, `High`, `Low`, `Close`, `Volume`.
//   error: An error if fetching data fails (e.g., network issue, invalid API key, symbol not found).
func FetchHistoricalData(symbol, startDate, endDate, tiingoAPIKey string) (quote.Quote, error) {
	// Use the go-quote library's function to fetch data directly from Tiingo.
	// This function handles the HTTP request and parsing of the response.
	stockData, err := quote.NewQuoteFromTiingo(symbol, startDate, endDate, tiingoAPIKey)
	if err != nil {
		// If an error occurs (e.g., API key invalid, symbol not found, network error),
		// return an empty quote.Quote struct and the error.
		return quote.Quote{}, err
	}
	// Return the fetched stock data and a nil error if successful.
	return stockData, nil
}
