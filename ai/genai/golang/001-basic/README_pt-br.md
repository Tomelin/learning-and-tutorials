# 🚀 **Desvendando a IA com Golang e Google Gemini!** ✨

Estou iniciando uma série prática onde exploro a integração de Golang com o Google Gemini. Abordarei desde a configuração inicial até a construção de uma CLI para análise de logs e eventos do Kubernetes usando o poder do Gemini.

Neste primeiro passo, vamos entender:

- A estrutura de projetos Go para LLMs.
- Como configurar o acesso à API do Gemini (com foco no novo repositório go-genai).
- A criação de um comando CLI simples para interagir com o Gemini, transformando eventos brutos em insights acionáveis de SRE.

Prepare-se para ver como o Go pode ser uma ferramenta poderosa para construir soluções de IA robustas e eficientes!

💡 Curioso para ver o código e os detalhes técnicos? Confere o repositório/artigo completo [Link para o seu GitHub/Artigo Completo Aqui](https://github.com/Tomelin/learning-and-tutorials/blob/main/ai/genai/golang/001-basic)!

## Introdução
Nesse escopo iremos entender de forma básica, como trabalhar com o Golang e o Google Gemini. A proposta é ser prático.

Antes de começarmos, no momento que escrevo esse post (MAI/2025), o Google Gemini está com dois repositórios no Github, sendo que um deles, foi considerado legado há pouco tempo:
https://github.com/google/generative-ai-go (legacy)
https://github.com/googleapis/go-genai (new)

Porém, diversos projetos estão usando o repositório (legacy), o que muda um pouco os métodos e as chamadas, mas veremos os dois casos ao logo dessa série

## A estrutura de código
Nesse artigo e nos próximos irei usar a mesma estrutura de diretórios e de código, para facilitar o entendimento durante essa série.

Estrutura de diretórios:
```
-cmd    #diretório com as inicializações do golang 
--cli   #inicializa o código em cli (command line interface)
--rest  #inicializa o código em http
-config #configurações, que iremos ler do YAML
-internal  #código restritio a essa app
--core  #regras de negocio da app
---entity      #a estrutura das nossas regras
---service     #regra de negocio
---repository  #tudo que for conexão com tipo de dados (db,mq,cache)
--handler @regras mais mais agrangencia, onde colocaremos http,mq, ...
-pkg ## código compartilhado
--storage
---database
---mq
---cache
--llm
---gemini
```

Essa será a estrutura básica do nosso código.  Não irei detalhar o arquivo config, apenas deixarei compartilhado, pois a unica coisa que ele faz, é ler o arquivo YAML, arquivo que contém as configurações da nossa app


## Arquivo de configuração
Nosso arquivo YAML, que terá as configurações de token e demais serviços, terá a seguinte estrutura inicial:
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

Conforme avançarmos, vamos ajustando os valores e paramêtro, pois nesse momento iremos usar apenas o token GEMINI

O link do config.go, [está aqui](https://github.com/Tomelin/learning-and-tutorials/blob/main/ai/genai/golang/001-basic/config/config.go).   O config verifica se existe a variável "PATH_CONFIG" que é correspondente a onde se encontra o arquivo config.yaml.  Caso não exista, irá procurar o arquivo config.yaml na raiz de onde está se executando o projeto.  Caso não encontre em nenhum desse dois lugares, retornará erro

O config, irá retornar um map[string]interface das nossas configurações e cada componente, terá que tratar as suas configs.

## Inicio do projeto
Nessa primeira etapa, iremos configurar o Golang com o Gemini, apenas para responder a uma request e na próxima etapa iremos configurar com sessão, para manter a conversa.

Também configuraremos o cobra-cli.  Faremos algumas sessões da serie, implementando o HTTP.  Nessa primeiro, teremos o HTTP e CLI, para exemplificar o processo, nas próximas, focaremos mais na CLI.

Então, nesse ponto, entendo que você já está com os diretórios criados, conforme mencionado anteriormente.

iniciando o projeto em golang:
```
go mod init github.com/tomelin/learnai
```

Agora irei acessar o diretório cmd, para criar a cli
```
cd cmd
cobra-cli init learnai
mv learnai cli
```

Projeto iniciado!

## Criando o command do CLI
Vamos criar o comando chamado tshot, que vem de troubleshootig.

```
cobra-cli add tshot
```

ao lista o diretório CMD dentro de cli, teremos dois arquivos go:
```
ls  cmd/
root.go
tshot.go
```

Agora vamos colocar as flags necessárias no nosso troubleshooting, iremos usar as seguintes flags:

**question (q)** para alterar o valor padrão se necessário  
**event (e)** o evento que iremos passar

Entendido os parametros, vamos criar os mesmos dentro do arquivo tshot.go na func init().

```
func init() {
	rootCmd.AddCommand(tshotCmd)

	// Cobra supports local flags which will only run when this command
  // Flag com objetivo de definir a Role do prompt a ser criado (opcional) ao executar a cli
	tshotCmd.Flags().StringP("question", "q", "", "Role: Act as an expert Kubernetes SRE Engineer. Focus on cluster health data analysis (Kubernetes, Cilium CNI, Ingress, Nginx, Karpenter, addons). Proactively identify, diagnose, and remediate issues and create a report and propose the solution")
  
  // Flag onde passaremos os logs ou eventos do kubernetes.  Lembra, teremos limites de caracteres , por causa da LLM
	tshotCmd.Flags().StringP("events", "e", "", "kubernetes events or logs")
}
```

Continuando no arquivo tshot.go, configuraremos a variable tshotCmd ficará da seguinte forma:
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

Definindo o import em tshot.go
```
import (
	"log"

	"github.com/spf13/cobra"
	"learn-ai/config"
	"learn-ai/pkg/llms/llm_agent"
	"learn-ai/pkg/llms/llm_gemini"
)
```

## Configurando a LLM Gemini
Vamos escrever o código para que o Gemini, entenda o que estamos pesqusando.  Vamos criar o arquivo  **gemini.go** dentro do seguinte path **pkg/llm/gemini/gemini.go**

Esse arquivo deve ser um pouco extenso, mas ao decorrer da série, vamos complementando o mesmo, onde teremos alguns métodos.

Iniciando com os import, teremos os seguintes packages:
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

Agora definiremos 2 struct e uma interface, conforme o que iremos desenvolvendo, deixarei os comentários:
```
// Struct para definir a API KEY que vem do config.yaml e o Model da LLM que iremos utilizar
type Gemini struct {
	APIKey string `yaml:"api_key" json:"api_key"`
	Model  string `yaml:"model" json:"model"`
}

// Struct que iremos utilizar nos métodos, para facilitar as chamadas
type GenAIConfig struct {
	Gemini Gemini `yaml:"gemini" `
	client *genai.Client
	model  *genai.GenerativeModel
	Role   string `yaml:"role" json:"role"`
}

// Interface, no qual é retornada pela funcã́o NewGeminiAgent e utilizada dentro de tshot.go
type AgentAI interface {
	Close() error
	Short(ctx context.Context, events *string) (*string, error)
}
```

Agora criaremos o método de validação a API Key e do model, que ficará da seguinte forma:
```
func (gai *GenAIConfig) validate() error {

	if gai == nil || gai.Gemini.APIKey == "" {
		return errors.New("genai config cannot be empty or nil")
	}

// Aqui poderemos ter outro models se for o caso
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

Criando a função NewGeminiAgent, que criará a instância da interface
```
// Recebe GenAIConfig, que conterá a API Key e o model e a role
// Podemos perceber que estamos retornando a interface, que contém Close() e Short()
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
		log.Fatalf("Erro ao criar cliente GenAI: %v", err)
	}

	gai.model = gai.client.GenerativeModel(gai.Gemini.Model)
	gai.Role = *query

	return gai, err
}
```

Criada a função que instância a LLM, vamos definir o método Close(), que é o mais simples desse arquivo até agora.
```
func (gai *GenAIConfig) Close() error {
	return gai.client.Close()
}
```

E por útimo e o mais importante, a lógica de passar os logs e eventos do kubernetes e fazer a LLM entender o que estamos solicitando
```
// Perceba, que estamos passando os eventos para o método Short
func (gai *GenAIConfig) Short(ctx context.Context, events *string) (*string, error) {

	if events == nil || *events == "" {
		return nil, errors.New("events cannot be empty or nil")
	}

// iniciando o chart, que foi inicializado o model dentro de NewGeminiAgent
	session := gai.model.StartChat()
  // definindo a role da LLM
	session.History = []*genai.Content{
		{
			Parts: []genai.Part{genai.Text(gai.Role)},
			Role:  "model",
		},
	}

// indicando que o user, está enviando os dados a serem analisados
	inputForLLM := fmt.Sprintf("From %s: %s", "user", *events)

// encaminha a strutura da mensagem para LLM
	resp, err := session.SendMessage(ctx, genai.Text(inputForLLM))
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}

// validando se temos a resposta da LLM
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("no response from model")
	}

// retornando o conteúdo gerado pela LLM
	response := fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0])

	return &response, nil

}

```


Nesse momento, deve estar tudo funcionando


## Executando a CLI
Para executar a CLI e passar os parâmetros para ela, vamos seguir os seguintes passos:

Acesse o diretório cmd/cli
```
cd cmd/cli/ 
```

Vamos ver quais as opções temos na nossa CLI
```
go run main.go tshot --help
```

Por útimo, vamos executar a CLI
```
go run main.go tshot -e "$(kubectl get events -n my-namespace|head -n 10)"
```

Observe as aspas duplas em torno da saída do comando **kubectl get events**, isso se torna importante, por causa dos espaços.

Ao invés de colocar o **head -n 10**, podemos usar o grep, para pegar algum tipo de reason do erro
```
go run main.go tshot -e "$(kubectl get events -n my-namespace|grep CrashLoopBackOff)"
```

Outra forma de utilizar, é pegar o describe
```
go run main.go tshot -e "$(kubectl describe pod my-pod|tac|head -n 10)"
```

Nessa saído, inverti a saída do arquivo e filtrei as 10 primeiras linhas

## Conclusão
Bom, nesse momento, podemos entender um pouco mais do logs e ventos que está sendo reportado.

Nos próximos passos, iremos criar um chat com sessão via CLI e mais adiante, usar o RAG, para aprimorar mais a nossa estrutura