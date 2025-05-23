# 🚀 **Unveiling AI with Golang and Google Gemini!** ✨
I'm starting a practical series where I'll explore the integration of Golang with Google Gemini. I'll cover everything from initial setup to building a CLI for Kubernetes log and event analysis using the power of Gemini.

In this first step, we'll delve into:

- The Go project structure for LLMs.
- How to configure access to the Gemini API (focusing on the new go-genai repository).
- Creating a simple CLI command to interact with Gemini, transforming raw events into actionable SRE insights.

Get ready to see how Go can be a powerful tool for building robust and efficient AI solutions!

💡 Curious to see the code and technical details? Check out the full repository/article [Link para o seu GitHub/Artigo Completo Aqui](https://github.com/Tomelin/learning-and-tutorials/blob/main/ai/genai/golang/001-basic)!

## Introduction
In this scope, we'll gain a basic understanding of how to work with Golang and Google Gemini. The aim here is to be practical.

Before we begin, as of this writing (May 2025), Google Gemini has two repositories on GitHub. One of them was recently marked as legacy:

https://github.com/google/generative-ai-go (legacy)
https://github.com/googleapis/go-genai (new)
However, many projects are still using the legacy repository, which slightly changes the methods and calls. We'll explore both cases throughout this series.

## Code Structure
In this article and the ones to follow, I'll be using the same directory and code structure to make it easier to understand throughout this series.

Here's the directory structure:
```
-cmd         # directory for Go initializations
--cli        # initializes the code for CLI (command-line interface)
--rest       # initializes the code for HTTP
-config      # configurations, which we'll read from YAML
-internal    # code restricted to this app
--core       # app business rules
---entity    # the structure of our rules
---service   # business logic
---repository# everything related to data connections (db, mq, cache)
--handler    # broader rules, where we'll put http, mq, etc.
-pkg         # shared code
--storage
---database
---mq
---cache
--llm
---gemini
```

This will be the basic structure of our code. I won't go into detail about the config file; I'll just share it as is, since its only purpose is to read the YAML file containing our app's configurations.


## Configuration File
Our **YAML** file, which will hold token configurations and other service settings, will have the following initial structure:
config.yaml
```
config:
  teste: "string"
genai:
  llm:
    gemini:
      api_key: {TOKEN}
      model: "gemini-2.0-flash"
webservice:
  port: 8080
datatase:
  user: root
  password: root
  host: localhost
```
As we progress, we'll adjust the values and parameters, as right now we'll only be using the **Gemini token**.

The link to the config.go [file is here](https://github.com/Tomelin/learning-and-tutorials/blob/main/ai/genai/golang/001-basic/config/config.go). The config checks for a variable named **"PATH_CONFIG"** which specifies the location of the config.yaml file. If it doesn't exist, it will look for config.yaml in the root directory where the project is being executed. If it's not found in either of these locations, an error will be returned.

The config will return a map[string]interface{} of our configurations, and each component will need to handle its own specific settings.

## Project Kickoff
In this first stage, we'll set up **Golang** with **Gemini** just to handle a single request. In the next step, we'll configure it with sessions to maintain conversation context.

We'll also set up cobra-cli. While we'll implement HTTP in a few sessions of this series, for this initial one, we'll include both HTTP and CLI to illustrate the process. In subsequent sessions, we'll focus more on the CLI.

At this point, I'm assuming you've already created the directories as mentioned previously.

Let's start the Golang project:
```
go mod init learnai
```

Now, I'll access the **cmd** directory to create the CLI.
```
cd cmd
cobra-cli init learnai
mv learnai cli
```

Project initiated!

## Creating the CLI Command
Let's create a command called tshoot, short for troubleshooting.

```
cobra-cli add tshot
```

When listing the **cmd** directory inside of **cli**, you'll find two Go files:
```
ls  cmd/
root.go
tshot.go
```

Now, let's add the necessary flags to our troubleshooting command. We'll use the following flags:

**question (q)**: to change the default value if needed  
**event (e)**: the event we'll pass

With the parameters understood, let's create them inside the tshoot.go file within the init() function.

```
func init() {
	rootCmd.AddCommand(tshotCmd)

	// Cobra supports local flags which will only run when this command
  // A flag whose purpose is to define the Role of the prompt to be created (optional) when executing the CLI.
	tshotCmd.Flags().StringP("question", "q", "", "Role: Act as an expert Kubernetes SRE Engineer. Focus on cluster health data analysis (Kubernetes, Cilium CNI, Ingress, Nginx, Karpenter, addons). Proactively identify, diagnose, and remediate issues and create a report and propose the solution")
  
  // A flag where we'll pass the Kubernetes logs or events. Remember, we'll have character limits due to the LLM.
	tshotCmd.Flags().StringP("events", "e", "", "kubernetes events or logs")
}
```

Continuing in the **tshot.go** file, we'll configure the tshotCmd variable as follows:
```
var tshotCmd = &cobra.Command{
	Use:   "tshot",
	Short: "Analyze events and logs from Kubernetes",
	Long: `Analyze events and logs from Kubernetes`,
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
```

Import define at tshot.go file
```
import (
	"log"

	"github.com/spf13/cobra"
	"learn-ai/config"
	"learn-ai/pkg/llms/llm_agent"
	"learn-ai/pkg/llms/llm_gemini"
)
```

## Configuring the Gemini LLM
Let's write the code for Gemini to understand what we're querying. We'll create the file **gemini.go** within the path **pkg/llm/gemini/gemini.go**.

This file will be somewhat extensive, but throughout the series, we'll keep adding to it, incorporating several methods.

Starting with the imports, we'll have the following packages:
```
package gemini

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

```

Now we will define 2 structs and one interface. I will leave comments as we develop them:
```
// Struct to define the API KEY coming from config.yaml and the LLM Model we will use
type Gemini struct {
	APIKey string `yaml:"api_key" json:"api_key"`
	Model  string `yaml:"model" json:"model"`
}

// A struct that we'll use in the methods to simplify calls.
type GenAIConfig struct {
	Gemini Gemini `yaml:"gemini" `
	client *genai.Client
	model  *genai.GenerativeModel
	Role   string `yaml:"role" json:"role"`
}

// An interface returned by the **NewGeminiAgent** function and used within **tshot.go**.
type AgentAI interface {
	Close() error
	Short(ctx context.Context, events *string) (*string, error)
}
```

Now we'll create the API Key and model validation method, which will look like this:
```
func (gai *GenAIConfig) validate() error {

	if gai == nil || gai.Gemini.APIKey == "" {
		return errors.New("genai config cannot be empty or nil")
	}

// Here, we can include other models if applicable.
	switch gai.Gemini.Model {
	case "gemini-2.0-flash":
		gai.Gemini.Model = "gemini-2.0-flash"
	case "gemini-2.0-pro":
		gai.Gemini.Model = "gemini-2.0-pro"
	default:
		gai.Gemini.Model = "gemini-2.0-flash"
	}

	return nil
}
```

Creating the **NewGeminiAgent** function, which will create an instance of the interface.
```
// Receives GenAIConfig, which will contain the API Key, model, and role.
// We can see that we are returning the interface, which contains Close() and Short().
func NewGeminiAgent(config *GenAIConfig, query *string) (AgentAI, error) {

	gai := config
	var err error
	if err = gai.validate(); err != nil {
		return nil, err
	}

	if query == nil || *query == "" {
		return nil, errors.New("query cannot be empty or nil")
	}

	ctx := context.Background()
	gai.client, err = genai.NewClient(ctx, option.WithAPIKey(gai.Gemini.APIKey))
	if err != nil {
		log.Fatalf("error to create GenAI client: %v", err)
	}

	gai.model = gai.client.GenerativeModel(gai.Gemini.Model)
	gai.Role = *query

	return gai, err
}
```

Now that the function instantiating the **LLM** is created, let's define the Close() method, which is the simplest one in this file so far.
```
func (gai *GenAIConfig) Close() error {
	return gai.client.Close()
}
```

And finally, and most importantly, the logic for passing Kubernetes logs and events and making the LLM understand what we are asking for.
```
// Notice that we are passing the events to the Short method.
func (gai *GenAIConfig) Short(ctx context.Context, events *string) (*string, error) {

	if events == nil || *events == "" {
		return nil, errors.New("events cannot be empty or nil")
	}

// Initiating the chat, which was initialized with the model inside NewGeminiAgent.
	session := gai.model.StartChat()
  // Defining the LLM's role.
	session.History = []*genai.Content{
		{
			Parts: []genai.Part{genai.Text(gai.Role)},
			Role:  "model",
		},
	}

// indicating that the user is sending the data to be analyzed
	inputForLLM := fmt.Sprintf("From %s: %s", "user", *events)

// forwards the message structure to the LLM
	resp, err := session.SendMessage(ctx, genai.Text(inputForLLM))
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}

// Validating if we have a response from the LLM.
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("no response from model")
	}

// Returning the content generated by the LLM.
	response := fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0])

	return &response, nil

}

```


At this point, everything should be working!

## Executing the CLI
To run the CLI and pass parameters to it, follow these steps:

Navigate to the **cmd/cli** directory.
```
cd cmd/cli/ 
```

Let's see what options we have in our CLI.
```
go run main.go tshot --help
```

Lastly, let's run the CLI.
```
go run main.go tshot -e "$(kubectl get events -n my-namespace|head -n 10)"
```

Notice the double quotes around the output of the **kubectl get events** command; this is important due to the spaces.

Instead of using **head -n 10**, we can use grep to filter for a specific error reason.
```
go run main.go tshot -e "$(kubectl get events -n my-namespace|grep CrashLoopBackOff)"
```

Another way to use it is by getting the describe output.
```
go run main.go tshot -e "$(kubectl describe pod my-pod|tac|head -n 10)"
```

In this output, I inverted the file's output and filtered the first 10 lines.

## Conclusion
At this point, we can better understand the logs and events being reported.

In the next steps, we'll create a chat with a session via the CLI, and later, we'll use RAG to further enhance our structure.