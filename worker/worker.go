package main

import (
	"MapReduce/utils"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Struttura che rappresenta il Worker
type Worker struct {
	id          int            // Identificatore del worker
	data        []int          // Dati finali assegnati al worker
	mu          sync.Mutex     // Mutex per la sincronizzazione dell'accesso ai dati
	waitSwap    sync.WaitGroup // WaitGroup per sincronizzare la fase di scambio
	mapperCount int            // contatore
	numWorkers  int            // per i colleghi reducer, indica il numero di mapper
}

func (w *Worker) ProcessChunk(request *utils.WorkerRequest, response *utils.WorkerResponse) error {
	// Imposta il numero di worker in base alle porte ricevute
	w.numWorkers = len(request.MapperPorts)

	// Determina il ruolo del worker
	switch request.WorkerRole {
	case "Mapper":
		// Processa il ruolo Mapper
		err := w.ProcessMapper(request, response)
		if err != nil {
			return fmt.Errorf("errore durante la fase di mapping: %w", err)
		}

	default:
		// Genera un errore se il ruolo non è riconosciuto
		return fmt.Errorf("ruolo worker non riconosciuto: %s", request.WorkerRole)
	}

	return nil
}

// mapDataToRanges assegna i valori di un chunk ai range specificati.
// ranges: array di range, ciascuno rappresentato come [min, max].
// chunk: array di interi da processare.
// Ritorna una mappa con chiave "indice del range" e valore "array di interi nel range".
func mapDataToRanges(chunk []int, ranges []int) map[int][]int {
	// Mappa risultato
	mappedData := make(map[int][]int)

	// Itera su ogni valore del chunk
	for _, value := range chunk {
		// Trova il range corretto
		for i := 0; i < len(ranges)-1; i++ {
			// Verifica se il valore è nel range [ranges[i], ranges[i+1])
			if value >= ranges[i] && value < ranges[i+1] {
				mappedData[i] = append(mappedData[i], value)
				break
			}
		}
	}
	return mappedData
}

func (w *Worker) ReceiveData(data *utils.WorkerData, ack *bool) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Aggiungi i dati al worker
	w.data = append(w.data, data.Values...)

	// Incrementa il mapperCount e verifica se tutti i mapper hanno inviato i dati

	if w.mapperCount == w.numWorkers {
		fmt.Println("Tutti i mapper hanno inviato i dati, ordinamento e scrittura su file.")

		// Ordina i dati
		sort.Ints(w.data)

		// Scrivi i dati ordinati su file
		filepath := fmt.Sprintf("./Output/Output%d.txt", data.TargetRangeID)
		err := AppendToFile(filepath, w.data, w.id)
		if err != nil {
			return err
		}
	}

	w.mapperCount++

	*ack = true // Conferma la ricezione
	return nil
}

func (w *Worker) ProcessMapper(request *utils.WorkerRequest, response *utils.WorkerResponse) error {
	// Estrai i dati e i range dal master
	chunk := request.Chunk
	intervals := request.Intervals
	w.id = request.Id
	reducerPorts := request.ReducerPorts

	// Mappiamo i dati ai ranges specifici
	mappedData := mapDataToRanges(chunk, intervals)

	fmt.Println("Stampa dati mappati ", mappedData)

	// Itera su tutti i worker tranne se stesso
	for reducerID := 0; reducerID < len(reducerPorts); reducerID++ {

		sort.Ints(mappedData[reducerID])
		fmt.Println("L'output intermedio ordinato è", mappedData[reducerID])

		// Simula il lavoro del worker
		go func(reducerID int) {

			// Crea i dati da inviare
			dataToPass := &utils.WorkerData{
				SourceWorkerID: w.id,
				TargetRangeID:  reducerID,
				Values:         mappedData[reducerID],
			}

			// Connetti al worker sulla porta corrispondente
			client, err := rpc.Dial("tcp", "localhost:"+reducerPorts[reducerID])
			if err != nil {
				log.Printf("Errore nella connessione al reducer %d (porta %s): %v", reducerID+1, reducerPorts[reducerID], err)
				return // Se c'è un errore di connessione, la goroutine termina senza chiamare Done
			}
			defer client.Close()

			var ack bool = false
			// Invia i dati al worker
			err = client.Call("Worker.ReceiveData", dataToPass, &ack)
			if err != nil {
				log.Printf("Errore durante la chiamata RPC al worker %d: %v", reducerID+1, err)
			}

		}(reducerID)

		fmt.Println("Operazioni di Mapping e Shuffling terminate")

	}
	return nil
}

// Funzione che scrive un vettore di interi in un file in modalità append
func AppendToFile(filepath string, data []int, id int) error {

	// Apri il file in modalità append, se il file non esiste verrà creato e se esiste verrà pulito
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("errore nell'aprire il file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// Converti il vettore di interi in una stringa
	var dataStr []string
	for _, num := range data {
		dataStr = append(dataStr, strconv.Itoa(num))
	}
	// Unisci gli interi in una stringa separata da spazi
	line := strings.Join(dataStr, " ") + "\n"

	// Scrivi la stringa nel file
	_, err = file.WriteString(line)
	if err != nil {
		return fmt.Errorf("errore durante la scrittura nel file: %v", err)
	}

	return nil
}

func main() {
	// Ottieni la porta dalla riga di comando
	if len(os.Args) != 2 {
		log.Fatal("Errore: specificare la porta come argomento. Comando per lanciare il worker sulla porta xxxx: go run worker.go 1235")
	}
	port := os.Args[1]

	// Crea un'istanza del worker
	worker := new(Worker)
	worker.mapperCount = 0

	// Registra il tipo di server RPC per il worker
	err := rpc.Register(worker)
	if err != nil {
		log.Fatal("Errore nel registro RPC:", err)
	}

	// Ascolta sulla porta specificata
	listener, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		log.Fatal("Errore nell'ascolto sulla porta "+port+":", err)
	}

	fmt.Println("Worker in ascolto sulla porta " + port + "...")

	// Accetta connessioni e gestiscile in parallelo
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Errore durante l'accettazione della connessione:", err)
		}

		// Gestisce la connessione RPC in parallelo
		go rpc.ServeConn(conn)
	}
}
