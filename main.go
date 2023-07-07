package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Struct para a carga útil da mensagem recebida
type MessagePayload struct {
	Object string `json:"object"`
	Entry  []struct {
		Changes []struct {
			Value struct {
				MessagingProduct string `json:"messaging_product"`
				Metadata         struct {
					DisplayPhoneNumber string `json:"display_phone_number"`
					PhoneNumberID      string `json:"phone_number_id"`
				} `json:"metadata"`
				Contacts []struct {
					Profile struct {
						Name string `json:"name"`
					} `json:"profile"`
					WAID string `json:"wa_id"`
				} `json:"contacts"`
				Messages []struct {
					From      string `json:"from"`
					ID        string `json:"id"`
					Timestamp string `json:"timestamp"`
					Type      string `json:"type"`
					Text      struct {
						Body string `json:"body"`
					} `json:"text"`
				} `json:"messages"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

func main() {
	// Carrega as variáveis de ambiente do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env")
	}

	// Obtenha o token de acesso a partir das variáveis de ambiente
	token := os.Getenv("WHATSAPP_TOKEN")

	// Cria o roteador do GIN
	router := gin.Default()

	// Rota POST para o webhook
	router.POST("/webhook", func(c *gin.Context) {
		// Parse a carga útil da mensagem recebida
		var payload MessagePayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			log.Println("Erro ao decodificar a carga útil da mensagem:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao decodificar a carga útil da mensagem"})
			return
		}

		// Imprime a mensagem recebida
		fmt.Println("Mensagem recebida:", payload)

		// Verifique se a mensagem é do WhatsApp
		if payload.Object == "whatsapp_business_account" && len(payload.Entry) > 0 && len(payload.Entry[0].Changes) > 0 {
			change := payload.Entry[0].Changes[0]
			if len(change.Value.Messages) > 0 {
				message := change.Value.Messages[0]

				// Extraia as informações necessárias da mensagem
				phoneID := change.Value.Metadata.PhoneNumberID
				from := message.From
				body := message.Text.Body

				body2 := "qualquer coisa"
				fmt.Println("Variável chave:", body)
				// Envie uma mensagem de resposta de volta para o WhatsApp
				response := map[string]interface{}{
					"messaging_product": "whatsapp",
					"to":                from,
					"text": map[string]string{
						"body": "Ack: " + body2,
					},
				}
				responseJSON, _ := json.Marshal(response)
				url := fmt.Sprintf("https://graph.facebook.com/v12.0/%s/messages?access_token=%s", phoneID, token)
				_, err := http.Post(url, "application/json", strings.NewReader(string(responseJSON)))
				if err != nil {
					log.Println("Erro ao enviar mensagem de resposta:", err)
				}
			}
		}

		c.Status(http.StatusOK)
	})

	// Rota GET para a verificação do webhook
	router.GET("/webhook", func(c *gin.Context) {
		// Obtenha os parâmetros da solicitação de verificação
		verifyToken := os.Getenv("VERIFY_TOKEN")
		mode := c.Query("hub.mode")
		token := c.Query("hub.verify_token")
		challenge := c.Query("hub.challenge")

		// Verifique se o token e o modo estão corretos
		if mode == "subscribe" && token == verifyToken {
			fmt.Println("WEBHOOK_VERIFIED")
			c.String(http.StatusOK, challenge)
		} else {
			c.Status(http.StatusForbidden)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Inicia o servidor HTTP do GIN
	log.Println("Webhook está ouvindo na porta", port)
	log.Fatal(router.Run(":" + port))
}
