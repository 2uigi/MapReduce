package main

import (
	"MapReduce/utils"
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
	"time"
)

// Definisci un array fisso di 100 numeri
/*var fixedNumbers = []int{
	56, 23, 89, 15, 12, 45, 78, 100, 34, 66,
	1, 99, 77, 4, 32, 67, 89, 42, 10, 54,
	32, 21, 43, 18, 64, 9, 22, 35, 29, 30,
	84, 60, 13, 7, 44, 28, 38, 74, 91, 57,
	79, 70, 50, 59, 25, 68, 49, 3, 26, 55,
	17, 61, 20, 72, 16, 36, 14, 41, 51, 62,
	19, 75, 58, 31, 8, 37, 46, 65, 76, 63,
	73, 90, 5, 52, 85, 2, 53, 33, 69, 47,
	80, 81, 27, 23, 40, 56, 18, 48, 82, 83,
	93, 94, 95, 96, 97, 98, 0, 60, 71, 69,
}*/

func generateRandomNumbers() []int {
	// Inizializza il generatore di numeri casuali con un seed basato sull'orario
	rand.Seed(time.Now().UnixNano())

	// Slice per memorizzare i numeri casuali
	fixedNumbers := make([]int, 100)

	// Genera 100 numeri casuali tra 0 e 999
	for i := 0; i < 100; i++ {
		fixedNumbers[i] = rand.Intn(100) // Genera un numero tra 0 e 999
	}

	return fixedNumbers
}

func main() {
	// Connessione al server RPC
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Errore di connessione al server RPC: ", err)
	}
	defer func(client *rpc.Client) {
		err := client.Close()
		if err != nil {

		}
	}(client)

	data := generateRandomNumbers()
	fmt.Println(data)

	// Crea la richiesta con l'array fisso di numeri
	request := &utils.SortRequest{Data: data}

	// Crea la risposta
	response := &utils.ClientResponse{}

	// Invia la richiesta al server
	err = client.Call("Master.HandleRequest", request, response)
	if err != nil {
		log.Fatal("Errore durante la chiamata RPC: ", err)
	}

	fmt.Println("Compito Client Eseguito.\nIl client verrà chiuso")
}
