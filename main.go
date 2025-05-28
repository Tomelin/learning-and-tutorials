package main

import (
	"fmt"
	"log"
	"os"
	// Import the stockanalyzer package, assuming the module is 'stockanalyzer'
	// and the package directory 'stockanalyzer' is at the root of the module.
	"stockanalyzer/stockanalyzer"
	"time"
	"math" // For math.IsNaN, used by helper functions
)

// getLastValidPrice finds the last valid (non-NaN, non-zero) price from a slice of prices.
// This is useful as historical data might have trailing zero or NaN values.
func getLastValidPrice(prices []float64) (float64, bool) {
	for i := len(prices) - 1; i >= 0; i-- {
		if !math.IsNaN(prices[i]) && prices[i] != 0 {
			return prices[i], true
		}
	}
	return 0, false
}

// getLastValidIndicator finds the last non-NaN and non-zero indicator value.
// getLastValidIndicator finds the last valid (non-NaN, non-zero) value from a slice of indicator data.
// TA-Lib functions (like those in go-talib) can return NaN or zero for initial periods
// where the indicator cannot yet be calculated.
func getLastValidIndicator(data []float64) (float64, bool) {
	for i := len(data) - 1; i >= 0; i-- {
		// Check for NaN and also for 0, as go-talib might use 0 for initial unset values.
		if !math.IsNaN(data[i]) && data[i] != 0 {
			return data[i], true
		}
	}
    return 0, false
}


func main() {
	// --- 1. Define Parameters ---
	// Stock symbol to analyze (e.g., "AAPL", "GOOGL"). Can be changed by the user.
	stockSymbol := "AAPL"
	// Start date for historical data.
	startDate := "2023-01-01"
	// End date for historical data. Uses yesterday's date to ensure recent data.
	endDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	// Period for RSI calculation (commonly 14).
	rsiPeriod := 14
	// Period for SMA calculation (e.g., 50, 200).
	smaPeriod := 50

	fmt.Printf("Analyzing %s from %s to %s\n", stockSymbol, startDate, endDate)

	// --- 2. Get Tiingo API Key ---
	// The Tiingo API key is required to fetch stock data.
	// It must be set as an environment variable.
	tiingoAPIKey := os.Getenv("TIINGO_API_KEY")
	if tiingoAPIKey == "" {
		log.Fatal("Error: TIINGO_API_KEY environment variable not set. Please set it and try again.")
		return
	}

	// --- 3. Fetch Historical Data ---
	// Calls the FetchHistoricalData function from the stockanalyzer package.
	quoteData, err := stockanalyzer.FetchHistoricalData(stockSymbol, startDate, endDate, tiingoAPIKey)
	if err != nil {
		log.Fatalf("Error fetching historical data for %s: %v", stockSymbol, err)
		return
	}

	// Validate that data was actually fetched.
	if len(quoteData.Close) == 0 {
		log.Fatalf("No closing price data fetched for %s. Check symbol, date range, or API key validity.", stockSymbol)
		return
	}
	// The go-quote library populates various fields; we primarily use quoteData.Close ([]float64).

	// --- 4. Calculate Technical Indicators ---
	closingPrices := quoteData.Close

	// Ensure there's enough data for calculations, checking against the longest period used.
	// This prevents errors/panics in TA-Lib functions if insufficient data is provided.
	if len(closingPrices) < smaPeriod { // Longest period used by SMA
		log.Fatalf("Insufficient data for SMA(%d) calculation. Need at least %d data points, got %d.", smaPeriod, smaPeriod, len(closingPrices))
		return
	}
	if len(closingPrices) < rsiPeriod { // RSI period
		log.Fatalf("Insufficient data for RSI(%d) calculation. Need at least %d data points, got %d.", rsiPeriod, rsiPeriod, len(closingPrices))
		return
	}
	// EMA periods (7, 21, 49) are typically shorter than SMA(50), so this check covers them.

	// Calculate various EMAs, SMA, and RSI using functions from the stockanalyzer package.
	ema7 := stockanalyzer.CalculateEMA(closingPrices, 7)
	ema21 := stockanalyzer.CalculateEMA(closingPrices, 21)
	ema49 := stockanalyzer.CalculateEMA(closingPrices, 49)
	sma50 := stockanalyzer.CalculateSMA(closingPrices, smaPeriod)
	rsi14 := stockanalyzer.CalculateRSI(closingPrices, rsiPeriod)

	// --- 5. Determine Trend ---
	// Uses the calculated EMAs to determine the current market trend.
	trend := stockanalyzer.DetermineTrend(ema7, ema21, ema49)

	// --- 6. Output Results ---
	// Display the analyzed stock, its latest price, and key indicators.
	fmt.Printf("\n--- Results for %s ---\n", stockSymbol)

	latestClosingPrice, okPrice := getLastValidPrice(closingPrices)
	if !okPrice {
		// This should ideally not happen if initial data checks passed and data was fetched.
		log.Fatalf("Could not retrieve a valid latest closing price for %s.", stockSymbol)
		return
	}
	fmt.Printf("Latest Closing Price: %.2f\n", latestClosingPrice)

	// Retrieve the latest valid values for each indicator for display.
	latestEMA7, okEMA7 := getLastValidIndicator(ema7)
	latestEMA21, okEMA21 := getLastValidIndicator(ema21)
	latestEMA49, okEMA49 := getLastValidIndicator(ema49)
	latestSMA50, okSMA50 := getLastValidIndicator(sma50)
	latestRSI14, okRSI14 := getLastValidIndicator(rsi14)

	// Print formatted indicator values.
	fmt.Printf("EMA(7): %s\n", formatIndicatorValue(latestEMA7, okEMA7))
	fmt.Printf("EMA(21): %s\n", formatIndicatorValue(latestEMA21, okEMA21))
	fmt.Printf("EMA(49): %s\n", formatIndicatorValue(latestEMA49, okEMA49))
	fmt.Printf("SMA(%d): %s\n", smaPeriod, formatIndicatorValue(latestSMA50, okSMA50))
	fmt.Printf("RSI(%d): %s\n", rsiPeriod, formatIndicatorValue(latestRSI14, okRSI14))
	fmt.Printf("Determined Trend: %s\n", trend)

	// --- 7. Advanced Validation: Previous Low Before Downtrend ---
	// This section performs an additional validation if the current trend is "downtrend".
	// It tries to find a significant low price that occurred in a non-downtrend period
	// immediately preceding the current downtrend.
	if trend == "downtrend" {
		// First, obtain the historical trend data for all valid periods.
		// This is necessary for FindPreviousLowBeforeDowntrend to analyze the sequence of trends.
		historicalTrends := stockanalyzer.DetermineHistoricalTrend(ema7, ema21, ema49)

		// Ensure quoteData.Date and quoteData.Low are available and correctly populated.
		// quote.NewQuoteFromTiingo populates these.
		if len(quoteData.Low) == 0 || len(quoteData.Date) == 0 {
			log.Println("Warning: Low prices or dates data is missing, cannot perform previous low validation.")
		} else if len(quoteData.Low) != len(quoteData.Date) {
			log.Println("Warning: Mismatch between low prices and dates data length, cannot perform previous low validation.")
		} else {
			prevLowPrice, prevLowDate, found := stockanalyzer.FindPreviousLowBeforeDowntrend(quoteData.Low, quoteData.Date, historicalTrends)
			if found {
				fmt.Printf("Validation: Previous significant low of %.2f on %s found. ", prevLowPrice, prevLowDate.Format("2006-01-02"))
				if prevLowPrice > latestClosingPrice {
					fmt.Println("This previous low is ABOVE the current closing price.")
				} else if prevLowPrice < latestClosingPrice {
					fmt.Println("This previous low is BELOW the current closing price.")
				} else {
					fmt.Println("This previous low is EQUAL to the current closing price.")
				}
			} else {
				fmt.Println("Validation: No significant previous low found before the current downtrend.")
			}
		}
	}

	// --- 8. Prepare and Print Gemini Prompt ---
	// Formats the collected data into a prompt suitable for an AI model like Gemini.
	geminiPrompt := stockanalyzer.FormatDataForGemini(stockSymbol, latestClosingPrice, ema7, ema21, ema49, rsi14, trend)

	fmt.Println("\n--- Gemini Prompt ---")
	fmt.Println(geminiPrompt)
	// TODO: Implement Gemini API call here with the prompt above.
	// This section is a placeholder for the user to integrate with a generative AI service.
}

// formatIndicatorValue is a helper function to format a float64 indicator value for printing.
// If the indicator value is not available (ok=false), it returns "N/A".
func formatIndicatorValue(value float64, ok bool) string {
	if !ok {
		return "N/A"
	}
	return fmt.Sprintf("%.2f", value)
}
