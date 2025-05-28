package stockanalyzer

import (
	"fmt"
	"github.com/markcheno/go-talib"
	"math" // Required for math.IsNaN, used in getLastValidValue
)

// CalculateEMA calculates the Exponential Moving Average (EMA) for a given slice of closing prices and period.
// It uses the EMA function from the go-talib library.
// Parameters:
//   closingPrices: A slice of float64 representing the historical closing prices of a stock.
//   period: An integer representing the period for the EMA calculation (e.g., 7, 21, 49).
// Returns:
//   A slice of float64 representing the EMA values. The initial values in the slice might be
//   zero or NaN until there's enough data to compute the EMA for the given period.
func CalculateEMA(closingPrices []float64, period int) []float64 {
	emaValues := talib.Ema(closingPrices, period)
	return emaValues
}

// getLastValidValue finds the last valid (non-NaN and non-zero) value in a slice of float64.
// TA-Lib functions (like those in go-talib) often return NaN or zero for initial values
// where the indicator cannot yet be calculated due to an insufficient number of preceding data points.
// This helper function is used to extract the most recent, actual calculated indicator value.
// Parameters:
//   data: A slice of float64, typically output from a TA-Lib function.
// Returns:
//   float64: The last valid value found in the slice.
//   bool: True if a valid value was found, false otherwise.
func getLastValidValue(data []float64) (float64, bool) {
	for i := len(data) - 1; i >= 0; i-- {
		// Check for NaN and also for 0, as go-talib might use 0 for initial unset values
		// where other libraries might use NaN.
		if !math.IsNaN(data[i]) && data[i] != 0 {
			return data[i], true
		}
	}
	return 0, false // Return 0 and false if no valid value is found
}

// DetermineTrend determines the stock trend (uptrend, downtrend, or undetermined)
// based on the relative positions of three Exponential Moving Averages (EMAs):
// EMA7 (short-term), EMA21 (medium-term), and EMA49 (long-term).
//
// Trend logic:
// - Uptrend: If EMA7 > EMA21 and EMA21 > EMA49.
// - Downtrend: If EMA7 < EMA21 and EMA21 < EMA49.
// - Undetermined: Otherwise, or if any EMA does not have a valid latest value.
//
// Parameters:
//   ema7: A slice of float64 representing the 7-period EMA.
//   ema21: A slice of float64 representing the 21-period EMA.
//   ema49: A slice of float64 representing the 49-period EMA.
// Returns:
//   A string: "uptrend", "downtrend", or "undetermined".
//            Returns "undetermined (insufficient data)" if not enough data to determine trend.
func DetermineTrend(ema7, ema21, ema49 []float64) string {
	// Get the last valid (most recent) value for each EMA.
	lastEMA7, ok7 := getLastValidValue(ema7)
	lastEMA21, ok21 := getLastValidValue(ema21)
	lastEMA49, ok49 := getLastValidValue(ema49)

	// If any EMA doesn't have a valid recent value, the trend cannot be determined.
	if !ok7 || !ok21 || !ok49 {
		return "undetermined (insufficient data)"
	}

	// Apply trend determination logic based on EMA positions.
	if lastEMA7 > lastEMA21 && lastEMA21 > lastEMA49 {
		return "uptrend"
	} else if lastEMA7 < lastEMA21 && lastEMA21 < lastEMA49 {
		return "downtrend"
	} else {
		return "undetermined" // EMAs are crossed or not in a clear order.
	}
}

// CalculateSMA calculates the Simple Moving Average (SMA) for a given slice of closing prices and period.
// It uses the SMA function from the go-talib library.
// Parameters:
//   closingPrices: A slice of float64 representing the historical closing prices.
//   period: An integer representing the period for the SMA calculation (e.g., 50, 200).
// Returns:
//   A slice of float64 representing the SMA values. Similar to EMA, initial values might be zero/NaN.
func CalculateSMA(closingPrices []float64, period int) []float64 {
	smaValues := talib.Sma(closingPrices, period)
	return smaValues
}

// CalculateRSI calculates the Relative Strength Index (RSI) for a given slice of closing prices and period.
// It uses the RSI function from the go-talib library.
// RSI is a momentum oscillator that measures the speed and change of price movements.
// Parameters:
//   closingPrices: A slice of float64 representing the historical closing prices.
//   period: An integer representing the period for the RSI calculation (commonly 14).
// Returns:
//   A slice of float64 representing the RSI values, typically ranging from 0 to 100.
//   Initial values might be zero/NaN.
func CalculateRSI(closingPrices []float64, period int) []float64 {
	rsiValues := talib.Rsi(closingPrices, period)
	return rsiValues
}

// FormatDataForGemini prepares a formatted string prompt containing stock data and technical indicators,
// intended for analysis by a generative AI model like Google's Gemini.
// The prompt includes the stock symbol, current price, latest values of key EMAs (7, 21, 49),
// latest RSI (assuming a 14-period), and the determined trend.
// It then asks for a buy/sell/hold recommendation and reasoning.
//
// Parameters:
//   symbol: The stock ticker symbol (e.g., "AAPL").
//   currentPrice: The latest closing price of the stock.
//   ema7, ema21, ema49: Slices of float64 for the respective EMAs.
//   rsi: A slice of float64 for the RSI values (e.g., 14-period RSI).
//   trend: A string describing the current trend (e.g., "uptrend", "downtrend").
//
// Returns:
//   A string formatted as a prompt for an AI model.
//   If indicator values are not available, "N/A" is used.
func FormatDataForGemini(symbol string, currentPrice float64, ema7, ema21, ema49 []float64, rsi []float64, trend string) string {
	var latestEMA7Str, latestEMA21Str, latestEMA49Str, latestRSIStr string

	// Retrieve the latest valid values for each indicator, format them as strings, or use "N/A".
	latestEMA7, ok := getLastValidValue(ema7)
	if ok {
		latestEMA7Str = fmt.Sprintf("%.2f", latestEMA7)
	} else {
		latestEMA7Str = "N/A"
	}

	latestEMA21, ok21 := getLastValidValue(ema21)
	if ok21 {
		latestEMA21Str = fmt.Sprintf("%.2f", latestEMA21)
	} else {
		latestEMA21Str = "N/A"
	}

	latestEMA49, ok49 := getLastValidValue(ema49)
	if ok49 {
		latestEMA49Str = fmt.Sprintf("%.2f", latestEMA49)
	} else {
		latestEMA49Str = "N/A"
	}

	latestRSI, okRSI := getLastValidValue(rsi)
	if ok {
		latestRSIStr = fmt.Sprintf("%.2f", latestRSI)
	} else {
		latestRSIStr = "N/A"
	}

	// Construct the prompt string.
	// The prompt includes current data and asks for a recommendation and reasoning.
	// The RSI period (14) is explicitly mentioned, this could be made dynamic if needed.
	prompt := fmt.Sprintf("Stock: %s\n"+
		"Current Price: %.2f\n"+
		"EMA(7): %s\n"+
		"EMA(21): %s\n"+
		"EMA(49): %s\n"+
		"RSI(14): %s\n"+ // Assuming RSI period is 14, a common default.
		"Current Trend: %s\n\n"+
		"Based on this data, what is your buy/sell/hold recommendation and a brief reasoning?",
		symbol, currentPrice, latestEMA7Str, latestEMA21Str, latestEMA49Str, latestRSIStr, trend)

	return prompt
}
