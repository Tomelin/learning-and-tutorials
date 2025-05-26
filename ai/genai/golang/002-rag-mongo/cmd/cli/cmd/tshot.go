/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"learn-ai/config"
	"learn-ai/pkg/llm/gemini"
	"log"

	"github.com/spf13/cobra"
)

// tshotCmd represents the tshot command
var tshotCmd = &cobra.Command{
	Use:   "tshot",
	Short: "Analyze events and logs from Kubernetes",
	Long:  `Analyze events and logs from Kubernetes`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(log.LstdFlags | log.Lshortfile)

		query, err := cmd.Flags().GetString("query")
		if err != nil {
			log.Fatal(err.Error())
		}

		events, err := cmd.Flags().GetString("events")
		if err != nil {
			log.Fatal(err.Error())
		}

		configs, err := config.LoadConfig()
		if err != nil {
			log.Fatal(err.Error())
		}

		var genai gemini.GenAIConfig
		b, err := json.Marshal(configs.Fields["genai"])
		if err != nil {
			log.Fatal(err.Error())
		}

		err = json.Unmarshal(b, &genai)
		if err != nil {
			log.Fatal(err.Error())
		}

		if genai == (gemini.GenAIConfig{}) {
			log.Fatalln(errors.New("config cannot be empty"))
		}

		client, err := gemini.NewGeminiAgent(&genai, &query)
		if err != nil {
			log.Fatal(err.Error())
		}

		resp, err := client.Short(context.Background(), &events)
		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Println("************************************")
		fmt.Println("*************RESPONSE***************")
		fmt.Println(*resp)
	},
}

func init() {
	rootCmd.AddCommand(tshotCmd)

	// Cobra supports local flags which will only run when this command
	tshotCmd.Flags().StringP("query", "q", "Role: Act as an expert Kubernetes SRE Engineer. Focus on cluster health data analysis (Kubernetes, Cilium CNI, Ingress, Nginx, Karpenter, addons). Proactively identify, diagnose, and remediate issues and create a report and propose the solution", "Role: Act as an expert Kubernetes SRE Engineer. Focus on cluster health data analysis (Kubernetes, Cilium CNI, Ingress, Nginx, Karpenter, addons). Proactively identify, diagnose, and remediate issues and create a report and propose the solution.")
	tshotCmd.Flags().StringP("events", "e", "", "kubernetes events or logs")
	tshotCmd.Flags().Int8P("agent", "a", 3, "number og agents")
	tshotCmd.Flags().StringP("model", "m", "gemini-2.0-flash", "model for Google Gemini")
}
