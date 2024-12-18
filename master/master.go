package main

import (
	"MapReduce/utils"
	"fmt"
	"log"
	"math"
	"net"
	"net/rpc"
	"sort"
	"sync"
)

type Master struct {
	numWorkers  int
	numReducers int
}

func (m *Master) SampleData(data []int) []int {
	// Definizione dei bounds
	lowerBound := 0
	upperBound := math.MaxInt16

	// Campionamento per stimare la distribuzione
	sampleSize := len(data) / 10
	if sampleSize == 0 {
		sampleSize = 1
	}
	// Creiamo un campione e aggiungiamo lowerBound e upperBound
	sample := append(append([]int{}, data[:sampleSize]...), lowerBound, upperBound)
	sort.Ints(sample) // Ordiniamo il campione

	// Calcoliamo gli intervalli
	intervalCount := m.numReducers // Numero di intervalli
	intervals := make([]int, intervalCount+1)
	chunkSize := len(sample) / intervalCount

	// Definizione degli intervalli
	intervals[0] = lowerBound
	for i := 1; i < intervalCount; i++ {
		intervals[i] = sample[i*chunkSize]
	}
	intervals[intervalCount] = upperBound + 1

	// Suddivisione dei dati in partizioni
	partitions := make([][]int, intervalCount)
	for _, value := range data {
		for i := 0; i < intervalCount; i++ {
			if value >= intervals[i] && value < intervals[i+1] {
				partitions[i] = append(partitions[i], value)
				break
			}
		}
	}

	// Log per vedere gli intervalli
	fmt.Println("Intervalli calcolati:", intervals)
	// Impostiamo gli intervalli nella richiesta
	return intervals
}

func (m *Master) partitionData(data []int, intervals []int, numberOfWorkers int) [][]int {
	// Controllo degli intervalli
	if len(intervals) < 2 {
		log.Fatal("Errore: Gli intervalli devono essere almeno due per suddividere i dati.")
	}

	// Creiamo una slice di slice per contenere i chunks
	chunks := make([][]int, numberOfWorkers)

	// Partizioniamo i dati in base agli intervalli
	for i, value := range data {
		chunks[i%numberOfWorkers] = append(chunks[i%numberOfWorkers], value)
	}

	// Verifica che tutti i chunks siano corretti e non vuoti
	for i, chunk := range chunks {
		if len(chunk) == 0 {
			log.Printf("Attenzione: il chunk %d è vuoto.", i+1)
		}
	}
	return chunks
}

func (m *Master) DistributeData(data []int, intervals []int) error {
	// Log degli intervalli ricevuti
	fmt.Println("Intervalli ricevuti in DistributeData:", intervals)
	if len(intervals) < 2 {
		return fmt.Errorf("Errore: Gli intervalli devono essere almeno due per suddividere i dati.")
	}

	// Suddividi i dati nei chunks per i mapper
	chunks := m.partitionData(data, intervals, m.numWorkers)

	// Ottieni i worker disponibili
	workerPorts := utils.GlobalMapperPorts.GetPorts()
	reducerPorts := utils.GlobalReducerPorts.GetReducerPorts()

	// Controlla che ci siano abbastanza worker disponibili per mapper e reducer
	if len(workerPorts)+len(reducerPorts) < m.numWorkers+m.numReducers {
		return fmt.Errorf("Errore: non ci sono abbastanza worker disponibili (necessari %d, trovati %d)", m.numWorkers+m.numReducers, len(workerPorts))
	}

	// Crea un WaitGroup per sincronizzare l'esecuzione delle goroutine
	var wg sync.WaitGroup

	// Assegna i mapper ai primi worker disponibili
	for i, chunk := range chunks {
		workerRequest := &utils.WorkerRequest{
			Chunk:        chunk,
			Intervals:    intervals,
			Id:           i,
			MapperPorts:  workerPorts,
			ReducerPorts: reducerPorts,
			WorkerRole:   "Mapper",
		}
		workerResponse := &utils.WorkerResponse{}

		// Incrementa il WaitGroup per ogni mapper
		wg.Add(1)

		// Lancia la goroutine per ogni mapper
		go func(i int, workerRequest *utils.WorkerRequest) {
			defer wg.Done() // Segnala il completamento del lavoro di questa goroutine

			fmt.Printf("Chiamo il Mapper %s\n", workerPorts[i])
			client, err := rpc.Dial("tcp", "localhost:"+workerPorts[i])
			if err != nil {
				log.Printf("Errore nella connessione al Mapper %d (porta %s): %v", i+1, workerPorts[i], err)
				return // Se c'è un errore nella connessione, esci dalla goroutine
			}
			defer client.Close()

			// Invia il chunk al mapper
			err = client.Call("Worker.ProcessChunk", workerRequest, workerResponse)
			if err != nil {
				log.Printf("Errore durante la chiamata RPC al Mapper %d: %v", i+1, err)
			}

			fmt.Printf("Chiamo il Reducer %s\n", reducerPorts[i])
			client, err = rpc.Dial("tcp", "localhost:"+reducerPorts[i])
			if err != nil {
				log.Printf("Errore nella connessione al Mapper %d (porta %s): %v", i+1, reducerPorts[i], err)
				return // Se c'è un errore nella connessione, esci dalla goroutine
			}
			defer client.Close()

			// Invia il chunk al mapper
			err = client.Call("Worker.ProcessChunk", workerRequest, workerResponse)
			if err != nil {
				log.Printf("Errore durante la chiamata RPC al Reducer %d: %v", i+1, err)
			}

		}(i, workerRequest)
	}

	// Aspetta che tutti i mapper abbiano completato il loro lavoro
	wg.Wait()

	fmt.Println("Task Completato")

	return nil
}

// Funzione principale per gestire l'intero flusso
func (m *Master) HandleRequest(request *utils.SortRequest, response *utils.ClientResponse) error {
	// Passo 1: Calcola gli intervalli
	fmt.Println("Calcolo degli intervalli...")
	intervals := m.SampleData(request.Data)

	// Passo 2: Assegna i chunks ai worker
	fmt.Println("Assegnazione dei chunks ai worker...")
	err := m.DistributeData(request.Data, intervals)
	if err != nil {
		return fmt.Errorf("Errore in DistributeData: %v", err)
	}

	return nil
}

// Funzione principale che gestisce il server RPC
func main() {
	// Crea un'istanza del tipo Master
	master := new(Master)

	utils.GlobalMapperPorts.AddWorker("1230")
	utils.GlobalMapperPorts.AddWorker("1231")
	utils.GlobalMapperPorts.AddWorker("1232")
	utils.GlobalMapperPorts.AddWorker("1233")

	utils.GlobalReducerPorts.AddReducer("1235")
	utils.GlobalReducerPorts.AddReducer("1236")
	utils.GlobalReducerPorts.AddReducer("1237")
	utils.GlobalReducerPorts.AddReducer("1239")

	master.numWorkers = 4
	master.numReducers = 4

	// Registra il tipo di server RPC
	err := rpc.RegisterName("Master", master)

	if err != nil {
		log.Fatal("Errore nel registro RPC:", err)
	}

	// Ascolta sulla porta 1234
	listener, err := net.Listen("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Errore nell'ascolto sulla porta 1234:", err)
	}

	fmt.Println("Server RPC in ascolto sulla porta 1234...")

	// Accetta connessioni e gestiscile in parallelo
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Errore durante l'accettazione della connessione:", err)
		}

		fmt.Println(conn.RemoteAddr().String())
		// Gestisci la connessione RPC in parallelo
		go rpc.ServeConn(conn)
	}
}
