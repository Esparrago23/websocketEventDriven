package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var clients = make(map[string]*websocket.Conn)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	var clientId string

	for {
		var msg map[string]interface{}

		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error leyendo mensaje: %v", err)
			delete(clients, clientId)
			break
		}

		log.Printf("Mensaje recibido: %v", msg)

		if msg["type"] == "register" {
			clientId = msg["clientId"].(string)
			clients[clientId] = ws
			fmt.Printf("Cliente registrado: %s\n", clientId)
			continue
		}

		if msg["type"] == "payOrder" {
			log.Printf("Procesando pago: %v", msg)

			orderId := msg["orderId"]
			clientId := msg["clientId"]

			updatedOrder := map[string]interface{}{
				"order_id": orderId,
				"status":   "contratado",
			}

			updateMsg := map[string]interface{}{
				"type":     "orderUpdate",
				"orderId":  orderId,
				"clientId": clientId,
				"order":    updatedOrder,
			}

			if conn, exists := clients[clientId.(string)]; exists {
				err := conn.WriteJSON(updateMsg)
				if err != nil {
					log.Printf("Error enviando mensaje a %s: %v", clientId, err)
					conn.Close()
					delete(clients, clientId.(string))
				} else {
					log.Printf("Mensaje enviado a %s", clientId)
				}
			} else {
				log.Printf("Cliente %s no encontrado", clientId)
			}
		}
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No se pudo cargar el archivo .env, usando valores por defecto")
	}

	port := os.Getenv("WS_PORT")
	

	http.HandleFunc("/ws", handleConnections)

	log.Printf("WebSocket server started on :%s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
