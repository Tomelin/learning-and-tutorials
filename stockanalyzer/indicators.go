package stockanalyzer

import (
	"fmt"
	"github.com/markcheno/go-talib"
	"math" // Required for math.IsNaN, used in getLastValidValue
	"time" // Required for time.Time
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

// DetermineHistoricalTrend calculates the trend for each time period where all three EMAs are valid.
// Trend logic:
// - Uptrend: If EMA7 > EMA21 and EMA21 > EMA49.
// - Downtrend: If EMA7 < EMA21 and EMA21 < EMA49.
// - Undetermined: Otherwise.
// Parameters:
//   ema7: A slice of float64 representing the 7-period EMA.
//   ema21: A slice of float64 representing the 21-period EMA.
//   ema49: A slice of float64 representing the 49-period EMA.
// Returns:
//   A slice of strings ("uptrend", "downtrend", "undetermined") for each valid period.
//   The length of this slice may be shorter than input EMAs due to warm-up periods.
func DetermineHistoricalTrend(ema7, ema21, ema49 []float64) []string {
	var historicalTrends []string

	// Determine the minimum length of the EMA slices to avoid out-of-bounds errors.
	minLength := len(ema7)
	if len(ema21) < minLength {
		minLength = len(ema21)
	}
	if len(ema49) < minLength {
		minLength = len(ema49)
	}

	// Iterate through the EMA data points.
	// Start index must be great enough to have valid EMAs, TA-Lib usually returns NaN or 0 for initial values.
	// The longest EMA period is 49, so we might expect around 48-~50 initial invalid points.
	// A simple way is to check for non-NaN and non-zero values.
	for i := 0; i < minLength; i++ {
		// Ensure all EMAs at index i are valid (not NaN and not zero)
		// go-talib might use 0 for initial unset values.
		if math.IsNaN(ema7[i]) || ema7[i] == 0 ||
			math.IsNaN(ema21[i]) || ema21[i] == 0 ||
			math.IsNaN(ema49[i]) || ema49[i] == 0 {
			// If any EMA is not valid at this point, skip to the next period.
			continue
		}

		currentTrend := "undetermined"
		if ema7[i] > ema21[i] && ema21[i] > ema49[i] {
			currentTrend = "uptrend"
		} else if ema7[i] < ema21[i] && ema21[i] < ema49[i] {
			currentTrend = "downtrend"
		}
		historicalTrends = append(historicalTrends, currentTrend)
	}
	return historicalTrends
}

// DetermineTrend determines the *current* stock trend based on the latest available EMA values.
// It utilizes DetermineHistoricalTrend to get the sequence of trends and returns the last one.
// Parameters:
//   ema7: A slice of float64 representing the 7-period EMA.
//   ema21: A slice of float64 representing the 21-period EMA.
//   ema49: A slice of float64 representing the 49-period EMA.
// Returns:
//   A string: "uptrend", "downtrend", or "undetermined".
//            Returns "undetermined (insufficient data)" if not enough data to determine trend.
func DetermineTrend(ema7, ema21, ema49 []float64) string {
	historicalTrends := DetermineHistoricalTrend(ema7, ema21, ema49)

	if len(historicalTrends) == 0 {
		return "undetermined (insufficient data)"
	}
	// Return the last trend in the historical sequence, which represents the current trend.
	return historicalTrends[len(historicalTrends)-1]
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
//   ema7, ema21, ema49: Slices of float64 for the respective EMAs. These are used by getLastValidValue.
//   rsi: A slice of float64 for the RSI values (e.g., 14-period RSI). This is used by getLastValidValue.
//   trend: A string describing the current trend (e.g., "uptrend", "downtrend"),
//          typically obtained from the updated DetermineTrend function.
//
// Returns:
//   A string formatted as a prompt for an AI model.
//   If indicator values are not available, "N/A" is used.
func FormatDataForGemini(symbol string, currentPrice float64, ema7, ema21, ema49 []float64, rsi []float64, trend string) string {
	var latestEMA7Str, latestEMA21Str, latestEMA49Str, latestRSIStr string

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
	if okRSI {
		latestRSIStr = fmt.Sprintf("%.2f", latestRSI)
	} else {
		latestRSIStr = "N/A"
	}

	prompt := fmt.Sprintf("Stock: %s\n"+
		"Current Price: %.2f\n"+
		"EMA(7): %s\n"+
		"EMA(21): %s\n"+
		"EMA(49): %s\n"+
		"RSI(14): %s\n"+
		"Current Trend: %s\n\n"+
		"Based on this data, what is your buy/sell/hold recommendation and a brief reasoning?",
		symbol, currentPrice, latestEMA7Str, latestEMA21Str, latestEMA49Str, latestRSIStr, trend)

	return prompt
}

// FindPreviousLowBeforeDowntrend identifies a significant low price and its date
// that occurred in a non-downtrend period (uptrend or undetermined)
// immediately preceding the most recent continuous downtrend sequence.
//
// Parameters:
//   lowPrices: Slice of historical low prices for the stock.
//   dates: Slice of corresponding dates for the lowPrices. Must be same length as lowPrices.
//   historicalTrends: Slice of trend strings ("uptrend", "downtrend", "undetermined")
//                     generated by DetermineHistoricalTrend. This slice is shorter than
//                     lowPrices/dates due to EMA warm-up periods.
//
// Returns:
//   price: The minimum low price found in the identified prior period.
//   date: The date of this minimum low price.
//   found: Boolean, true if a qualifying low was found, false otherwise.
//
// Logic Details:
// 1. Data Alignment: Calculates an `offset` to align indices between the `lowPrices`/`dates` slices
//    (which are full length) and the `historicalTrends` slice (which is shorter due to EMA warm-up periods).
//    This `offset` is `len(lowPrices) - len(historicalTrends)`. All accesses to `lowPrices`/`dates`
//    using an index derived from `historicalTrends` must apply this offset.
// 2. Current Downtrend Check: Verifies if the most recent trend in `historicalTrends` is "downtrend".
//    If not, the primary condition for the function's purpose isn't met, so it returns (0.0, time.Time{}, false).
// 3. Downtrend Start Identification: Iterates backwards from the end of `historicalTrends` to find the
//    starting index of the most recent continuous "downtrend" sequence.
// 4. Prior Non-Downtrend Search: Searches backwards from the point just before the identified current
//    downtrend start to find the first contiguous period of "uptrend" or "undetermined" trends.
//    This identifies the relevant period *before* the current downtrend.
// 5. Minimum Low in Prior Period: If a prior non-downtrend period is found, this function maps its
//    start and end indices (from `historicalTrends`) to the corresponding indices in `lowPrices`
//    (using the `offset`). It then iterates through this mapped range in `lowPrices` to find the
//    minimum low price and its associated date.
// 6. Return: Returns the minimum low price, its date, and `true` if all conditions are met and a
//    low is found. Otherwise, returns `(0.0, time.Time{}, false)` for various edge cases
//    (e.g., no current downtrend, no prior non-downtrend period, empty input slices, data length mismatch).
func FindPreviousLowBeforeDowntrend(lowPrices []float64, dates []time.Time, historicalTrends []string) (float64, time.Time, bool) {
	// Basic validation for input slices
	if len(historicalTrends) == 0 || len(lowPrices) == 0 || len(dates) == 0 || len(lowPrices) != len(dates) {
		return 0.0, time.Time{}, false
	}

	// 1. Data Alignment: Calculate offset for trend indices vs price/date indices.
	// historicalTrends is shorter due to EMA warm-up (e.g., longest EMA is 49, so ~48 initial data points are "lost" for trends).
	offset := len(lowPrices) - len(historicalTrends)
	// Ensure offset is non-negative; it implies historicalTrends isn't longer than price data.
	if offset < 0 {
		return 0.0, time.Time{}, false // Should not happen if trends are derived from prices
	}

	// 2. Current Downtrend Check
	// Check if the most recent trend is "downtrend"
	if historicalTrends[len(historicalTrends)-1] != "downtrend" {
		return 0.0, time.Time{}, false
	}

	currentDowntrendStartIndex_InTrends := -1
	for i := len(historicalTrends) - 1; i >= 0; i-- {
		if historicalTrends[i] == "downtrend" {
			currentDowntrendStartIndex_InTrends = i
		} else {
			// The sequence of "downtrend" at the end has been broken
			break 
		}
	}
	// If the loop finished, currentDowntrendStartIndex_InTrends is the start of the current downtrend.
	// If the entire series was downtrend, it will be 0.

	if currentDowntrendStartIndex_InTrends == -1 {
		// This case should technically be caught by the check historicalTrends[len(historicalTrends)-1] != "downtrend"
		// but kept for robustness.
		return 0.0, time.Time{}, false
	}
	
	// 3. Search for Prior Non-Downtrend Period
	priorPeriodEndIndex_InTrends := currentDowntrendStartIndex_InTrends - 1
	priorPeriodStartIndex_InTrends := -1

	if priorPeriodEndIndex_InTrends < 0 {
		// Current downtrend starts at the very beginning of available trend data. No prior period.
		return 0.0, time.Time{}, false
	}

	for i := priorPeriodEndIndex_InTrends; i >= 0; i-- {
		if historicalTrends[i] != "downtrend" { // "uptrend" or "undetermined"
			priorPeriodStartIndex_InTrends = i // Keep updating start as we find more non-downtrend
		} else {
			// End of the non-downtrend block found
			break
		}
	}

	if priorPeriodStartIndex_InTrends == -1 {
		// No prior non-downtrend period found (e.g., all prior trends were "downtrend" or no trends before current)
		return 0.0, time.Time{}, false
	}
	// The prior non-downtrend period runs from priorPeriodStartIndex_InTrends to priorPeriodEndIndex_InTrends (inclusive)

	// 4. Find Minimum Low in Prior Period
	// Map trend indices to lowPrices/dates indices
	mappedStartIndex_InPrices := priorPeriodStartIndex_InTrends + offset
	mappedEndIndex_InPrices := priorPeriodEndIndex_InTrends + offset

	if mappedStartIndex_InPrices < 0 || mappedEndIndex_InPrices >= len(lowPrices) || mappedStartIndex_InPrices > mappedEndIndex_InPrices {
		// Safety check for mapped indices
		return 0.0, time.Time{}, false
	}
	
	minLowPrice := math.MaxFloat64
	var dateOfMinLow time.Time
	foundMinInPrior := false

	for i := mappedStartIndex_InPrices; i <= mappedEndIndex_InPrices; i++ {
		if lowPrices[i] < minLowPrice {
			minLowPrice = lowPrices[i]
			dateOfMinLow = dates[i]
			foundMinInPrior = true
		}
	}

	if !foundMinInPrior {
		// Should not happen if mapped indices are valid and period has at least one day,
		// but as a safeguard.
		return 0.0, time.Time{}, false
	}

	// 5. Return Values
	return minLowPrice, dateOfMinLow, true
}
